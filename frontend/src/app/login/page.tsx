'use client'

import React, { useState, useEffect, Suspense } from 'react'
import Link from 'next/link'
import { useRouter, usePathname, useSearchParams } from 'next/navigation'
import { useAuth, DEFAULT_LOGIN_ERROR } from '@/contexts/AuthContext'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { AuthCard } from '@/components/auth/AuthCard'
import { AuthMessages } from '@/components/auth/AuthMessages'
import { AuthFormSkeleton } from '@/components/auth/AuthFormSkeleton'
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
      setError(err instanceof Error ? err.message : DEFAULT_LOGIN_ERROR)
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <AuthCard title="ログイン" subtitle="MyAnimeLogsへようこそ" icon="📺">
      <form onSubmit={handleSubmit} className="space-y-6">
        <AuthMessages error={error} info={info} />

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
    </AuthCard>
  )
}

const LoginFormFallback: React.FC = () => (
  <AuthFormSkeleton title="ログイン" subtitle="MyAnimeLogsへようこそ" icon="📺" />
)

const LoginPage: React.FC = () => (
  <Suspense fallback={<LoginFormFallback />}>
    <LoginFormInner />
  </Suspense>
)

export default LoginPage
