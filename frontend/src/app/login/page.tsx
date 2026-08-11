'use client'

import React, { useState, useEffect, Suspense } from 'react'
import Link from 'next/link'
import { useRouter, usePathname, useSearchParams } from 'next/navigation'
import { useAuth, DEFAULT_LOGIN_ERROR } from '@/contexts/AuthContext'
import { AuthLayout } from '@/components/auth/AuthLayout'
import { AuthPosterPanel } from '@/components/auth/AuthPosterPanel'
import { AuthMessages } from '@/components/auth/AuthMessages'
import { GoogleSignInButton } from '@/components/auth/GoogleSignInButton'
import { LOGOUT_REASON_MESSAGES } from '@/lib/constants'

const EXTRA_MESSAGES: Record<string, string> = {
  'auth-callback-failed': 'Googleログインを完了できませんでした。お手数ですが再度お試しください。',
}

const isValidLogoutReason = (reason: string): reason is keyof typeof LOGOUT_REASON_MESSAGES => {
  return reason in LOGOUT_REASON_MESSAGES
}

const LoginPanel: React.FC = () => (
  <AuthPosterPanel
    headline={
      <>
        観た作品を、
        <br />
        そのまま記録に。
      </>
    }
  >
    <p className="mt-3.5 max-w-[340px] text-[13.5px] leading-relaxed text-white/70 max-lg:hidden">
      今期アニメから旧作まで、視聴状況とお気に入りを1か所で管理できます。
    </p>
  </AuthPosterPanel>
)

const LoginInner: React.FC = () => {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()
  const { loginWithGoogle } = useAuth()

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

  const handleSignIn = async () => {
    setError('')
    setInfo('')
    clearReasonParam()
    setIsLoading(true)

    try {
      await loginWithGoogle()
      // 成功時はGoogleへリダイレクトするため、ローディング表示を維持したままにする
    } catch (err) {
      setError(err instanceof Error ? err.message : DEFAULT_LOGIN_ERROR)
      setIsLoading(false)
    }
  }

  return (
    <>
      <h1 className="mb-1.5 text-2xl font-bold text-foreground lg:text-[26px]">ログイン</h1>
      <p className="mb-6 text-[13.5px] text-muted-foreground">
        アカウントにアクセスして記録を続けましょう
      </p>

      <div className="mb-5 empty:mb-0 space-y-4">
        <AuthMessages error={error} info={info} />
      </div>

      <GoogleSignInButton label="Googleでログイン" onClick={handleSignIn} isLoading={isLoading} />

      <div className="mt-6 border-t border-border pt-5 text-[13px] text-muted-foreground">
        アカウントをお持ちでない方は{' '}
        <Link href="/register" className="font-semibold text-primary hover:underline">
          新規登録
        </Link>
      </div>
    </>
  )
}

const LoginFallback: React.FC = () => (
  <div className="animate-pulse">
    <div className="mb-3 h-7 w-32 rounded bg-muted" />
    <div className="mb-6 h-4 w-56 rounded bg-muted" />
    <div className="h-14 rounded-md bg-muted" />
  </div>
)

const LoginPage: React.FC = () => (
  <AuthLayout panel={<LoginPanel />}>
    <Suspense fallback={<LoginFallback />}>
      <LoginInner />
    </Suspense>
  </AuthLayout>
)

export default LoginPage
