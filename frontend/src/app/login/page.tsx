'use client'

import React, { useState, useEffect, Suspense } from 'react'
import Link from 'next/link'
import { useRouter, usePathname, useSearchParams } from 'next/navigation'
import { useAuth } from '@/contexts/AuthContext'
import { Container } from '@/components/layout'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { LOGOUT_REASON_MESSAGES } from '@/lib/constants'

const EXTRA_MESSAGES: Record<string, string> = {
  'confirm-email': '確認メールを送信しました。受信トレイをご確認のうえ、リンクをクリックしてログインしてください。',
  'auth-callback-failed': '認証リンクの有効期限が切れているか、無効です。お手数ですが再度お試しください。',
  'password-reset-success': 'パスワードを変更しました。新しいパスワードでログインしてください。',
}

const isValidLogoutReason = (reason: string): reason is keyof typeof LOGOUT_REASON_MESSAGES => {
  return reason in LOGOUT_REASON_MESSAGES
}

// ログインフォームの内部コンポーネント（useSearchParamsを使用するためSuspense内で使用）
const LoginFormInner: React.FC = () => {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()
  const { login } = useAuth()
  
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [info, setInfo] = useState('')
  const [isLoading, setIsLoading] = useState(false)

  // URLパラメータからメッセージを設定
  useEffect(() => {
    const reason = searchParams.get('reason')
    if (!reason) return
    if (isValidLogoutReason(reason)) {
      setInfo(LOGOUT_REASON_MESSAGES[reason])
    } else if (reason in EXTRA_MESSAGES) {
      setInfo(EXTRA_MESSAGES[reason])
    }
  }, [searchParams])

  // URLからreasonパラメータを削除
  const clearReasonParam = () => {
    const reason = searchParams.get('reason')
    if (reason) {
      const params = new URLSearchParams(searchParams.toString())
      params.delete('reason')
      const newUrl = params.toString() ? `${pathname}?${params.toString()}` : pathname
      router.replace(newUrl)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setInfo('')
    clearReasonParam()
    setIsLoading(true)

    try {
      await login(email, password)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'ログインに失敗しました')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Container className="py-12">
      <div className="max-w-md mx-auto">
        <div className="text-center mb-8">
          <div className="w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <span className="text-blue-600 text-3xl">📺</span>
          </div>
          <h1 className="text-3xl font-bold text-gray-900 mb-2">ログイン</h1>
          <p className="text-gray-600">MyAnimeLogsへようこそ</p>
        </div>

        <div className="bg-white rounded-lg border border-gray-200 p-8 shadow-sm">
          <form onSubmit={handleSubmit} className="space-y-6">
            {info && (
              <div className="p-4 bg-blue-50 border border-blue-200 rounded-md">
                <p className="text-sm text-blue-700">{info}</p>
              </div>
            )}
            
            {error && (
              <div className="p-4 bg-red-50 border border-red-200 rounded-md">
                <p className="text-sm text-red-600">{error}</p>
              </div>
            )}

            <Input
              label="メールアドレス"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="example@email.com"
              required
              autoComplete="email"
            />

            <Input
              label="パスワード"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="8文字以上"
              required
              autoComplete="current-password"
            />

            <Button
              type="submit"
              variant="primary"
              size="lg"
              isLoading={isLoading}
              className="w-full"
            >
              ログイン
            </Button>
          </form>

          <div className="mt-4 text-center">
            <Link
              href="/auth/reset-password"
              className="text-sm text-blue-600 hover:text-blue-700 transition-colors"
            >
              パスワードをお忘れの方
            </Link>
          </div>

          <div className="mt-6 pt-6 border-t border-gray-200 text-center">
            <p className="text-sm text-gray-600">
              アカウントをお持ちでない方は{' '}
              <Link
                href="/register"
                className="text-blue-600 hover:text-blue-700 font-medium transition-colors"
              >
                新規登録
              </Link>
            </p>
          </div>
        </div>

        <div className="mt-8">
          <p className="text-center text-sm text-gray-600 mb-4">
            ログインすると以下の機能をご利用いただけます
          </p>
          <div className="grid grid-cols-2 gap-4">
            <div className="text-center p-4 bg-white rounded-lg border border-gray-200">
              <div className="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center mx-auto mb-2">
                <span className="text-blue-600 text-lg">📺</span>
              </div>
              <p className="text-sm text-gray-700 font-medium">視聴管理</p>
            </div>
            <div className="text-center p-4 bg-white rounded-lg border border-gray-200">
              <div className="w-10 h-10 bg-green-100 rounded-lg flex items-center justify-center mx-auto mb-2">
                <span className="text-green-600 text-lg">📊</span>
              </div>
              <p className="text-sm text-gray-700 font-medium">視聴統計</p>
            </div>
            <div className="text-center p-4 bg-white rounded-lg border border-gray-200">
              <div className="w-10 h-10 bg-purple-100 rounded-lg flex items-center justify-center mx-auto mb-2">
                <span className="text-purple-600 text-lg">⭐</span>
              </div>
              <p className="text-sm text-gray-700 font-medium">評価投稿</p>
            </div>
            <div className="text-center p-4 bg-white rounded-lg border border-gray-200">
              <div className="w-10 h-10 bg-orange-100 rounded-lg flex items-center justify-center mx-auto mb-2">
                <span className="text-orange-600 text-lg">🔍</span>
              </div>
              <p className="text-sm text-gray-700 font-medium">アニメ検索</p>
            </div>
          </div>
        </div>
      </div>
    </Container>
  )
}

const LoginFormFallback: React.FC = () => (
  <Container className="py-12">
    <div className="max-w-md mx-auto">
      <div className="text-center mb-8">
        <div className="w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-4">
          <span className="text-blue-600 text-3xl">📺</span>
        </div>
        <h1 className="text-3xl font-bold text-gray-900 mb-2">ログイン</h1>
        <p className="text-gray-600">MyAnimeLogsへようこそ</p>
      </div>
      <div className="bg-white rounded-lg border border-gray-200 p-8 shadow-sm">
        <div className="animate-pulse space-y-6">
          <div className="h-10 bg-gray-200 rounded"></div>
          <div className="h-10 bg-gray-200 rounded"></div>
          <div className="h-12 bg-gray-200 rounded"></div>
        </div>
      </div>
    </div>
  </Container>
)

const LoginPage: React.FC = () => (
  <Suspense fallback={<LoginFormFallback />}>
    <LoginFormInner />
  </Suspense>
)

export default LoginPage
