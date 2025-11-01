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
import { post, ApiClientError, setUnauthorizedHandler } from '@/lib/api/client'

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
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

// ストレージキーを定数として定義
const TOKEN_KEY = 'auth_token'
const USER_KEY = 'user_data'

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [user, setUser] = useState<User | null>(null)
  const [token, setToken] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const router = useRouter()

  // 401エラー時のグローバルハンドラーを登録
  useEffect(() => {
    setUnauthorizedHandler(() => {
      // 認証状態をクリア
      setUser(null)
      setToken(null)
      storage.remove(TOKEN_KEY)
      storage.remove(USER_KEY)
    })
  }, [])

  // トークンとユーザー情報の初期化
  useEffect(() => {
    try {
      const storedToken = storage.getUnsafe(TOKEN_KEY)
      const storedUser = storage.getUnsafe(USER_KEY)

      if (storedToken && storedUser) {
        setToken(storedToken)
        setUser(storedUser)
      }
    } catch (error) {
      console.error('Failed to initialize auth state:', error)
      // 不正なデータの場合はストレージをクリア（エラー回復）
      storage.remove(TOKEN_KEY)
      storage.remove(USER_KEY)
    } finally {
      setIsLoading(false)
    }
  }, [])

  const login = useCallback(
    async (email: string, password: string) => {
      try {
        const data = await post<{ access_token: string; user: User }>(
          '/api/login',
          { email, password }
        )

        // OpenAPI仕様に従い、access_tokenを取得
        const accessToken = data.access_token

        setToken(accessToken)
        setUser(data.user)

        storage.set(TOKEN_KEY, accessToken)
        storage.set(USER_KEY, data.user)

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

  const logout = useCallback(() => {
    setUser(null)
    setToken(null)
    storage.remove(TOKEN_KEY)
    storage.remove(USER_KEY)
    router.push('/')
  }, [router])

  const updateUser = useCallback((userData: User) => {
    setUser(userData)
    storage.set(USER_KEY, userData)
  }, [])

  const value = {
    user,
    token,
    login,
    register,
    logout,
    updateUser,
    isLoading,
    isAuthenticated: !!user && !!token,
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
