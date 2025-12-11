'use client'

import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
} from 'react'
import { useRouter } from 'next/navigation'
import { storage } from '@/lib/utils'
import { post, ApiClientError, setUnauthorizedHandler, setRefreshTokenHandler } from '@/lib/api/client'

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

interface AuthContextType {
  user: User | null
  token: string | null
  login: (email: string, password: string) => Promise<void>
  register: (name: string, email: string, password: string) => Promise<void>
  logout: () => void
  updateUser: (userData: User) => void
  isLoading: boolean
  isAuthenticated: boolean
  refreshAccessToken: () => Promise<string | null>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

// ストレージキーを定数として定義
const TOKEN_KEY = 'auth_token'
const REFRESH_TOKEN_KEY = 'refresh_token'
const USER_KEY = 'user_data'
const TOKEN_EXPIRES_AT_KEY = 'token_expires_at'

// タイムアウト設定
const INACTIVITY_TIMEOUT = 30 * 60 * 1000 // 30分

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [user, setUser] = useState<User | null>(null)
  const [token, setToken] = useState<string | null>(null)
  const [refreshToken, setRefreshToken] = useState<string | null>(null)
  const [tokenExpiresAt, setTokenExpiresAt] = useState<number | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const router = useRouter()

  // トークンリフレッシュ関数
  const refreshAccessToken = useCallback(async (): Promise<string | null> => {
    try {
      const storedRefreshToken = storage.getUnsafe(REFRESH_TOKEN_KEY)
      if (!storedRefreshToken) {
        console.warn('⚠️ No refresh token available')
        return null
      }

      console.log('🔄 Refreshing access token...')

      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: storedRefreshToken }),
      })

      if (!response.ok) {
        throw new Error('Failed to refresh token')
      }

      const data = await response.json()
      const newAccessToken = data.access_token
      const expiresIn = data.expires_in || 900 // デフォルト15分

      setToken(newAccessToken)
      setTokenExpiresAt(Date.now() + expiresIn * 1000)

      storage.set(TOKEN_KEY, newAccessToken)
      storage.set(TOKEN_EXPIRES_AT_KEY, Date.now() + expiresIn * 1000)

      console.log('✅ Token refreshed successfully')
      console.log(`⏰ New token expires in ${expiresIn}s`)

      return newAccessToken
    } catch (error) {
      console.error('❌ Token refresh failed:', error)
      // リフレッシュ失敗時はログアウト
      setUser(null)
      setToken(null)
      setRefreshToken(null)
      setTokenExpiresAt(null)
      storage.remove(TOKEN_KEY)
      storage.remove(REFRESH_TOKEN_KEY)
      storage.remove(USER_KEY)
      storage.remove(TOKEN_EXPIRES_AT_KEY)
      return null
    }
  }, [])

  // 401エラー時のグローバルハンドラーとリフレッシュハンドラーを登録
  useEffect(() => {
    setUnauthorizedHandler(() => {
      // 認証状態をクリア
      setUser(null)
      setToken(null)
      setRefreshToken(null)
      setTokenExpiresAt(null)
      storage.remove(TOKEN_KEY)
      storage.remove(REFRESH_TOKEN_KEY)
      storage.remove(USER_KEY)
      storage.remove(TOKEN_EXPIRES_AT_KEY)
    })

    setRefreshTokenHandler(refreshAccessToken)
  }, [refreshAccessToken])

  // トークンとユーザー情報の初期化
  useEffect(() => {
    const initAuth = async () => {
      try {
        const storedToken = storage.getUnsafe(TOKEN_KEY)
        const storedRefreshToken = storage.getUnsafe(REFRESH_TOKEN_KEY)
        const storedUser = storage.getUnsafe(USER_KEY)
        const storedExpiresAt = storage.getUnsafe(TOKEN_EXPIRES_AT_KEY)

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
        storage.remove(TOKEN_KEY)
        storage.remove(REFRESH_TOKEN_KEY)
        storage.remove(USER_KEY)
        storage.remove(TOKEN_EXPIRES_AT_KEY)
      } finally {
        setIsLoading(false)
      }
    }

    initAuth()
  }, [refreshAccessToken])

  const login = useCallback(
    async (email: string, password: string) => {
      try {
        const data = await post<{ 
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

        storage.set(TOKEN_KEY, accessToken)
        storage.set(REFRESH_TOKEN_KEY, refreshTokenValue)
        storage.set(TOKEN_EXPIRES_AT_KEY, Date.now() + expiresIn * 1000)
        storage.set(USER_KEY, data.user)

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
        const data = await post<{ access_token?: string; user: User }>(
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
            storage.set(TOKEN_KEY, accessToken)
            storage.set(USER_KEY, data.user)
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

  const logout = useCallback(async () => {
    try {
      const storedRefreshToken = storage.getUnsafe(REFRESH_TOKEN_KEY)
      if (storedRefreshToken) {
        // バックエンドにログアウトを通知
        await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/logout`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refresh_token: storedRefreshToken }),
        }).catch(err => console.error('Logout API error:', err))
      }
    } catch (error) {
      console.error('Logout error:', error)
    } finally {
      // トークンをクリア
      setUser(null)
      setToken(null)
      setRefreshToken(null)
      setTokenExpiresAt(null)
      storage.remove(TOKEN_KEY)
      storage.remove(REFRESH_TOKEN_KEY)
      storage.remove(USER_KEY)
      storage.remove(TOKEN_EXPIRES_AT_KEY)

      console.log('✅ Logged out')
      router.push('/')
    }
  }, [router])

  const updateUser = useCallback((userData: User) => {
    setUser(userData)
    storage.set(USER_KEY, userData)
  }, [])

  // 自動リフレッシュタイマー
  useEffect(() => {
    if (!tokenExpiresAt || !user) return

    // トークン期限の1分前にリフレッシュ
    const refreshTime = tokenExpiresAt - Date.now() - 60000

    if (refreshTime > 0) {
      console.log(`⏰ Token will be refreshed in ${Math.round(refreshTime / 1000)}s`)

      const timer = setTimeout(() => {
        console.log('🔄 Auto-refreshing token...')
        refreshAccessToken()
      }, refreshTime)

      return () => clearTimeout(timer)
    } else if (refreshTime > -60000) {
      // 期限切れから1分以内の場合は即座にリフレッシュ
      console.log('⚠️ Token expired recently, refreshing now...')
      refreshAccessToken()
    }
  }, [tokenExpiresAt, user, refreshAccessToken])

  // 非アクティブタイムアウト
  useEffect(() => {
    if (!user) return // ログインしていない場合はスキップ

    let inactivityTimer: NodeJS.Timeout

    const resetTimer = () => {
      clearTimeout(inactivityTimer)
      inactivityTimer = setTimeout(() => {
        console.log('⏱️ Inactivity timeout - logging out')
        logout()
        if (typeof window !== 'undefined') {
          alert('非アクティブのため自動ログアウトしました')
        }
      }, INACTIVITY_TIMEOUT)
    }

    // ユーザーの操作を監視
    const events = ['mousedown', 'keydown', 'scroll', 'touchstart']
    events.forEach(event => {
      document.addEventListener(event, resetTimer)
    })

    resetTimer()

    return () => {
      clearTimeout(inactivityTimer)
      events.forEach(event => {
        document.removeEventListener(event, resetTimer)
      })
    }
  }, [user, logout])

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
