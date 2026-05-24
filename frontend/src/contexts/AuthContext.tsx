'use client'

import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
} from 'react'
import { useRouter } from 'next/navigation'
import { supabase } from '@/lib/supabase'
import { storage } from '@/lib/utils'
import { ApiClientError, setUnauthorizedHandler, setRefreshTokenHandler } from '@/lib/api/client'
import { INACTIVITY_TIMEOUT, STORAGE_KEYS, LOGOUT_REDIRECT_PATHS, type LogoutReason } from '@/lib/constants'

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

/** localStorageに認証トークンを保存する（APIクライアントが auth_token を読む） */
const storeToken = (token: string) => {
  localStorage.setItem(STORAGE_KEYS.TOKEN, JSON.stringify(token))
}

const clearToken = () => {
  localStorage.removeItem(STORAGE_KEYS.TOKEN)
}

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null)
  const [token, setToken] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [lastActivityTime, setLastActivityTime] = useState<number>(0)
  const router = useRouter()

  const getLastActivityFromStorage = useCallback((): number => {
    return storage.getNumber(STORAGE_KEYS.LAST_ACTIVITY_TIMESTAMP) ?? Date.now()
  }, [])

  useEffect(() => {
    setLastActivityTime(getLastActivityFromStorage())
  }, [getLastActivityFromStorage])

  // バックエンドからユーザープロフィールを取得
  const fetchUserProfile = useCallback(async (accessToken: string): Promise<User | null> => {
    try {
      const res = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/users/me`,
        { headers: { Authorization: `Bearer ${accessToken}` } }
      )
      if (!res.ok) return null
      const data = await res.json()
      if (isUser(data)) return data
      return null
    } catch {
      return null
    }
  }, [])

  /** Supabaseセッションからトークン・ユーザーを同期する共通処理 */
  const syncSession = useCallback(async (accessToken: string | null) => {
    if (!accessToken) {
      setToken(null)
      setUser(null)
      clearToken()
      return
    }
    storeToken(accessToken)
    setToken(accessToken)
    const profile = await fetchUserProfile(accessToken)
    if (profile) {
      setUser(profile)
    }
  }, [fetchUserProfile])

  /** トークンリフレッシュ（APIクライアントから呼ばれる） */
  const refreshAccessToken = useCallback(async (): Promise<string | null> => {
    const { data: { session } } = await supabase.auth.getSession()
    if (session?.access_token) {
      storeToken(session.access_token)
      setToken(session.access_token)
      return session.access_token
    }
    return null
  }, [])

  const clearAuthData = useCallback((redirectToLogin = true) => {
    setUser(null)
    setToken(null)
    clearToken()
    storage.remove(STORAGE_KEYS.LAST_ACTIVITY_TIMESTAMP)

    if (redirectToLogin) {
      router.push('/login?reason=unauthorized')
    }
  }, [router])

  // グローバルハンドラー登録
  useEffect(() => {
    setUnauthorizedHandler(() => clearAuthData(true))
    setRefreshTokenHandler(refreshAccessToken)
  }, [clearAuthData, refreshAccessToken])

  // Supabase認証状態を監視して状態を初期化・同期
  useEffect(() => {
    let mounted = true

    // 起動時に既存セッションを確認
    supabase.auth.getSession().then(async ({ data: { session } }) => {
      if (!mounted) return
      if (session) {
        await syncSession(session.access_token)
      }
      setIsLoading(false)
    })

    // 認証状態変化を購読
    const { data: { subscription } } = supabase.auth.onAuthStateChange(
      async (_event, session) => {
        if (!mounted) return
        await syncSession(session?.access_token ?? null)
        if (!session) setIsLoading(false)
      }
    )

    return () => {
      mounted = false
      subscription.unsubscribe()
    }
  }, [syncSession])

  const login = useCallback(async (email: string, password: string) => {
    const { error } = await supabase.auth.signInWithPassword({ email, password })
    if (error) {
      throw new ApiClientError(error.message, 401)
    }
    router.push('/')
  }, [router])

  const register = useCallback(async (name: string, email: string, password: string) => {
    const { data, error } = await supabase.auth.signUp({
      email,
      password,
      options: { data: { name } },
    })
    if (error) {
      throw new ApiClientError(error.message, 400)
    }
    if (data.session) {
      // メール確認なし: 即座にログイン状態
      router.push('/')
    } else {
      // メール確認あり: 確認メール送信済みページへ
      router.push('/login?reason=confirm-email')
    }
  }, [router])

  const logout = useCallback(async (reason: LogoutReason = 'manual') => {
    await supabase.auth.signOut()
    clearToken()
    storage.remove(STORAGE_KEYS.LAST_ACTIVITY_TIMESTAMP)
    setUser(null)
    setToken(null)
    router.push(LOGOUT_REDIRECT_PATHS[reason])
  }, [router])

  const updateUser = useCallback((userData: User) => {
    setUser(userData)
  }, [])

  // 非アクティブタイムアウト
  useEffect(() => {
    if (!user) return

    let inactivityTimer: ReturnType<typeof setTimeout>

    const handleInactivity = () => {
      logout('timeout')
    }

    const resetTimer = () => {
      const now = Date.now()
      setLastActivityTime(now)
      storage.set(STORAGE_KEYS.LAST_ACTIVITY_TIMESTAMP, now)
      clearTimeout(inactivityTimer)
      inactivityTimer = setTimeout(handleInactivity, INACTIVITY_TIMEOUT)
    }

    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible') {
        const lastActivity = getLastActivityFromStorage()
        setLastActivityTime(lastActivity)
        const elapsed = Date.now() - lastActivity
        if (elapsed >= INACTIVITY_TIMEOUT) {
          logout('timeout')
        } else {
          clearTimeout(inactivityTimer)
          inactivityTimer = setTimeout(handleInactivity, INACTIVITY_TIMEOUT - elapsed)
        }
      }
    }

    const handlePageShow = (event: PageTransitionEvent) => {
      if (event.persisted) handleVisibilityChange()
    }

    const handleFocus = () => handleVisibilityChange()

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
        } catch { /* ignore */ }
      }
    }

    const events = ['mousedown', 'keydown', 'scroll', 'touchstart', 'touchend', 'touchmove']
    events.forEach(event => {
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
      events.forEach(event => document.removeEventListener(event, resetTimer))
      document.removeEventListener('visibilitychange', handleVisibilityChange)
      window.removeEventListener('storage', handleStorageChange)
      window.removeEventListener('pageshow', handlePageShow)
      window.removeEventListener('focus', handleFocus)
    }
  }, [user, logout, getLastActivityFromStorage])

  return (
    <AuthContext.Provider value={{
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
    }}>
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
