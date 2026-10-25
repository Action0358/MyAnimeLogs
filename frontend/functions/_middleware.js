// Cloudflare Pages Functions のミドルウェア: メンテナンス中は、ページの HTML を返す前に
// メンテナンス画面へ差し替える（#1136）。
//
// 静的サイトは最初に通常ページの HTML を返すため、クライアントだけで判定すると
// 判定が付くまでの一瞬、通常のトップが見える。HTML を返す時点で差し替えて防ぐ。
//
// 判定元は既存の2つだけで、Cloudflare 側に設定は持たない。
//   1. /maintenance.json の enabled（静的フラグ。デプロイで切り替え）
//   2. バックエンド GET /api/maintenance（fly secrets の MAINTENANCE_MODE）
// バックエンドの URL はビルド時に書き出す /_edge-config.json から読む
// （scripts/write-edge-config.js。NEXT_PUBLIC_API_URL と同じ値）。

import {
  ENDS_AT_ROW_ATTR,
  ENDS_AT_TEXT_ATTR,
  MAINTENANCE_HTML_ATTR,
  MAINTENANCE_META_NAME,
  MAINTENANCE_PAGE_PATH,
  endsAtText,
  isMaintenanceExemptPath,
} from '../src/lib/maintenance/shared'

/** 判定結果をこの isolate 内で使い回す時間。Fly で切り替えたときの反映はこの分遅れる */
export const STATUS_CACHE_TTL_MS = 30_000
/** バックエンドが応答しないときに待つ上限。超えたら静的フラグだけで判定する */
export const BACKEND_TIMEOUT_MS = 1_500
/** 終了予定がないときの Retry-After（秒） */
const DEFAULT_RETRY_AFTER_SECONDS = 300

const NOT_IN_MAINTENANCE = { enabled: false }

/**
 * ページ（HTML）の表示要求か。静的ファイル・Next.js のチャンク・RSC の取得
 * （サイト内遷移。拡張子 .txt）は対象外にし、余計な判定をしない。
 */
export const isPageRequest = (request) => {
  if (request.method !== 'GET' && request.method !== 'HEAD') return false
  if (!(request.headers.get('Accept') || '').includes('text/html')) return false
  const { pathname } = new URL(request.url)
  if (pathname.startsWith('/_next/')) return false
  const lastSegment = pathname.split('/').pop() || ''
  return !lastSegment.includes('.')
}

/** 静的ファイル（ASSETS）から JSON を読む。取れなければ null */
const readAssetJson = async (env, requestUrl, path) => {
  try {
    const res = await env.ASSETS.fetch(new URL(path, requestUrl))
    if (!res.ok) return null
    return await res.json()
  } catch {
    return null
  }
}

const readStaticFlag = async (env, requestUrl) => {
  const data = await readAssetJson(env, requestUrl, '/maintenance.json')
  if (!data || data.enabled !== true) return null
  return { enabled: true, endsAt: typeof data.endsAt === 'string' ? data.endsAt : undefined }
}

/** バックエンドの状態。応答しない・形式が違うときは null（判定できない） */
const readBackendStatus = async (env, requestUrl) => {
  const config = await readAssetJson(env, requestUrl, '/_edge-config.json')
  const apiUrl = config && typeof config.apiUrl === 'string' ? config.apiUrl : ''
  if (!apiUrl) return null

  const controller = new AbortController()
  const timer = setTimeout(() => controller.abort(), BACKEND_TIMEOUT_MS)
  try {
    const res = await fetch(`${apiUrl.replace(/\/$/, '')}/api/maintenance`, {
      headers: { Accept: 'application/json' },
      signal: controller.signal,
    })
    if (!res.ok) return null
    const data = await res.json()
    if (!data || typeof data.enabled !== 'boolean') return null
    if (!data.enabled) return NOT_IN_MAINTENANCE
    return { enabled: true, endsAt: typeof data.ends_at === 'string' ? data.ends_at : undefined }
  } catch {
    return null
  } finally {
    clearTimeout(timer)
  }
}

let cache = null

/** テスト用: 判定のキャッシュを捨てる */
export const resetStatusCache = () => {
  cache = null
}

