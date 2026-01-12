'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'

// 旧アニメ一覧ページ → 今期アニメページへリダイレクト
// SEO・ブックマーク対応のため、完全削除ではなくリダイレクト
export default function AnimeListPage() {
  const router = useRouter()

  useEffect(() => {
    router.replace('/anime/current')
  }, [router])

  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="text-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
        <p className="mt-4 text-gray-600">リダイレクト中...</p>
      </div>
    </div>
  )
}
