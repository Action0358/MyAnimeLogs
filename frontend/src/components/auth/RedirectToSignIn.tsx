'use client'

import React, { useEffect } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'

/**
 * 旧 /login・/register から /signin への転送。
 *
 * 本番は静的エクスポート（output: 'export'）のため next.config の redirects()
 * が使えない。Cloudflare Pages の _redirects でも実現できるが、ローカルの
 * dev サーバーでは効かず環境差が生まれるため、クライアント側で転送する。
 * 認証画面のためSEOへの影響は考慮不要。
 *
 * ?reason= などのクエリは引き継ぐ。セッション切れの案内が消えてしまうため。
 */
export const RedirectToSignIn: React.FC = () => {
  const router = useRouter()
  const searchParams = useSearchParams()

  useEffect(() => {
    const query = searchParams.toString()
    router.replace(query ? `/signin?${query}` : '/signin')
  }, [router, searchParams])

  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <p className="text-sm text-muted-foreground">ログイン画面へ移動しています...</p>
    </div>
  )
}
