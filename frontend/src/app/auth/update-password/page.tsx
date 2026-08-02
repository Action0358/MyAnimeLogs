'use client'

import { Suspense, useEffect, useRef, useState } from 'react'
import Link from 'next/link'
import { useRouter, useSearchParams } from 'next/navigation'
import { supabase } from '@/lib/supabase'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { AuthCard } from '@/components/auth/AuthCard'
import { AuthMessages } from '@/components/auth/AuthMessages'
import { AuthFormSkeleton } from '@/components/auth/AuthFormSkeleton'
import { Container } from '@/components/layout'

type Status = 'verifying' | 'ready' | 'invalid' | 'submitting'

const UPDATE_PASSWORD_ERROR_MESSAGES: Record<string, string> = {
  weak_password: 'パスワードが弱すぎます。より強力なパスワードを設定してください',
  same_password: '現在のパスワードと同じパスワードは使用できません',
  over_request_rate_limit: 'しばらく時間をおいてから再試行してください',
}

// Supabase の PKCE コードはシングルユース。Strict Mode のダブル invoke や
// コンポーネントの再マウントで複数回呼ばれても同じ Promise を返すよう、
// モジュールレベルでキャッシュする。
type ExchangeResult = Awaited<ReturnType<typeof supabase.auth.exchangeCodeForSession>>
let cachedExchange: { code: string; promise: Promise<ExchangeResult> } | null = null

const exchangeCodeOnce = (code: string): Promise<ExchangeResult> => {
  if (cachedExchange?.code !== code) {
    cachedExchange = {
      code,
      promise: supabase.auth.exchangeCodeForSession(code),
    }
  }
  return cachedExchange.promise
}

const LoadingView: React.FC<{ message?: string }> = ({ message = 'リンクを確認中...' }) => (
  <Container className="py-24">
    <div className="flex flex-col items-center justify-center gap-4">
      <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600" />
      <p className="text-gray-600">{message}</p>
    </div>
  </Container>
)

const UpdatePasswordInner: React.FC = () => {
  const router = useRouter()
  const searchParams = useSearchParams()
  // Strict Mode での二重実行と code の二重消費を防ぐためのガード
  const hasRunRef = useRef(false)

  const [password, setPassword] = useState('')
  const [passwordConfirm, setPasswordConfirm] = useState('')
  const [error, setError] = useState('')
  const [status, setStatus] = useState<Status>('verifying')

  // 初回マウントで recovery code を session に交換、または既存セッションを確認
  useEffect(() => {
    if (hasRunRef.current) return
    hasRunRef.current = true

    const verify = async () => {
      const errorDescription =
        searchParams.get('error_description') ?? searchParams.get('error')
      if (errorDescription) {
        setStatus('invalid')
        return
      }

      const code = searchParams.get('code')
      if (code) {
        const { data, error } = await exchangeCodeOnce(code)
        if (error) {
          console.error('[update-password] exchangeCodeForSession failed:', error.message)
        }
        setStatus(error || !data?.session ? 'invalid' : 'ready')
        return
      }

      // code がない場合：PASSWORD_RECOVERY 経由で既にセッションが張られているか確認
      const { data: { session } } = await supabase.auth.getSession()
      setStatus(session ? 'ready' : 'invalid')
    }
    verify()
  }, [searchParams])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (password !== passwordConfirm) {
      setError('パスワードが一致しません')
      return
    }
    if (password.length < 8) {
      setError('パスワードは8文字以上で入力してください')
      return
    }

    setStatus('submitting')

    const { error: updateError } = await supabase.auth.updateUser({ password })

    if (updateError) {
      const message =
        (updateError.code && UPDATE_PASSWORD_ERROR_MESSAGES[updateError.code])
          ? UPDATE_PASSWORD_ERROR_MESSAGES[updateError.code]
          : 'パスワードの更新に失敗しました。時間をおいて再度お試しください'
      setError(message)
      setStatus('ready')
      return
    }

    // recovery セッションを破棄し、新パスワードでの再ログインを促す
    await supabase.auth.signOut()
    router.replace('/login?reason=password-reset-success')
  }

  if (status === 'verifying') {
    return <LoadingView />
  }

  if (status === 'invalid') {
    return (
      <Container className="py-12">
        <div className="max-w-md mx-auto">
          <div className="bg-red-50 border border-red-200 rounded-lg p-6 text-center">
            <p className="font-bold text-red-800 mb-2">リンクが無効です</p>
            <p className="text-sm text-red-700">
              パスワードリセットリンクの有効期限が切れているか、無効です。お手数ですが、再度パスワードリセットをお試しください。
            </p>
          </div>
          <div className="mt-6 text-center">
            <Link href="/auth/reset-password" className="text-sm text-blue-600 hover:underline">
              パスワードリセットページへ戻る
            </Link>
          </div>
        </div>
      </Container>
    )
  }

  return (
    <AuthCard
      title="新しいパスワードを設定"
      subtitle="新しいパスワードを入力してください。"
    >
      <form onSubmit={handleSubmit} className="space-y-5">
        <AuthMessages error={error} />

            <Input
              label="新しいパスワード"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="8文字以上"
              required
              autoComplete="new-password"
            />

            <Input
              label="新しいパスワード（確認）"
              type="password"
              value={passwordConfirm}
              onChange={(e) => setPasswordConfirm(e.target.value)}
              placeholder="もう一度入力"
              required
              autoComplete="new-password"
            />

        <Button
          type="submit"
          variant="primary"
          size="lg"
          isLoading={status === 'submitting'}
          className="w-full"
        >
          パスワードを更新
        </Button>
      </form>

      <div className="mt-6 pt-6 border-t border-gray-200 text-center">
        <Link href="/login" className="text-sm text-blue-600 hover:text-blue-700 transition-colors">
          ログインページに戻る
        </Link>
      </div>
    </AuthCard>
  )
}

export default function UpdatePasswordPage() {
  return (
    <Suspense fallback={<LoadingView />}>
      <UpdatePasswordInner />
    </Suspense>
  )
}
