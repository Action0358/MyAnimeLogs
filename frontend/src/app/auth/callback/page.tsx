'use client'

import { Suspense, useEffect } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { supabase } from '@/lib/supabase'
import { Container } from '@/components/layout'

const LoadingView: React.FC = () => (
  <Container className="py-24">
    <div className="flex flex-col items-center justify-center gap-4">
      <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600" />
      <p className="text-gray-600">認証中...</p>
    </div>
  </Container>
)

const AuthCallbackInner: React.FC = () => {
  const router = useRouter()
  const searchParams = useSearchParams()

  useEffect(() => {
    const errorDescription =
      searchParams.get('error_description') ?? searchParams.get('error')
    if (errorDescription) {
      router.replace('/login?reason=auth-callback-failed')
      return
    }

    const code = searchParams.get('code')
    const finalize = async () => {
      if (code) {
        const { error } = await supabase.auth.exchangeCodeForSession(code)
        if (error) {
          router.replace('/login?reason=auth-callback-failed')
          return
        }
      }
      const { data: { session } } = await supabase.auth.getSession()
      router.replace(session ? '/' : '/login?reason=auth-callback-failed')
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
