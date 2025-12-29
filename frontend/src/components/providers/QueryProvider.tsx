'use client'

import React, { useEffect } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import axios from 'axios'
import { storage } from '@/lib/utils'
import { STORAGE_KEYS } from '@/lib/constants'

// axiosのデフォルト設定
// バックエンドAPIは /api プレフィックスを使用
const apiBaseUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
axios.defaults.baseURL = `${apiBaseUrl}/api`

// リクエストインターセプター: 認証トークンを自動付与
axios.interceptors.request.use(
  (config) => {
    if (typeof window !== 'undefined') {
      const token = storage.get(STORAGE_KEYS.TOKEN)
      if (token) {
        config.headers.Authorization = `Bearer ${token}`
      }
    }
    return config
  },
  (error) => Promise.reject(error)
)

interface QueryProviderProps {
  children: React.ReactNode
}

// QueryClientのインスタンスを作成
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5分間キャッシュを有効
      gcTime: 1000 * 60 * 10, // 10分間ガベージコレクション
      retry: 3, // 3回まで再試行
      refetchOnWindowFocus: false, // ウィンドウフォーカス時の自動再取得を無効
    },
  },
})

const QueryProvider: React.FC<QueryProviderProps> = ({ children }) => {
  return (
    <QueryClientProvider client={queryClient}>
      {children}
      {/* 開発環境でのみReact Query DevToolsを表示 */}
      {process.env.NODE_ENV === 'development' && <ReactQueryDevtools />}
    </QueryClientProvider>
  )
}

export default QueryProvider