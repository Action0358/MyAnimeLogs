import React from 'react'
import { render, screen, waitFor, act } from '@testing-library/react'
import '@testing-library/jest-dom'
import { AuthProvider, useAuth } from './AuthContext'
import { supabase } from '@/lib/supabase'
import { useRouter } from 'next/navigation'
import {
  getMaintenanceState,
  resetMaintenance,
  setMaintenance,
} from '@/lib/maintenance/maintenanceStore'

jest.mock('next/navigation')
jest.mock('@/lib/supabase', () => ({
  supabase: {
    auth: {
      getSession: jest.fn(),
      onAuthStateChange: jest.fn(),
      signOut: jest.fn(),
      signInWithOAuth: jest.fn(),
    },
  },
  initialUrlHadCode: { value: false },
}))
jest.mock('@/lib/api/client', () => ({
  ApiClientError: class extends Error {},
  setUnauthorizedHandler: jest.fn(),
  setRefreshTokenHandler: jest.fn(),
}))

const mockUseRouter = useRouter as jest.Mock
const mockSupabase = supabase as unknown as {
  auth: {
    getSession: jest.Mock
    onAuthStateChange: jest.Mock
    signOut: jest.Mock
  }
}

const ACCESS_TOKEN = 'access-token'

const AuthProbe: React.FC = () => {
  const { isAuthenticated, isLoading } = useAuth()
  if (isLoading) return <div>loading</div>
  return <div>{isAuthenticated ? 'authenticated' : 'anonymous'}</div>
}

/**
 * プロフィール取得に失敗したときの扱いを検証する（#797）。
 *
 * 以前は失敗を一律 null に潰していたため、トークンだけあってユーザーが無い
 * 状態が残り、画面上は「ログインしても何も起きない」としか見えなかった。
 * 失敗が理由付きで /signin に現れることを保証する。
 */
describe('AuthProvider のプロフィール取得', () => {
  const mockPush = jest.fn()

  const renderWithSession = async () => {
    await act(async () => {
      render(
        <AuthProvider>
          <AuthProbe />
        </AuthProvider>
      )
    })
  }

  beforeEach(() => {
    jest.clearAllMocks()
    mockUseRouter.mockReturnValue({ push: mockPush, replace: jest.fn() })
    mockSupabase.auth.getSession.mockResolvedValue({
      data: { session: { access_token: ACCESS_TOKEN } },
    })
    mockSupabase.auth.onAuthStateChange.mockReturnValue({
      data: { subscription: { unsubscribe: jest.fn() } },
    })
    mockSupabase.auth.signOut.mockResolvedValue({ error: null })
    localStorage.clear()
  })

  it('取得できたらログイン状態になる', async () => {
    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ id: 1, name: 'テストユーザー', email: 'test@example.com' }),
    })

    await renderWithSession()

    expect(await screen.findByText('authenticated')).toBeInTheDocument()
    expect(mockPush).not.toHaveBeenCalled()
  })

  it('500が続く場合はサインアウトして理由付きで/signinへ送る', async () => {
    global.fetch = jest.fn().mockResolvedValue({ ok: false, status: 500 })

    await renderWithSession()

    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith('/signin?reason=profile-unavailable')
    })
    // 中途半端な認証状態を残さない
    expect(mockSupabase.auth.signOut).toHaveBeenCalled()
    expect(localStorage.getItem('auth_token')).toBeNull()
  })

  it('一時的な500は1度だけ再試行し、成功すればログインできる', async () => {
    global.fetch = jest
      .fn()
      .mockResolvedValueOnce({ ok: false, status: 500 })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ id: 1, name: 'テストユーザー', email: 'test@example.com' }),
      })

    await renderWithSession()

    expect(await screen.findByText('authenticated')).toBeInTheDocument()
    expect(global.fetch).toHaveBeenCalledTimes(2)
    expect(mockPush).not.toHaveBeenCalled()
  })

  it('401は再試行せず未認証として扱う', async () => {
    global.fetch = jest.fn().mockResolvedValue({ ok: false, status: 401 })

    await renderWithSession()

    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith('/signin?reason=unauthorized')
    })
    // 認証が無効なら再試行しても結果は変わらない
    expect(global.fetch).toHaveBeenCalledTimes(1)
  })

  it('通信障害でも黙って終わらせない', async () => {
    global.fetch = jest.fn().mockRejectedValue(new TypeError('Failed to fetch'))

    await renderWithSession()

    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith('/signin?reason=profile-unavailable')
    })
  })

  describe('メンテナンス中', () => {
    afterEach(() => resetMaintenance())

    it('503 + MAINTENANCE ならセッションを残し、/signin へ送らない', async () => {
      const body = { code: 'MAINTENANCE', ends_at: '2026-10-25T06:00:00Z' }
      const res = { ok: false, status: 503, json: async () => body }
      global.fetch = jest.fn().mockResolvedValue({ ...res, clone: () => res })

      await renderWithSession()

      await waitFor(() => expect(getMaintenanceState().active).toBe(true))
      expect(screen.getByText('anonymous')).toBeInTheDocument()
      // /news を見ている人をログアウトさせ、メンテナンス画面へ戻さない
      expect(mockSupabase.auth.signOut).not.toHaveBeenCalled()
      expect(mockPush).not.toHaveBeenCalled()
      expect(localStorage.getItem('auth_token')).toBe(JSON.stringify(ACCESS_TOKEN))
      // 再試行しても結果は変わらない
      expect(global.fetch).toHaveBeenCalledTimes(1)
    })

    it('maintenance.json で宣言済みなら、通信障害でもセッションを残す', async () => {
      act(() => setMaintenance({}))
      global.fetch = jest.fn().mockRejectedValue(new TypeError('Failed to fetch'))

      await renderWithSession()

      await waitFor(() => expect(global.fetch).toHaveBeenCalledTimes(2))
      await act(async () => {})
      expect(mockSupabase.auth.signOut).not.toHaveBeenCalled()
      expect(mockPush).not.toHaveBeenCalled()
    })

    it('メンテナンス以外の 503 は従来どおり取得失敗として扱う', async () => {
      const body = { code: 'SERVICE_UNAVAILABLE' }
      const res = { ok: false, status: 503, json: async () => body }
      global.fetch = jest.fn().mockResolvedValue({ ...res, clone: () => res })

      await renderWithSession()

      await waitFor(() => {
        expect(mockPush).toHaveBeenCalledWith('/signin?reason=profile-unavailable')
      })
      expect(getMaintenanceState().active).toBe(false)
    })
  })

  it('200でも形が違えば取得失敗として扱う', async () => {
    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ unexpected: 'shape' }),
    })

    await renderWithSession()

    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith('/signin?reason=profile-unavailable')
    })
  })
})
