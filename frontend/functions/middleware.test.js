/**
 * @jest-environment node
 */
import {
  BACKEND_TIMEOUT_MS,
  STATUS_CACHE_TTL_MS,
  isPageRequest,
  onRequest,
  resetStatusCache,
  resolveMaintenanceStatus,
  retryAfterSeconds,
} from './_middleware.js'

/**
 * メンテナンス中にページの HTML をメンテナンス画面へ差し替えるミドルウェア（#1136）。
 * Cloudflare の実行環境（ASSETS・HTMLRewriter）は最小限の代役で置き換える。
 */

const ORIGIN = 'https://myanimelogs.example'
const API_URL = 'https://api.example'

/** HTMLRewriter の代役。登録されたセレクタと処理を記録し、本文はそのまま返す */
class FakeHTMLRewriter {
  static instances = []
  constructor() {
    this.handlers = []
    FakeHTMLRewriter.instances.push(this)
  }
  on(selector, handler) {
    this.handlers.push({ selector, handler })
    return this
  }
  transform(response) {
    return response
  }
  get selectors() {
    return this.handlers.map((h) => h.selector)
  }
}

const json = (data, init) =>
  new Response(JSON.stringify(data), { headers: { 'Content-Type': 'application/json' }, ...init })

/** 静的ファイル。path → Response を返す関数 */
const makeAssets = (files) => ({
  fetch: jest.fn(async (input) => {
    const { pathname } = new URL(input instanceof Request ? input.url : input)
    const make = files[pathname]
    return make ? make() : new Response('not found', { status: 404 })
  }),
})

const defaultFiles = ({ staticEnabled = false, endsAt = null, edgeConfig = true } = {}) => ({
  '/maintenance.json': () => json({ enabled: staticEnabled, endsAt }),
  ...(edgeConfig ? { '/_edge-config.json': () => json({ apiUrl: API_URL }) } : {}),
  '/maintenance/': () =>
    new Response('<html><head></head><body>maintenance</body></html>', {
      headers: { 'Content-Type': 'text/html', 'Content-Security-Policy': "default-src 'self'" },
    }),
})

const pageRequest = (path, accept = 'text/html,application/xhtml+xml') =>
  new Request(`${ORIGIN}${path}`, { headers: { Accept: accept } })

const makeContext = (request, files = defaultFiles()) => {
  const nextResponse = new Response('<html><head></head><body>normal</body></html>', {
    headers: { 'Content-Type': 'text/html' },
  })
  return {
    request,
    env: { ASSETS: makeAssets(files) },
    next: jest.fn(async () => nextResponse),
    nextResponse,
  }
}

/** バックエンドの GET /api/maintenance の応答を決める */
const mockBackend = (impl) => jest.spyOn(global, 'fetch').mockImplementation(impl)

beforeEach(() => {
  resetStatusCache()
  FakeHTMLRewriter.instances = []
  global.HTMLRewriter = FakeHTMLRewriter
})

afterEach(() => {
  jest.restoreAllMocks()
  delete global.HTMLRewriter
})

describe('isPageRequest', () => {
  it('HTML を求める GET / HEAD だけを対象にする', () => {
    expect(isPageRequest(pageRequest('/'))).toBe(true)
    expect(isPageRequest(pageRequest('/mypage/'))).toBe(true)
    expect(isPageRequest(new Request(`${ORIGIN}/`, { method: 'HEAD', headers: { Accept: 'text/html' } }))).toBe(true)
    expect(isPageRequest(new Request(`${ORIGIN}/`, { method: 'POST', headers: { Accept: 'text/html' } }))).toBe(false)
    expect(isPageRequest(pageRequest('/', 'application/json'))).toBe(false)
  })

  it('静的ファイル・Next.js のチャンク・RSC の取得は対象外', () => {
    expect(isPageRequest(pageRequest('/_next/static/chunks/main.js'))).toBe(false)
    expect(isPageRequest(pageRequest('/news/index.txt'))).toBe(false)
    expect(isPageRequest(pageRequest('/logo.svg'))).toBe(false)
  })
})

