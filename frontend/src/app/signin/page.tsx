'use client'

import React, { useState, useEffect, Suspense } from 'react'
import { useRouter, usePathname, useSearchParams } from 'next/navigation'
import Link from 'next/link'
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

const BENEFITS = [
  '今期アニメの放送状況をまとめてチェック',
  '視聴予定・視聴中・完了で記録を整理',
  'お気に入り作品をいつでも見返せる',
]

const CheckIcon: React.FC = () => (
  <svg
    width="15"
    height="15"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2.5"
    strokeLinecap="round"
    strokeLinejoin="round"
    aria-hidden="true"
  >
    <path d="M20 6 9 17l-5-5" />
  </svg>
)

const SignInPanel: React.FC = () => (
  <AuthPosterPanel
    headline={
      <>
        観た作品を、
        <br />
        そのまま記録に。
      </>
    }
  >
    <ul className="mt-[18px] flex flex-col gap-2.5 max-lg:hidden">
      {BENEFITS.map((benefit) => (
        <li key={benefit} className="flex items-center gap-2.5 text-[13.5px] text-white/80">
          <span className="flex text-primary-foreground/90">
            <CheckIcon />
          </span>
          {benefit}
        </li>
      ))}
    </ul>
  </AuthPosterPanel>
)

const SignInInner: React.FC = () => {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()
  const { loginWithGoogle } = useAuth()

  const [error, setError] = useState('')
  const [info, setInfo] = useState('')
  const [isLoading, setIsLoading] = useState(false)

  useEffect(() => {
    const reason = searchParams.get('reason')
    if (!reason) return
    if (isValidLogoutReason(reason)) {
      setInfo(LOGOUT_REASON_MESSAGES[reason])
    } else if (reason in EXTRA_MESSAGES) {
      setInfo(EXTRA_MESSAGES[reason])
    }
  }, [searchParams])

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
      {/*
        Google認証では「初回なら登録・以降はログイン」が同じ操作になるため、
        ログインと新規登録を1画面に統合している。見出しは操作そのものを表す
        「はじめる」にし、両方を担うことはリード文で説明する。
      */}
      <h1 className="mb-1.5 text-2xl font-bold tracking-tight text-foreground lg:text-[27px]">
        はじめる
      </h1>
      <p className="text-[13.5px] leading-relaxed text-muted-foreground">
        初めての方はそのまま登録が完了します。
      </p>

      <div className="mt-7 space-y-4 empty:mt-0">
        <AuthMessages error={error} info={info} />
      </div>

      {/*
        loginWithGoogle は常に prompt=select_account を付けるため、この1ボタンで
        既存アカウントでの再開と別アカウントでの新規作成の両方を賄える。
        かつて別導線として置いていた「別のGoogleアカウントを使う」は不要になった。
      */}
      <div className="mt-7">
        <GoogleSignInButton label="Googleで続ける" onClick={handleSignIn} isLoading={isLoading} />
      </div>

      {/*
        規約名を含む1文なので改行させない。iPhone SE幅（375px、実測で本文コラム
        335px）でも1行に収まるフォントサイズに合わせている。11.5pxだと
        iPhone SE幅はもちろん、このコラム幅（360px）でも収まらない。
      */}
      <p className="mt-3.5 whitespace-nowrap text-[10px] tracking-tight leading-relaxed text-muted-foreground">
        続行すると
        <Link href="/terms" className="text-primary hover:underline">
          利用規約
        </Link>
        ・
        <Link href="/privacy" className="text-primary hover:underline">
          プライバシーポリシー
        </Link>
        に同意したことになります。
      </p>

      {/* ログインが必須ではない導線。押すことに迷いのあるユーザーの離脱先を用意する */}
      <div className="mt-6 border-t border-border pt-5 text-[12.5px] text-muted-foreground">
        ログインせずに{' '}
        <Link href="/anime/current" className="font-semibold text-primary hover:underline">
          今期アニメを見る
        </Link>
      </div>
    </>
  )
}

const SignInFallback: React.FC = () => (
  <div className="animate-pulse">
    <div className="mb-3 h-7 w-56 rounded bg-muted" />
    <div className="mb-6 h-4 w-64 rounded bg-muted" />
    <div className="h-14 rounded-md bg-muted" />
  </div>
)

const SignInPage: React.FC = () => (
  <AuthLayout panel={<SignInPanel />}>
    <Suspense fallback={<SignInFallback />}>
      <SignInInner />
    </Suspense>
  </AuthLayout>
)

export default SignInPage
