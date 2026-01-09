'use client'

import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
  useRef,
} from 'react'
import { useRouter } from 'next/navigation'
import { storage } from '@/lib/utils'
import { postInternal, ApiClientError, setUnauthorizedHandler, setRefreshTokenHandler } from '@/lib/api/client'
import { INACTIVITY_TIMEOUT, TOKEN_REFRESH_MARGIN, STORAGE_KEYS, LOGOUT_REDIRECT_PATHS, type LogoutReason } from '@/lib/constants'

interface User {
  id: number
  name: string
  email: string
  description?: string
  icon?: string
  favorite_anime?: string
  sns_link?: string
  created_at?: string
}

// Userオブジェクトの型ガード
const isUser = (value: unknown): value is User => {
  if (typeof value !== 'object' || value === null) return false
  const obj = value as Record<string, unknown>
  return (
    typeof obj.id === 'number' &&
    typeof obj.name === 'string' &&
    typeof obj.email === 'string'
  )
}

interface AuthContextType {
  user: User | null
  token: string | null
  login: (email: string, password: string) => Promise<void>
  register: (name: string, email: string, password: string) => Promise<void>
  logout: (reason?: LogoutReason) => void
  updateUser: (userData: User) => void
  isLoading: boolean
  isAuthenticated: boolean
  refreshAccessToken: () => Promise<string | null>
  lastActivityTime: number
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [user, setUser] = useState<User | null>(null)
  const [token, setToken] = useState<string | null>(null)
  const [refreshToken, setRefreshToken] = useState<string | null>(null)
  const [tokenExpiresAt, setTokenExpiresAt] = useState<number | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [lastActivityTime, setLastActivityTime] = useState<number>(() => {
    // 初期値はlocalStorageから読み込み、なければ現在時刻
    if (typeof window === 'undefined') return Date.now()
    return storage.getNumber(STORAGE_KEYS.LAST_ACTIVITY_TIMESTAMP) ?? Date.now()
  })
  const router = useRouter()

  // localStorageから最終アクティビティ時刻を取得
  const getLastActivityFromStorage = useCallback((): number => {
    return storage.getNumber(STORAGE_KEYS.LAST_ACTIVITY_TIMESTAMP) ?? Date.now()
  }, [])

  // logoutをrefで保持（useEffectの依存配列から外すため）
  const logoutRef = useRef<(reason?: LogoutReason) => Promise<void>>()

  // 認証情報をクリアするヘルパー関数
  const clearAuthData = useCallback((redirectToLogin = true) => {
    setUser(null)
    setToken(null)
    setRefreshToken(null)
    setTokenExpiresAt(null)
    storage.remove(STORAGE_KEYS.TOKEN)
    storage.remove(STORAGE_KEYS.REFRESH_TOKEN)
    storage.remove(STORAGE_KEYS.USER)
    storage.remove(STORAGE_KEYS.TOKEN_EXPIRES_AT)
    storage.remove(STORAGE_KEYS.LAST_ACTIVITY_TIMESTAMP)
    
    if (redirectToLogin) {
      router.push('/login?reason=unauthorized')
    }
  }, [router])

  // トークンリフレッシュ関数（API呼び出しに専念）
  const refreshAccessToken = useCallback(async (): Promise<string | null> => {
    const storedRefreshToken = storage.getString(STORAGE_KEYS.REFRESH_TOKEN)
    if (!storedRefreshToken) {
      console.warn('⚠️ No refresh token available')
      throw new Error('No refresh token available')
    }

    console.log('🔄 Refreshing access token...')

    // postInternalを使用（無限ループを避けるため）
    const data = await postInternal<{ access_token: string; expires_in: number }>(
      '/api/auth/refresh',
      { refresh_token: storedRefreshToken }
    )

    const newAccessToken = data.access_token
    const expiresIn = data.expires_in || 900 // デフォルト15分

    setToken(newAccessToken)
    setTokenExpiresAt(Date.now() + expiresIn * 1000)

    storage.set(STORAGE_KEYS.TOKEN, newAccessToken)
    storage.set(STORAGE_KEYS.TOKEN_EXPIRES_AT, Date.now() + expiresIn * 1000)

    console.log('✅ Token refreshed successfully')
    console.log(`⏰ New token expires in ${expiresIn}s`)

    return newAccessToken
  }, [])

  // 401エラー時のグローバルハンドラーとリフレッシュハンドラーを登録
  useEffect(() => {
    setUnauthorizedHandler(() => clearAuthData(true))
    setRefreshTokenHandler(refreshAccessToken)
  }, [clearAuthData, refreshAccessToken])