describe('resolveMaintenanceStatus', () => {
  it('maintenance.json が有効ならバックエンドに問い合わせずメンテナンス', async () => {
    const fetchSpy = mockBackend(async () => json({ enabled: false }))
    const ctx = makeContext(pageRequest('/'), defaultFiles({ staticEnabled: true, endsAt: '2026-10-25T06:00:00Z' }))

    await expect(resolveMaintenanceStatus(ctx)).resolves.toEqual({
      enabled: true,
      endsAt: '2026-10-25T06:00:00Z',
    })
    expect(fetchSpy).not.toHaveBeenCalled()
  })

  it('バックエンドが enabled=true ならメンテナンス（終了予定つき）', async () => {
    const fetchSpy = mockBackend(async () => json({ enabled: true, ends_at: '2026-10-25T06:00:00Z' }))

    await expect(resolveMaintenanceStatus(makeContext(pageRequest('/')))).resolves.toEqual({
      enabled: true,
      endsAt: '2026-10-25T06:00:00Z',
    })
    expect(fetchSpy).toHaveBeenCalledWith(`${API_URL}/api/maintenance`, expect.anything())
  })

  it('バックエンドが落ちている・応答が不正なら通常時として扱う', async () => {
    mockBackend(async () => {
      throw new TypeError('fetch failed')
    })
    await expect(resolveMaintenanceStatus(makeContext(pageRequest('/')))).resolves.toEqual({ enabled: false })

    resetStatusCache()
    jest.restoreAllMocks()
    mockBackend(async () => json({ unexpected: true }))
    await expect(resolveMaintenanceStatus(makeContext(pageRequest('/')))).resolves.toEqual({ enabled: false })
  })

  it('_edge-config.json がなければバックエンドには問い合わせない', async () => {
    const fetchSpy = mockBackend(async () => json({ enabled: true }))
    const ctx = makeContext(pageRequest('/'), defaultFiles({ edgeConfig: false }))

    await expect(resolveMaintenanceStatus(ctx)).resolves.toEqual({ enabled: false })
    expect(fetchSpy).not.toHaveBeenCalled()
  })

  it('応答しないバックエンドはタイムアウトで打ち切る', async () => {
    jest.useFakeTimers()
    mockBackend(
      (_url, init) =>
        new Promise((_resolve, reject) => {
          init.signal.addEventListener('abort', () => reject(new Error('aborted')))
        })
    )
    const pending = resolveMaintenanceStatus(makeContext(pageRequest('/')))
    await jest.advanceTimersByTimeAsync(BACKEND_TIMEOUT_MS)

    await expect(pending).resolves.toEqual({ enabled: false })
    jest.useRealTimers()
  })

  it('判定は TTL の間だけ使い回す', async () => {
    const fetchSpy = mockBackend(async () => json({ enabled: true }))
    const now = 1_000_000

    await resolveMaintenanceStatus(makeContext(pageRequest('/')), now)
    await resolveMaintenanceStatus(makeContext(pageRequest('/')), now + STATUS_CACHE_TTL_MS - 1)
    expect(fetchSpy).toHaveBeenCalledTimes(1)

    await resolveMaintenanceStatus(makeContext(pageRequest('/')), now + STATUS_CACHE_TTL_MS)
    expect(fetchSpy).toHaveBeenCalledTimes(2)
  })
})

describe('retryAfterSeconds', () => {
  const now = Date.parse('2026-10-25T05:50:00Z')
  it('終了予定までの秒数。なければ・過ぎていれば 300', () => {
    expect(retryAfterSeconds('2026-10-25T06:00:00Z', now)).toBe(600)
    expect(retryAfterSeconds(undefined, now)).toBe(300)
    expect(retryAfterSeconds('2026-10-25T05:00:00Z', now)).toBe(300)
    expect(retryAfterSeconds('broken', now)).toBe(300)
  })
})

