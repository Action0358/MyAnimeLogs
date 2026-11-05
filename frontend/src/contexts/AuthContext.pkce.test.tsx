/**
 * 実物の Supabase クライアントと AuthProvider を組み合わせて、PKCE のコード交換を検証する。
 *
 * Supabase は交換中のロックを持ったまま SIGNED_IN の購読者を待つ。購読者が同期の完了まで
 * 待つと、プロフィール取得に失敗したときの signOut とロックを待ち合って止まり、
 * ログイン画面が「認証中...」のまま固まる。モックのクライアントでは再現しない。
 */
import React from 'react'
import { render, waitFor } from '@testing-library/react'
import { useRouter } from 'next/navigation'

jest.mock('next/navigation')
jest.mock('@/lib/api/client', () => ({
  ApiClientError: class extends Error {},
  setUnauthorizedHandler: jest.fn(),
  setRefreshTokenHandler: jest.fn(),
}))

process.env.NEXT_PUBLIC_SUPABASE_URL = 'https://example.supabase.co'
process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY = 'anon'
window.history.pushState({}, '', '/auth/callback/?code=abc')

// 上の環境変数と URL を読ませるため、設定後に読み込む
const { supabase } = jest.requireActual('@/lib/supabase')
const { AuthProvider } = jest.requireActual('./AuthContext')

const jsonResponse = (status: number, body: unknown) =>
  ({
    ok: status >= 200 && status < 300,
    status,
    headers: new Headers({ 'content-type': 'application/json' }),
    json: async () => body,
    text: async () => JSON.stringify(body),
    clone() {
      return this
    },
  }) as unknown as Response

const tokenResponse = {
  access_token: 'access',
  refresh_token: 'refresh',
  token_type: 'bearer',
  expires_in: 3600,
  user: { id: 'user-1', aud: 'authenticated', app_metadata: {}, user_metadata: {}, created_at: '' },
}

describe('PKCE のコード交換と AuthProvider', () => {
  it('プロフィールが 401 でも固まらず、セッションを破棄して理由付きで /signin へ送る', async () => {
    const push = jest.fn()
    ;(useRouter as jest.Mock).mockReturnValue({ push, replace: jest.fn() })
    localStorage.setItem('sb-example-auth-token-code-verifier', JSON.stringify('verifier'))
    global.fetch = jest.fn(async (url: RequestInfo | URL) => {
      const target = String(url)
      if (target.includes('/token?grant_type=pkce')) return jsonResponse(200, tokenResponse)
      if (target.includes('/api/users/me')) return jsonResponse(401, {})
      return jsonResponse(200, {})
    }) as jest.Mock

    render(
      <AuthProvider>
        <div />
      </AuthProvider>
    )

    const { data, error } = await supabase.auth.exchangeCodeForSession('abc')

    expect(error).toBeNull()
    expect(data.session?.access_token).toBe('access')
    await waitFor(() => expect(push).toHaveBeenCalledWith('/signin?reason=unauthorized'))
    await waitFor(async () => {
      const { data: after } = await supabase.auth.getSession()
      expect(after.session).toBeNull()
    })
    expect(localStorage.getItem('auth_token')).toBeNull()
  })
})
