'use client'

import { useState } from 'react'
import Link from 'next/link'
import { supabase } from '@/lib/supabase'
import { translateAuthError } from '@/contexts/AuthContext'
import { Container } from '@/components/layout'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'

const RESET_PASSWORD_DEFAULT_ERROR =
  'リセットメールの送信に失敗しました。しばらく時間をおいて再度お試しください'

export default function ResetPasswordPage() {
  const [email, setEmail] = useState('')
  const [sent, setSent] = useState(false)
  const [error, setError] = useState('')
  const [isLoading, setIsLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setIsLoading(true)

    const { error } = await supabase.auth.resetPasswordForEmail(email, {
      redirectTo: `${window.location.origin}/auth/update-password`,
    })

    setIsLoading(false)

    if (error) {
      setError(translateAuthError(error, RESET_PASSWORD_DEFAULT_ERROR))
    } else {
      setSent(true)
    }
  }

  if (sent) {
    return (
      <Container className="py-12">
        <div className="max-w-md mx-auto">
          <div className="bg-green-50 border border-green-200 rounded-lg p-6 text-center">
            <p className="font-bold text-green-800 mb-2">メールを送信しました</p>
            <p className="text-sm text-green-700">
              パスワードリセット用のリンクを <strong>{email}</strong> に送信しました。
              メールをご確認ください。
            </p>
          </div>
          <div className="mt-6 text-center">
            <Link href="/login" className="text-sm text-blue-600 hover:underline">
              ログインページに戻る
            </Link>
          </div>
        </div>
      </Container>
    )
  }

  return (
    <Container className="py-12">
      <div className="max-w-md mx-auto">
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold text-gray-900 mb-2">パスワードリセット</h1>
          <p className="text-gray-600 text-sm">
            登録済みのメールアドレスを入力してください。リセット用リンクをお送りします。
          </p>
        </div>

        <div className="bg-white rounded-lg border border-gray-200 p-8 shadow-sm">
          <form onSubmit={handleSubmit} className="space-y-5">
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

            <Button
              type="submit"
              variant="primary"
              size="lg"
              isLoading={isLoading}
              className="w-full"
            >
              リセットメールを送信
            </Button>
          </form>

          <div className="mt-6 pt-6 border-t border-gray-200 text-center">
            <Link href="/login" className="text-sm text-blue-600 hover:text-blue-700 transition-colors">
              ログインページに戻る
            </Link>
          </div>
        </div>
      </div>
    </Container>
  )
}
