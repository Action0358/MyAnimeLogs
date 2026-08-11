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
      await loginWithGoogle()
      // 成功時はGoogleへリダイレクトするため、ローディング表示を維持したままにする
    } catch (err) {
      setError(err instanceof Error ? err.message : DEFAULT_LOGIN_ERROR)
      setIsLoading(false)
    }
  }

  return (
    <AuthLayout panel={<RegisterPanel />}>
      <h1 className="mb-1.5 text-2xl font-bold text-foreground lg:text-[26px]">新規登録</h1>
      <p className="mb-6 text-[13.5px] text-muted-foreground">
        アカウントを作成してアニメ視聴管理を始めましょう
      </p>

      <div className="mb-5 space-y-4 empty:mb-0">
        <AuthMessages error={error} />
      </div>

      <GoogleSignInButton label="Googleで登録" onClick={handleSignUp} isLoading={isLoading} />

      <p className="mt-4 text-center text-[11.5px] leading-relaxed text-muted-foreground">
        登録すると
        <Link href="/terms" className="text-primary hover:underline">
          利用規約
        </Link>
        および
        <Link href="/privacy" className="text-primary hover:underline">
          プライバシーポリシー
        </Link>
        に同意したものとみなされます
      </p>

      <div className="mt-5 border-t border-border pt-5 text-[13px] text-muted-foreground">
        既にアカウントをお持ちの方は{' '}
        <Link href="/login" className="font-semibold text-primary hover:underline">
          ログイン
        </Link>
      </div>
    </AuthLayout>
  )
}

export default RegisterPage
