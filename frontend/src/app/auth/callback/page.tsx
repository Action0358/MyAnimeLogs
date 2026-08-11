'use client'

import { Suspense, useEffect, useRef } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { supabase } from '@/lib/supabase'
import { Container } from '@/components/layout'

const LoadingView: React.FC = () => (
  <Container className="py-24">
    <div className="flex flex-col items-center justify-center gap-4">
      <div className="h-12 w-12 animate-spin rounded-full border-b-2 border-primary" />
      <p className="text-muted-foreground">認証中...</p>
    </div>
  </Container>
)

const AuthCallbackInner: React.FC = () => {
  const router = useRouter()
  const searchParams = useSearchParams()
  // Strict Modeでの二重実行と code の二重消費を防ぐためのガード
  const hasRunRef = useRef(false)

  useEffect(() => {
    if (hasRunRef.current) return
    hasRunRef.current = true

    const errorDescription =
      searchParams.get('error_description') ?? searchParams.get('error')
    if (errorDescription) {
      router.replace('/login?reason=auth-callback-failed')
      return
    }

    const code = searchParams.get('code')
    const failureUrl = '/login?reason=auth-callback-failed'
    const finalize = async () => {
      try {
        if (code) {
          // PKCEフロー: exchangeCodeForSession の戻り値にセッションが含まれるため
          // 後続の getSession() は呼ばない（ネットワーク往復を1回削減）
          const { data, error } = await supabase.auth.exchangeCodeForSession(code)
          router.replace(error || !data.session ? failureUrl : '/')
          return
        }
        // implicit フロー: URLハッシュ内のトークンをSupabaseクライアントが
        // 自動で読み取りセッションを張る。読み取り結果をここで確認する。
        const { data: { session } } = await supabase.auth.getSession()
        router.replace(session ? '/' : failureUrl)
      } catch {
        // ネットワーク/クライアント例外時もローディングのまま固まらないようにする
        router.replace(failureUrl)
      }
    }
    finalize()
  }, [router, searchParams])

  return <LoadingView />
}

export default function AuthCallbackPage() {
  return (
    <Suspense fallback={<LoadingView />}>
      <AuthCallbackInner />
    </Suspense>
  )
}
