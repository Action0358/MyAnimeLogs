/**
 * @jest-environment jsdom
 */
import { webcrypto } from 'crypto'
import { TextEncoder } from 'util'
import type { SupabaseClient } from '@supabase/supabase-js'

// jsdom には Web Crypto が無く、無いと PKCE が s256 ではなく plain に落ちる
Object.assign(globalThis, { TextEncoder })
Object.defineProperty(globalThis, 'crypto', { value: webcrypto })

const CODE_VERIFIER_KEY = 'sb-example-auth-token-code-verifier'

const tokenResponse = () => ({
  access_token: 'access',
  refresh_token: 'refresh',
  token_type: 'bearer',
  expires_in: 3600,
  user: { id: 'user-1', aud: 'authenticated', app_metadata: {}, user_metadata: {}, created_at: '' },
})

/** 実物のクライアントを、指定の URL を開いた直後の状態で読み込む */
const loadClient = (url: string): SupabaseClient => {
  window.history.pushState({}, '', url)
  process.env.NEXT_PUBLIC_SUPABASE_URL = 'https://example.supabase.co'
  process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY = 'anon'
  let client!: SupabaseClient
  jest.isolateModules(() => {
    client = require('./supabase').supabase
  })
  return client
}

describe('supabase client', () => {
  const originalFetch = global.fetch
  let fetchMock: jest.Mock

  const tokenRequests = () =>
    fetchMock.mock.calls.filter(([url]) => String(url).includes('/token?grant_type=pkce'))

  beforeEach(() => {
    localStorage.clear()
    fetchMock = jest.fn(
      async () =>
        ({
          ok: true,
          status: 200,
          headers: new Headers({ 'content-type': 'application/json' }),
          json: async () => tokenResponse(),
          text: async () => JSON.stringify(tokenResponse()),
        }) as unknown as Response
    )
    global.fetch = fetchMock
  })

  afterEach(() => {
    global.fetch = originalFetch
    window.history.pushState({}, '', '/')
  })

  it('ログイン開始時に PKCE の code_challenge を付け、code verifier を保存する', async () => {
    const client = loadClient('/signin/')

    const { data } = await client.auth.signInWithOAuth({
      provider: 'google',
      options: { skipBrowserRedirect: true },
    })

    const params = new URL(data.url as string).searchParams
    expect(params.get('code_challenge')).toBeTruthy()
    expect(params.get('code_challenge_method')).toBe('s256')
    expect(localStorage.getItem(CODE_VERIFIER_KEY)).toBeTruthy()
  })

  // 自動検出が有効だと、起動時の交換で code verifier が消えて callback の交換が失敗する
  it('?code= は起動時に交換せず、callback の交換が1回で成功する', async () => {
    localStorage.setItem(CODE_VERIFIER_KEY, JSON.stringify('verifier'))
    const client = loadClient('/auth/callback/?code=abc')

    const { data, error } = await client.auth.exchangeCodeForSession('abc')

    expect(error).toBeNull()
    expect(data.session?.access_token).toBe('access')
    expect(tokenRequests()).toHaveLength(1)
    expect(JSON.parse(tokenRequests()[0][1].body)).toMatchObject({
      auth_code: 'abc',
      code_verifier: 'verifier',
    })
  })

  it('URL のフラグメントのトークンを受け付けない', async () => {
    const client = loadClient(
      '/auth/callback/#access_token=stolen&refresh_token=r&expires_in=3600&token_type=bearer'
    )

    const { data } = await client.auth.getSession()

    expect(data.session).toBeNull()
  })
})