  // トークンとユーザー情報の初期化
  useEffect(() => {
    const initAuth = async () => {
      try {
        const storedToken = storage.getString(STORAGE_KEYS.TOKEN)
        const storedRefreshToken = storage.getString(STORAGE_KEYS.REFRESH_TOKEN)
        const storedUser = storage.getObject(STORAGE_KEYS.USER, isUser)
        const storedExpiresAt = storage.getNumber(STORAGE_KEYS.TOKEN_EXPIRES_AT)

        if (storedRefreshToken && storedUser) {
          setRefreshToken(storedRefreshToken)
          setUser(storedUser)

          // トークンが期限切れまたは存在しない場合はリフレッシュ
          if (!storedToken || !storedExpiresAt || Date.now() >= storedExpiresAt) {
            console.log('⚠️ Token expired or missing, refreshing...')
            await refreshAccessToken()
          } else {
            setToken(storedToken)
            setTokenExpiresAt(storedExpiresAt)
          }
        }
      } catch (error) {
        console.error('Failed to initialize auth state:', error)
        // 不正なデータの場合はストレージをクリア（エラー回復）
        clearAuthData(false)
      } finally {
        setIsLoading(false)
      }
    }

    initAuth()
  }, [refreshAccessToken, clearAuthData])

  const login = useCallback(
    async (email: string, password: string) => {
      try {
        // postInternalを使用（ログインAPIは認証不要なので、401エラー時のリダイレクトを避ける）
        const data = await postInternal<{ 
          access_token: string
          refresh_token: string
          expires_in: number
          user: User 
        }>(
          '/api/login',
          { email, password }
        )

        // トークンを保存
        const accessToken = data.access_token
        const refreshTokenValue = data.refresh_token
        const expiresIn = data.expires_in || 900 // デフォルト15分

        setToken(accessToken)
        setRefreshToken(refreshTokenValue)
        setTokenExpiresAt(Date.now() + expiresIn * 1000)
        setUser(data.user)

        storage.set(STORAGE_KEYS.TOKEN, accessToken)
        storage.set(STORAGE_KEYS.REFRESH_TOKEN, refreshTokenValue)
        storage.set(STORAGE_KEYS.TOKEN_EXPIRES_AT, Date.now() + expiresIn * 1000)
        storage.set(STORAGE_KEYS.USER, data.user)

        console.log('✅ Login successful')
        console.log(`⏰ Token expires in ${expiresIn}s`)

        router.push('/')
      } catch (error) {
        console.error('Login error:', error)
        if (error instanceof ApiClientError) {
          throw new Error(error.message)
        }
        throw error
      }
    },
    [router]
  )

  const register = useCallback(
    async (name: string, email: string, password: string) => {
      try {
        // postInternalを使用（登録APIは認証不要なので、401エラー時のリダイレクトを避ける）
        const data = await postInternal<{ access_token?: string; user: User }>(
          '/api/register',
          { name, email, password }
        )

        // 登録後、自動的にログイン状態にする
        if (data.user) {
          // 登録APIがトークンを返す場合（OpenAPI仕様では定義されていないが、実装によっては返す可能性がある）
          const accessToken = data.access_token
          if (accessToken) {
            setToken(accessToken)
            setUser(data.user)
            storage.set(STORAGE_KEYS.TOKEN, accessToken)
            storage.set(STORAGE_KEYS.USER, data.user)
            router.push('/')
          } else {
            // トークンがない場合はログインページへ
            router.push('/login')
          }
        }
      } catch (error) {
        console.error('Register error:', error)
        if (error instanceof ApiClientError) {
          throw new Error(error.message)
        }
        throw error
      }
    },
    [router]
  )

  const logout = useCallback(async (reason: LogoutReason = 'manual') => {
    try {
      const storedRefreshToken = storage.getString(STORAGE_KEYS.REFRESH_TOKEN)
      if (storedRefreshToken) {
        // バックエンドにログアウトを通知（postInternalを使用）
        await postInternal('/api/logout', { refresh_token: storedRefreshToken })
          .catch(err => console.error('Logout API error:', err))
      }
    } catch (error) {
      console.error('Logout error:', error)
    } finally {
      // トークンをクリア（リダイレクトは自身で制御するため false を渡す）
      clearAuthData(false)

      console.log('✅ Logged out')
      
      // ログアウト理由に応じたリダイレクト先をマップから取得
      const redirectPath = LOGOUT_REDIRECT_PATHS[reason]
      router.push(redirectPath)
    }
  }, [router, clearAuthData])

  // logoutRefを常に最新のlogout関数で更新
  useEffect(() => {
    logoutRef.current = logout
  }, [logout])

  const updateUser = useCallback((userData: User) => {
    setUser(userData)
    storage.set(STORAGE_KEYS.USER, userData)
  }, [])

