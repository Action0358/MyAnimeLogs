'use client'

import React, { Suspense, useCallback } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import { Container } from '@/components/layout'
import { AnimeGrid } from '@/components/anime'
import { Pagination } from '@/components/ui'
import { useAnimes } from '@/hooks/useAnimes'

function AnimeListContent() {
  const searchParams = useSearchParams()
  const router = useRouter()

  const currentPage = Number(searchParams.get('page')) || 1
  const sort = searchParams.get('sort') || undefined

  const { animes, isLoading, totalPages, total, setPage } = useAnimes({
    pageSize: 24,
    initialPage: currentPage,
  })

  const handlePageChange = useCallback(
    (page: number) => {
      setPage(page)
      const params = new URLSearchParams(searchParams.toString())
      params.set('page', String(page))
      router.push(`/anime?${params.toString()}`)
      window.scrollTo({ top: 0, behavior: 'smooth' })
    },
    [setPage, searchParams, router]
  )

  return (
    <Container className="py-8">
      <div className="max-w-7xl mx-auto">
        {/* ヘッダー */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">アニメ一覧</h1>
          <p className="text-gray-600">
            {total > 0 ? `${total}件のアニメ` : 'アニメを読み込み中...'}
          </p>
        </div>

        {/* アニメグリッド */}
        <AnimeGrid
          animes={animes}
          isLoading={isLoading}
          emptyMessage="アニメが見つかりませんでした"
          columns={{ mobile: 2, tablet: 3, desktop: 4, wide: 6 }}
        />

        {/* ページネーション */}
        {totalPages > 1 && (
          <div className="mt-8">
            <Pagination
              currentPage={currentPage}
              totalPages={totalPages}
              onPageChange={handlePageChange}
            />
          </div>
        )}
      </div>
    </Container>
  )
}

const AnimeListPage: React.FC = () => {
  return (
    <Suspense
      fallback={
        <Container className="py-12">
          <div className="text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
            <p className="mt-4 text-gray-600">読み込み中...</p>
          </div>
        </Container>
      }
    >
      <AnimeListContent />
    </Suspense>
  )
}

export default AnimeListPage
