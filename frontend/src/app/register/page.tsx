'use client'

import React, { useState } from 'react'
import Link from 'next/link'
import { useAuth, DEFAULT_LOGIN_ERROR } from '@/contexts/AuthContext'
import { AuthLayout } from '@/components/auth/AuthLayout'
import { AuthPosterPanel } from '@/components/auth/AuthPosterPanel'
import { AuthMessages } from '@/components/auth/AuthMessages'
import { GoogleSignInButton } from '@/components/auth/GoogleSignInButton'

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

const RegisterPanel: React.FC = () => (
  <AuthPosterPanel
    headline={
      <>
        30秒で、
        <br className="max-lg:hidden" />
        視聴記録をはじめる。
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

const RegisterPage: React.FC = () => {
  const { loginWithGoogle } = useAuth()
  const [error, setError] = useState('')
  const [isLoading, setIsLoading] = useState(false)

  const handleSignUp = async () => {
    setError('')
    setIsLoading(true)

    try {
      // 新規登録の導線では、Googleにサインイン済みでも必ずアカウント選択画面を出す。
      // 別アカウントで登録したいユーザーが、既存アカウントで無言でログインして
      // しまうのを防ぐため。
      await loginWithGoogle({ forceAccountSelection: true })
      // 成功時はGoogleへリダイレクトするため、ローディング表示を維持したままにする
    } catch (err) {
      setError(err instanceof Error ? err.message : DEFAULT_LOGIN_ERROR)
      setIsLoading(false)
    }
  }

  return (
    <AuthLayout panel={<RegisterPanel />}>
      {/* 余白の考え方は login と共通（3ブロック構成）。
          同意文言はボタンへの従属関係を示すため、他より近い位置に置く。 */}
      <h1 className="mb-1.5 text-2xl font-bold tracking-tight text-foreground lg:text-[27px]">
        新規登録
      </h1>
      <p className="text-[13.5px] leading-relaxed text-muted-foreground">
        Googleアカウントで登録して、アニメ視聴管理を始めましょう。
      </p>

      <div className="mt-7 space-y-4 empty:mt-0">
        <AuthMessages error={error} />
      </div>

      <div className="mt-7">
        <GoogleSignInButton label="Googleで登録" onClick={handleSignUp} isLoading={isLoading} />
      </div>

      <p className="mt-3.5 text-[11.5px] leading-relaxed text-muted-foreground">
        登録すると
        <Link href="/terms" className="text-primary hover:underline">
          利用規約
        </Link>
        および
        <Link href="/privacy" className="text-primary hover:underline">
          プライバシーポリシー
        </Link>
        に同意したものとみなされます。
      </p>

      <div className="mt-6 border-t border-border pt-5 text-[13px] text-muted-foreground">
        既にアカウントをお持ちの方は{' '}
        <Link href="/login" className="font-semibold text-primary hover:underline">
          ログイン
        </Link>
      </div>
    </AuthLayout>
  )
}

export default RegisterPage
