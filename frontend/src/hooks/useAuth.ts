'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import type { User } from '@/types'

interface AuthState {
    user: User | null
    isLoading: boolean
    isAuthenticated: boolean
}

export const useAuth = () => {
    const [authState, setAuthState] = useState<AuthState>({
        user: null,
        isLoading: true,
        isAuthenticated: false,
    })
    const router = useRouter()

    // 初期化時にローカルストレージからトークンを確認
    useEffect(() => {
        const initAuth = () => {
            try {
                const token = localStorage.getItem('auth_token')
                const userData = localStorage.getItem('user_data')

                if (token && userData) {
                    const user = JSON.parse(userData)
                    setAuthState({
                        user,
                        isLoading: false,
                        isAuthenticated: true,
                    })
                } else {
                    setAuthState({
                        user: null,
                        isLoading: false,
                        isAuthenticated: false,
                    })
                }
            } catch (error) {
                console.error('認証状態の初期化に失敗しました:', error)
                setAuthState({
                    user: null,
                    isLoading: false,
                    isAuthenticated: false,
                })
            }
        }

        initAuth()
    }, [])

    // ログイン処理
    const login = async (email: string, password: string) => {
        try {
            // TODO: 実際のAPI呼び出しに置き換える
            const response = await fetch('/api/auth/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ email, password }),
            })

            if (!response.ok) {
                throw new Error('ログインに失敗しました')
            }

            const data = await response.json()
            const { token, user } = data

            // トークンとユーザー情報を保存
            localStorage.setItem('auth_token', token)
            localStorage.setItem('user_data', JSON.stringify(user))

            setAuthState({
                user,
                isLoading: false,
                isAuthenticated: true,
            })

            return { success: true }
        } catch (error) {
            console.error('ログインエラー:', error)
            return {
                success: false,
                error: error instanceof Error ? error.message : 'ログインに失敗しました'
            }
        }
    }

    // ログアウト処理
    const logout = async () => {
        try {
            // トークンがある場合はサーバーにログアウト通知
            const token = localStorage.getItem('auth_token')
            if (token) {
                // TODO: 実際のAPI呼び出しに置き換える（オプション）
                await fetch('/api/auth/logout', {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${token}`,
                        'Content-Type': 'application/json',
                    },
                }).catch(() => {
                    // サーバーエラーは無視（ローカルの状態はクリアする）
                })
            }

            // ローカルストレージをクリア
            localStorage.removeItem('auth_token')
            localStorage.removeItem('user_data')

            // 認証状態をリセット
            setAuthState({
                user: null,
                isLoading: false,
                isAuthenticated: false,
            })

            // ホームページにリダイレクト
            router.push('/')

            return { success: true }
        } catch (error) {
            console.error('ログアウトエラー:', error)
            return {
                success: false,
                error: error instanceof Error ? error.message : 'ログアウトに失敗しました'
            }
        }
    }

    // ユーザー情報を更新
    const updateUser = (userData: User) => {
        localStorage.setItem('user_data', JSON.stringify(userData))
        setAuthState(prev => ({
            ...prev,
            user: userData,
        }))
    }

    return {
        ...authState,
        login,
        logout,
        updateUser,
    }
}