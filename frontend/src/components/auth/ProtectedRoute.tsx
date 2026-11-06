'use client'

import React, { useEffect, useRef, useState } from 'react'
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
  const { isAuthenticated, isLoading, isManualLogout, isAccountDeleted } = useAuth()
  const router = useRouter()
  const [shouldRender, setShouldRender] = useState(false)
  // このページを開いたままログアウトしたかを見分ける。ログアウト後にブラウザの
  // 「戻る」でログイン必須のページへ来た場合は、通常どおりログイン画面へ送る
  const wasAuthenticated = useRef(false)

  useEffect(() => {
    if (!isLoading) {
      if (!isAuthenticated) {
        // 自分でログアウトした場合は、ログインを促さずトップへ移す（履歴を残さない）
        if (wasAuthenticated.current && isManualLogout) {
          router.replace('/')
          return
        }
        // アカウントを削除した場合は、logout が削除完了画面へ移す。ここでログイン画面へ
        // 送ると競合し、削除完了画面が見られず、戻り先つきのログイン画面が残る（#1265）
        if (wasAuthenticated.current && isAccountDeleted) return
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
        wasAuthenticated.current = true
        setShouldRender(true)
      }
    }
  }, [isAuthenticated, isLoading, isManualLogout, isAccountDeleted, router, redirectTo])

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
