'use client'

import React, { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/contexts/AuthContext'

interface ProtectedRouteProps {
  children: React.ReactNode
  redirectTo?: string
}

export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({
  children,
  redirectTo = '/login',
}) => {
  const { isAuthenticated, isLoading } = useAuth()
  const router = useRouter()
  const [shouldRender, setShouldRender] = useState(false)

  useEffect(() => {
    if (!isLoading) {
      if (!isAuthenticated) {
        // 認証されていない場合は即座にリダイレクト（履歴を残さない）
        router.replace(redirectTo as any)
      } else {
        // 認証されている場合のみレンダリングを許可
        setShouldRender(true)
      }
    }
  }, [isAuthenticated, isLoading, router, redirectTo])

  // ローディング中または認証されていない場合はローディング画面を表示
  if (isLoading || !isAuthenticated || !shouldRender) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">認証情報を確認中...</p>
        </div>
      </div>
    )
  }

  return <>{children}</>
}
