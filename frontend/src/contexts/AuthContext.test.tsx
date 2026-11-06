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

/**
 * 自分で操作したログアウトは見ていたページにとどまる（#1173）。完了はヘッダーの
 * ログインボタンへの切り替わりで伝わるため、メッセージは出さない。
 * タイムアウトとアカウント削除は、これまでどおり理由に応じたページへ移動する。
 */
describe('AuthProvider のログアウト', () => {
  const mockPush = jest.fn()

  const LogoutProbe: React.FC = () => {
    const { logout, isAuthenticated, isManualLogout, isAccountDeleted } = useAuth()
    return (
      <div>
        <div>{isAuthenticated ? 'authenticated' : 'anonymous'}</div>
        <div>{isManualLogout ? 'manual-logout' : 'other-logout'}</div>
        <div>{isAccountDeleted ? 'account-deleted' : 'not-deleted'}</div>
        <button onClick={() => logout()}>manual</button>
        <button onClick={() => logout('timeout')}>timeout</button>
        <button onClick={() => logout('account-deleted')}>deleted</button>
      </div>
    )
  }

  const renderLoggedIn = async () => {
    await act(async () => {
      render(
        <AuthProvider>
          <LogoutProbe />
        </AuthProvider>
      )
    })
    expect(await screen.findByText('authenticated')).toBeInTheDocument()
  }

  const clickLogout = async (name: string) => {
    await act(async () => {
      screen.getByRole('button', { name }).click()
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
    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ id: 1, name: 'テストユーザー', email: 'test@example.com' }),
    })
    localStorage.clear()
  })

  it('自分で操作すると移動しない', async () => {
    await renderLoggedIn()

    await clickLogout('manual')

    expect(mockSupabase.auth.signOut).toHaveBeenCalled()
    expect(screen.getByText('anonymous')).toBeInTheDocument()
    expect(screen.getByText('manual-logout')).toBeInTheDocument()
    expect(mockPush).not.toHaveBeenCalled()
  })

  // #1265: ProtectedRoute が、削除後の移動先（/account-deleted）を邪魔しないための印
  it('アカウント削除だけが isAccountDeleted を立てる', async () => {
    await renderLoggedIn()
    expect(screen.getByText('not-deleted')).toBeInTheDocument()

    await clickLogout('deleted')

    expect(screen.getByText('account-deleted')).toBeInTheDocument()
  })

  it('手動ログアウトでは isAccountDeleted を立てない', async () => {
    await renderLoggedIn()

    await clickLogout('manual')

    expect(screen.getByText('not-deleted')).toBeInTheDocument()
  })

  it.each([
    ['timeout', '/signin?reason=timeout'],
    ['deleted', '/account-deleted'],
  ])('%s では理由に応じたページへ移動する', async (name, path) => {
    await renderLoggedIn()

    await clickLogout(name)

    expect(mockPush).toHaveBeenCalledWith(path)
    expect(screen.getByText('other-logout')).toBeInTheDocument()
  })

  /**
   * タブへの復帰などで始まったプロフィール取得が、ログアウト後に完了する競合（#1191）。
   * 古い結果を反映すると、ユーザーが状態に戻り、手動ログアウトの印も消えて
   * ProtectedRoute がトップではなく /signin へ送ってしまう。
   */
  it('ログアウト前に始まったプロフィール取得の結果は反映しない', async () => {
    let emit: (event: string, session: unknown) => Promise<void> = async () => {}
    mockSupabase.auth.onAuthStateChange.mockImplementation((callback) => {
      emit = callback
      return { data: { subscription: { unsubscribe: jest.fn() } } }
    })
    const UserProbe: React.FC = () => {
      const { user } = useAuth()
      return <div>{user ? 'has-user' : 'no-user'}</div>
    }
    await act(async () => {
      render(
        <AuthProvider>
          <LogoutProbe />
          <UserProbe />
        </AuthProvider>
      )
    })
    expect(await screen.findByText('authenticated')).toBeInTheDocument()

    let finishProfile: (response: unknown) => void = () => {}
    ;(global.fetch as jest.Mock).mockImplementationOnce(
      () => new Promise((resolve) => (finishProfile = resolve))
    )
    let pendingSync: Promise<void> = Promise.resolve()
    await act(async () => {
      // 同じトークンの再通知は取得しないため、取得が走る更新後のトークンで競合を再現する
      pendingSync = emit('TOKEN_REFRESHED', { access_token: 'refreshed-token' })
    })
    mockSupabase.auth.signOut.mockImplementation(async () => {
      await emit('SIGNED_OUT', null)
      return { error: null }
    })

    await clickLogout('manual')
    await act(async () => {
      finishProfile({
        ok: true,
        status: 200,
        json: async () => ({ id: 1, name: 'テストユーザー', email: 'test@example.com' }),
      })
      await pendingSync
    })

    expect(screen.getByText('no-user')).toBeInTheDocument()
    expect(screen.getByText('manual-logout')).toBeInTheDocument()
  })
})