describe('onRequest', () => {
  it('ページ以外はそのまま通す（判定もしない）', async () => {
    const fetchSpy = mockBackend(async () => json({ enabled: true }))
    const ctx = makeContext(pageRequest('/_next/static/chunks/main.js'))

    await expect(onRequest(ctx)).resolves.toBe(ctx.nextResponse)
    expect(fetchSpy).not.toHaveBeenCalled()
  })

  it('通常時はそのまま通す', async () => {
    mockBackend(async () => json({ enabled: false }))
    const ctx = makeContext(pageRequest('/mypage/'))

    await expect(onRequest(ctx)).resolves.toBe(ctx.nextResponse)
  })

  it('メンテナンス中は元の URL のまま、メンテナンス画面を 503 で返す', async () => {
    mockBackend(async () => json({ enabled: true, ends_at: '2099-01-01T00:00:00Z' }))
    const ctx = makeContext(pageRequest('/mypage/'))

    const res = await onRequest(ctx)

    expect(ctx.next).not.toHaveBeenCalled()
    expect(res.status).toBe(503)
    expect(res.headers.get('Cache-Control')).toBe('no-store')
    expect(Number(res.headers.get('Retry-After'))).toBeGreaterThan(300)
    // 静的ファイルのヘッダー（_headers の CSP など）は引き継ぐ
    expect(res.headers.get('Content-Security-Policy')).toBe("default-src 'self'")
    await expect(res.text()).resolves.toContain('maintenance')

    const [rewriter] = FakeHTMLRewriter.instances
    expect(rewriter.selectors).toEqual([
      'script',
      'link[as="script"]',
      '[data-maintenance-ends-at-row]',
      '[data-maintenance-ends-at]',
    ])
  })

  it('終了予定がなければ、終了予定の行には触らない', async () => {
    mockBackend(async () => json({ enabled: true }))

    const res = await onRequest(makeContext(pageRequest('/')))

    expect(res.status).toBe(503)
    expect(res.headers.get('Retry-After')).toBe('300')
    expect(FakeHTMLRewriter.instances[0].selectors).toEqual(['script', 'link[as="script"]'])
  })

  it('終了予定は「終了予定：…ごろ」の文言にして差し込む', async () => {
    mockBackend(async () => json({ enabled: true, ends_at: '2026-10-25T06:00:00Z' }))
    await onRequest(makeContext(pageRequest('/')))

    const { handler } = FakeHTMLRewriter.instances[0].handlers.find(
      (h) => h.selector === '[data-maintenance-ends-at]'
    )
    const el = { setInnerContent: jest.fn() }
    handler.element(el)
    expect(el.setInnerContent).toHaveBeenCalledWith('終了予定：2026年10月25日 15:00ごろ')
  })

  it('/news はメンテナンス中も開き、メンテナンス中であることを meta で伝える', async () => {
    mockBackend(async () => json({ enabled: true, ends_at: '2026-10-25T06:00:00Z' }))
    const ctx = makeContext(pageRequest('/news/'))

    const res = await onRequest(ctx)

    expect(ctx.next).toHaveBeenCalled()
    expect(res.status).toBe(200)
    const { selector, handler } = FakeHTMLRewriter.instances[0].handlers[0]
    expect(selector).toBe('head')
    const el = { append: jest.fn() }
    handler.element(el)
    expect(el.append).toHaveBeenCalledWith(
      '<meta name="myanimelogs-maintenance" content="2026-10-25T06:00:00.000Z">',
      { html: true }
    )
  })

  it('メンテナンス画面のファイルがなければ通常どおり返す（クライアントの判定に任せる）', async () => {
    mockBackend(async () => json({ enabled: true }))
    const files = defaultFiles()
    delete files['/maintenance/']
    const ctx = makeContext(pageRequest('/'), files)

    await expect(onRequest(ctx)).resolves.toBe(ctx.nextResponse)
  })
})
