'use client'

import React, { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import type { Route } from 'next'
import { useAuth } from '@/contexts/AuthContext'
import { buildSignInHref } from '@/lib/authReturnPath'

interface ProtectedRouteProps {
  children: React.ReactNode
  /** 未ログイン時の転送先。省略時はログイン後にこのページへ戻る /signin */
  redirectTo?: string
}

export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ children, redirectTo }) => {
  const { isAuthenticated, isLoading } = useAuth()
  const router = useRouter()
  const [shouldRender, setShouldRender] = useState(false)

  useEffect(() => {
    if (!isLoading) {
      if (!isAuthenticated) {
        // 認証されていない場合は即座にリダイレクト（履歴を残さない）
        // redirectTo は呼び出し側が渡す任意のパスで、typedRoutes の Route 型には収まらない
        const target =
          redirectTo ??
          buildSignInHref(
            `${window.location.pathname}${window.location.search}${window.location.hash}`
          )
        router.replace(target as Route)
      } else {
        // 認証されている場合のみレンダリングを許可
        setShouldRender(true)
      }
    }
  }, [isAuthenticated, isLoading, router, redirectTo])

  // ローディング中またはレンダリング準備ができていない場合はローディング画面を表示
  if (isLoading || !shouldRender) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
          <p className="text-muted-foreground">認証情報を確認中...</p>
        </div>
      </div>
    )
  }

  return <>{children}</>
}