/**
 * 同じトークンのプロフィール取得をまとめる（#1236）。
 * Supabase は起動時の INITIAL_SESSION やタブ復帰時の SIGNED_IN で同じトークンを何度も
 * 通知する。そのたびに取得すると、1回のページ表示で /api/users/me が何度も呼ばれる。
 */
describe('AuthProvider のプロフィール取得の回数', () => {
  const profile = { id: 1, name: 'テストユーザー', email: 'test@example.com' }
  const okResponse = { ok: true, status: 200, json: async () => profile }
  let emit: (event: string, session: unknown) => void = () => {}

  const session = (token: string) => ({ access_token: token })

  const renderProvider = async () => {
    await act(async () => {
      render(
        <AuthProvider>
          <AuthProbe />
        </AuthProvider>
      )
    })
  }

  const emitAndSettle = async (event: string, value: unknown) => {
    await act(async () => {
      emit(event, value)
      await new Promise((resolve) => setTimeout(resolve, 0))
    })
  }

  beforeEach(() => {
    jest.clearAllMocks()
    mockUseRouter.mockReturnValue({ push: jest.fn(), replace: jest.fn() })
    mockSupabase.auth.getSession.mockResolvedValue({ data: { session: session(ACCESS_TOKEN) } })
    mockSupabase.auth.onAuthStateChange.mockImplementation((callback) => {
      emit = callback
      return { data: { subscription: { unsubscribe: jest.fn() } } }
    })
    mockSupabase.auth.signOut.mockResolvedValue({ error: null })
    global.fetch = jest.fn().mockResolvedValue(okResponse)
    localStorage.clear()
  })

  it('起動時に getSession と INITIAL_SESSION が重なっても1回だけ取得する', async () => {
    mockSupabase.auth.onAuthStateChange.mockImplementation((callback) => {
      emit = callback
      callback('INITIAL_SESSION', session(ACCESS_TOKEN))
      return { data: { subscription: { unsubscribe: jest.fn() } } }
    })

    await renderProvider()

    expect(await screen.findByText('authenticated')).toBeInTheDocument()
    expect(global.fetch).toHaveBeenCalledTimes(1)
  })

  it('取得済みのトークンが再通知されても取得し直さない', async () => {
    await renderProvider()
    expect(await screen.findByText('authenticated')).toBeInTheDocument()

    await emitAndSettle('SIGNED_IN', session(ACCESS_TOKEN))

    expect(global.fetch).toHaveBeenCalledTimes(1)
    expect(screen.getByText('authenticated')).toBeInTheDocument()
  })

  it('トークンが変われば取得する', async () => {
    await renderProvider()
    expect(await screen.findByText('authenticated')).toBeInTheDocument()

    await emitAndSettle('TOKEN_REFRESHED', session('refreshed-token'))

    expect(global.fetch).toHaveBeenCalledTimes(2)
    expect(localStorage.getItem('auth_token')).toBe(JSON.stringify('refreshed-token'))
  })

  it('サインアウト後は同じトークンでも取得し直してログイン状態に戻す', async () => {
    await renderProvider()
    expect(await screen.findByText('authenticated')).toBeInTheDocument()

    await emitAndSettle('SIGNED_OUT', null)
    expect(screen.getByText('anonymous')).toBeInTheDocument()

    await emitAndSettle('SIGNED_IN', session(ACCESS_TOKEN))

    expect(global.fetch).toHaveBeenCalledTimes(2)
    expect(screen.getByText('authenticated')).toBeInTheDocument()
  })

  it('メンテナンス中で取得できなかったトークンは、次の通知で取得し直す', async () => {
    const body = { code: 'MAINTENANCE', ends_at: '2026-10-25T06:00:00Z' }
    const res = { ok: false, status: 503, json: async () => body }
    ;(global.fetch as jest.Mock).mockResolvedValueOnce({ ...res, clone: () => res })

    await renderProvider()
    await waitFor(() => expect(getMaintenanceState().active).toBe(true))
    resetMaintenance()

    await emitAndSettle('SIGNED_IN', session(ACCESS_TOKEN))

    expect(global.fetch).toHaveBeenCalledTimes(2)
    expect(screen.getByText('authenticated')).toBeInTheDocument()
  })
})
