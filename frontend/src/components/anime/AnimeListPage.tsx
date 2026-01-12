'use client'

import React, { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Container } from '@/components/layout'
import { AnimeGrid } from '@/components/anime'
import { Pagination } from '@/components/ui'
import type { DBAnime, DBAnimeResponse } from '@/lib/api/annict'
import { toAnimeStatus } from '@/lib/constants'

// キャッシュ時間の定数（バックエンドと同期）
const STALE_TIME = 1000 * 60 * 15 // 15分
const GC_TIME = 1000 * 60 * 30 // 30分

// 1ページあたりの表示件数
const PAGE_SIZE = 24

interface AnimeListPageProps {
  title: string
  queryKey: string[]
  queryFn: () => Promise<DBAnimeResponse>
  emptyMessage: string
  errorMessage: string
  totalLabel?: (total: number) => string
}

export const AnimeListPage: React.FC<AnimeListPageProps> = ({
  title,
  queryKey,
  queryFn,
  emptyMessage,
  errorMessage,
  totalLabel = (total) => `${total}件のアニメ`,
}) => {
  const [currentPage, setCurrentPage] = useState(1)

  const { data, isLoading, error, isFetching } = useQuery({
    queryKey,
    queryFn,
    staleTime: STALE_TIME,
    gcTime: GC_TIME,
  })

  // クライアントサイドページネーション
  const { paginatedAnimes, totalPages, total } = useMemo(() => {
    const allAnimes = data?.animes ?? []
    const total = allAnimes.length
    const totalPages = Math.ceil(total / PAGE_SIZE)
    const startIndex = (currentPage - 1) * PAGE_SIZE
    const paginatedAnimes = allAnimes.slice(startIndex, startIndex + PAGE_SIZE)

    return { paginatedAnimes, totalPages, total }
  }, [data?.animes, currentPage])

  // ページ変更時にスクロールトップ
  const handlePageChange = (page: number) => {
    setCurrentPage(page)
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  if (error) {
    return (
      <Container className="py-8">
        <div className="text-center py-12">
          <p className="text-red-600 mb-4">{errorMessage}</p>
        </div>
      </Container>
    )
  }

  return (
    <Container className="py-8">
      <div className="max-w-7xl mx-auto">
        {/* ヘッダー */}
        <div className="mb-8">
          <div className="flex items-center gap-2">
            <h1 className="text-3xl font-bold text-gray-900">{title}</h1>
            {isFetching && !isLoading && (
              <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-blue-600"></div>
            )}
          </div>
          <p className="text-gray-600 mt-1">
            {total > 0 ? totalLabel(total) : 'アニメを読み込み中...'}
          </p>
        </div>

        {/* アニメグリッド */}
        <AnimeGrid
          animes={paginatedAnimes.map((anime: DBAnime) => ({
            id: anime.id,
            title: anime.title,
            image_url: anime.image_url,
            season: anime.season,
            status: toAnimeStatus(anime.status),
            cover_image_large: anime.cover_image_large,
            cover_image_medium: anime.cover_image_medium,
          }))}
          isLoading={isLoading}
          emptyMessage={emptyMessage}
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

export default AnimeListPage