/**
 * メンテナンス中か。静的フラグを先に見て、なければバックエンドに問い合わせる。
 * どちらでも判定できなければ通常時として扱う（メンテナンス機能の不調でサイトを止めない）。
 * バックエンドが落ちているときも結果をキャッシュし、毎回タイムアウトを待たない。
 */
export const resolveMaintenanceStatus = async (context, now = Date.now()) => {
  if (cache && cache.expiresAt > now) return cache.status

  const { env, request } = context
  const status =
    (await readStaticFlag(env, request.url)) ??
    (await readBackendStatus(env, request.url)) ??
    NOT_IN_MAINTENANCE

  cache = { status, expiresAt: now + STATUS_CACHE_TTL_MS }
  return status
}

/** 終了予定までの秒数。なければ・過ぎていれば既定値 */
export const retryAfterSeconds = (endsAt, now = Date.now()) => {
  const end = endsAt ? new Date(endsAt).getTime() : NaN
  const remaining = Math.ceil((end - now) / 1000)
  return Number.isFinite(remaining) && remaining > 0 ? remaining : DEFAULT_RETRY_AFTER_SECONDS
}

const removeElement = { element: (el) => el.remove() }

/**
 * /maintenance の静的 HTML を、元の URL のまま 503 で返す。
 * スクリプトは取り除く。読み込むと Next.js のルーターが URL に合わせて通常ページを
 * 描画し直してしまうため（画面に操作がないので、スクリプトがなくても困らない）。
 */
export const renderMaintenancePage = async (context, status) => {
  const { env, request } = context
  const page = await env.ASSETS.fetch(new URL(`${MAINTENANCE_PAGE_PATH}/`, request.url))
  // 画面が見つからないときは通常どおり返す。クライアント側の判定が引き継ぐ
  if (!page.ok) return context.next()

  let rewriter = new HTMLRewriter()
    .on('script', removeElement)
    .on('link[as="script"]', removeElement)

  const text = endsAtText(status.endsAt)
  if (text) {
    rewriter = rewriter
      .on(`[${ENDS_AT_ROW_ATTR}]`, { element: (el) => el.removeAttribute('hidden') })
      .on(`[${ENDS_AT_TEXT_ATTR}]`, { element: (el) => el.setInnerContent(text) })
  }

  const headers = new Headers(page.headers)
  // 解除後に古い画面が残らないよう、ブラウザ・CDN にキャッシュさせない
  headers.set('Cache-Control', 'no-store')
  headers.set('Retry-After', String(retryAfterSeconds(status.endsAt)))
  return new Response(rewriter.transform(page).body, { status: 503, headers })
}

/**
 * メンテナンス中に開放しているページ（/news）に、メンテナンス中であることを伝える目印を足す。
 *   - <html data-maintenance>: JS が動く前から CSS でヘッダー・フッターをメンテナンス用の
 *     見た目にする（リロード時に通常のナビ・ログインが一瞬見えないように）
 *   - meta: クライアントがこれを読み、メンテナンス状態として描画し直す
 */
export const addMaintenanceMarker = (response, status) => {
  // HTML に埋め込むため、日時として解釈し直した値だけを使う（入力をそのまま入れない）
  const time = status.endsAt ? new Date(status.endsAt).getTime() : NaN
  const endsAt = Number.isNaN(time) ? '' : new Date(time).toISOString()
  return new HTMLRewriter()
    .on('html', { element: (el) => el.setAttribute(MAINTENANCE_HTML_ATTR, '') })
    .on('head', {
      element: (el) =>
        el.append(`<meta name="${MAINTENANCE_META_NAME}" content="${endsAt}">`, { html: true }),
    })
    .transform(response)
}

export async function onRequest(context) {
  if (!isPageRequest(context.request)) return context.next()

  const status = await resolveMaintenanceStatus(context)
  if (!status.enabled) return context.next()

  const { pathname } = new URL(context.request.url)
  if (isMaintenanceExemptPath(pathname)) {
    return addMaintenanceMarker(await context.next(), status)
  }
  return renderMaintenancePage(context, status)
}