  // 自動リフレッシュタイマー
  useEffect(() => {
    if (!tokenExpiresAt || !user) return

    // トークン期限の1分前にリフレッシュ
    const refreshTime = tokenExpiresAt - Date.now() - TOKEN_REFRESH_MARGIN

    if (refreshTime > 0) {
      console.log(`⏰ Token will be refreshed in ${Math.round(refreshTime / 1000)}s`)

      const timer = setTimeout(() => {
        console.log('🔄 Auto-refreshing token...')
        refreshAccessToken().catch(error => {
          console.error('❌ Auto-refresh failed:', error)
          clearAuthData(true)
        })
      }, refreshTime)

      return () => clearTimeout(timer)
    } else if (refreshTime > -TOKEN_REFRESH_MARGIN) {
      // 期限切れから1分以内の場合は即座にリフレッシュ
      console.log('⚠️ Token expired recently, refreshing now...')
      refreshAccessToken().catch(error => {
        console.error('❌ Immediate refresh failed:', error)
        clearAuthData(true)
      })
    }
  }, [tokenExpiresAt, user, refreshAccessToken, clearAuthData])

  // 非アクティブタイムアウト
  useEffect(() => {
    if (!user) return // ログインしていない場合はスキップ

    let inactivityTimer: ReturnType<typeof setTimeout>

    // タイムアウト時のコールバック（共通化）
    const handleInactivity = () => {
      console.log('⏱️ Inactivity timeout - logging out')
      logoutRef.current?.('timeout')
    }

    const resetTimer = () => {
      const now = Date.now()
      setLastActivityTime(now)
      // タブ間で共有するためlocalStorageに保存
      storage.set(STORAGE_KEYS.LAST_ACTIVITY_TIMESTAMP, now)
      clearTimeout(inactivityTimer)
      inactivityTimer = setTimeout(handleInactivity, INACTIVITY_TIMEOUT)
    }

    // ページ可視性変更時のハンドラー（モバイル対応）
    // モバイルブラウザではバックグラウンド時にsetTimeoutが停止するため、
    // フォアグラウンド復帰時に経過時間をチェックする
    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible') {
        // localStorageから最新のアクティビティ時刻を取得（他タブの操作を考慮）
        const lastActivity = getLastActivityFromStorage()
        setLastActivityTime(lastActivity)
        const elapsed = Date.now() - lastActivity
        if (elapsed >= INACTIVITY_TIMEOUT) {
          console.log('⏱️ Inactivity timeout detected after returning from background')
          logoutRef.current?.('timeout')
        } else {
          // タイムアウトまでの残り時間でタイマーを再設定
          clearTimeout(inactivityTimer)
          inactivityTimer = setTimeout(handleInactivity, INACTIVITY_TIMEOUT - elapsed)
        }
      }
    }

    // モバイルブラウザ向け: pageshow/focusイベントでの復帰検知
    // visibilitychangeが発火しないケースをカバー
    const handlePageShow = (event: PageTransitionEvent) => {
      // bfcache（戻る/進むキャッシュ）からの復帰時
      if (event.persisted) {
        handleVisibilityChange()
      }
    }

    const handleFocus = () => {
      // ウィンドウがフォーカスを得た時（アプリ切り替え後など）
      handleVisibilityChange()
    }

    // 他のタブからのlocalStorage変更を検知
    const handleStorageChange = (event: StorageEvent) => {
      if (event.key === STORAGE_KEYS.LAST_ACTIVITY_TIMESTAMP && event.newValue) {
        try {
          const parsed = JSON.parse(event.newValue)
          if (typeof parsed !== 'number') return
          
          setLastActivityTime(parsed)
          clearTimeout(inactivityTimer)
          const elapsed = Date.now() - parsed
          if (elapsed >= INACTIVITY_TIMEOUT) {
            handleInactivity()
          } else {
            inactivityTimer = setTimeout(handleInactivity, INACTIVITY_TIMEOUT - elapsed)
          }
        } catch {
          // パースエラーは無視
        }
      }
    }

    // ユーザーの操作を監視
    // PC: mousedown, keydown, scroll
    // モバイル: touchstart, touchend, touchmove（スワイプ・スクロール操作）
    const events = ['mousedown', 'keydown', 'scroll', 'touchstart', 'touchend', 'touchmove']
    events.forEach(event => {
      // touchmoveはパフォーマンスのためpassiveで登録
      const options = event === 'touchmove' ? { passive: true } : undefined
      document.addEventListener(event, resetTimer, options)
    })
    document.addEventListener('visibilitychange', handleVisibilityChange)
    window.addEventListener('storage', handleStorageChange)
    window.addEventListener('pageshow', handlePageShow)
    window.addEventListener('focus', handleFocus)

    resetTimer()

    return () => {
      clearTimeout(inactivityTimer)
      events.forEach(event => {
        document.removeEventListener(event, resetTimer)
      })
      document.removeEventListener('visibilitychange', handleVisibilityChange)
      window.removeEventListener('storage', handleStorageChange)
      window.removeEventListener('pageshow', handlePageShow)
      window.removeEventListener('focus', handleFocus)
    }
  }, [user, getLastActivityFromStorage]) // logoutはrefで参照するため依存配列から除外

  const value = {
    user,
    token,
    login,
    register,
    logout,
    updateUser,
    isLoading,
    isAuthenticated: !!user && !!token,
    refreshAccessToken,
    lastActivityTime,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
