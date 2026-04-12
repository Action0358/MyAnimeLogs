'use client'

import React, { useEffect, useMemo, useRef, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Container } from '@/components/layout'
import { AnimeGrid } from '@/components/anime'
import { MediaTabs, type MediaTabDefinition } from '@/components/anime/MediaTabs'
import { Pagination } from '@/components/ui'
import type { DBAnime, DBAnimeResponse } from '@/lib/api/annict'
import { toAnimeStatus } from '@/lib/constants'
import { filterAnimesByMedia, type MediaType } from '@/utils/mediaFilter'

// キャッシュ時間の定数（バックエンドのCurrentSeasonAnimeCacheTTL/PopularAnimeCacheTTLと同期）
const STALE_TIME = 1000 * 60 * 30 // 30分
const GC_TIME = 1000 * 60 * 60 // 60分

// メディアタイプタブの定義（タブ表示・空メッセージの両方で使用）
const MEDIA_TABS: MediaTabDefinition[] = [
  { id: 'all', label: '全て' },
  { id: 'tv', label: 'TV' },
  { id: 'web', label: 'Web' },
  { id: 'movie', label: '映画' },
]

// タブIDからラベルを取得するマップ
const mediaTabLabelMap = Object.fromEntries(
  MEDIA_TABS.map((tab) => [tab.id, tab.label])
) as Record<MediaType, string>

// 1ページあたりの表示件数
const PAGE_SIZE = 24

interface AnimeListPageProps {
  title: string
  queryKey: string[]
  queryFn: () => Promise<DBAnimeResponse>
  emptyMessage: string
  errorMessage: string
  totalLabel?: (total: number) => string
  showMediaTabs?: boolean
  columns?: {
    mobile?: number
    tablet?: number
    desktop?: number
    wide?: number
  }
}

export const AnimeListPage: React.FC<AnimeListPageProps> = ({
  title,
  queryKey,
  queryFn,
  emptyMessage,
  errorMessage,
  totalLabel = (total) => `${total}件のアニメ`,
  showMediaTabs = false,
  columns,
}) => {
  const [currentPage, setCurrentPage] = useState(1)
  const [activeMediaTab, setActiveMediaTab] = useState<MediaType>('all')
  const isInitialRender = useRef(true)

  // タブ変更・ページ変更時のスクロール処理（副作用をuseEffectで管理）
  useEffect(() => {
    // 初回レンダリング時はスクロールしない
    if (isInitialRender.current) {
      isInitialRender.current = false
      return
    }
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }, [activeMediaTab, currentPage])

  const { data, isLoading, error, isFetching } = useQuery({
    queryKey,
    queryFn,
    staleTime: STALE_TIME,
    gcTime: GC_TIME,
  })

  // メディアタイプでフィルタリング
  const filteredAnimes = useMemo(() => {
    const allAnimes = data?.animes ?? []
    if (!showMediaTabs) return allAnimes
    return filterAnimesByMedia(allAnimes, activeMediaTab)
  }, [data?.animes, activeMediaTab, showMediaTabs])

  // クライアントサイドページネーション
  const { paginatedAnimes, totalPages, total } = useMemo(() => {
    const total = filteredAnimes.length
    const totalPages = Math.ceil(total / PAGE_SIZE)
    const startIndex = (currentPage - 1) * PAGE_SIZE
    const paginatedAnimes = filteredAnimes.slice(startIndex, startIndex + PAGE_SIZE)
    return { paginatedAnimes, totalPages, total }
  }, [filteredAnimes, currentPage])

  // タブ変更時にページをリセット
  const handleMediaTabChange = (tab: MediaType) => {
    setActiveMediaTab(tab)
    setCurrentPage(1)
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
        <div className="mb-6">
          <div className="flex items-center gap-2">
            <h1 className="text-3xl font-bold text-gray-900">{title}</h1>
            {isFetching && !isLoading && (
              <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-blue-600"></div>
            )}
          </div>
          <p className="text-gray-600 mt-1">
            {isLoading ? 'アニメを読み込み中...' : totalLabel(total)}
          </p>
        </div>

        {/* メディアタイプタブ */}
        {showMediaTabs && (
          <div className="mb-6">
            <MediaTabs
              tabs={MEDIA_TABS}
              activeTab={activeMediaTab}
              onTabChange={handleMediaTabChange}
            />
          </div>
        )}

        {/* アニメグリッド */}
        <AnimeGrid
          animes={paginatedAnimes.map((anime: DBAnime) => ({
            id: anime.id,
            title: anime.title,
            image_url: anime.image_url,
            season: anime.season,
            status: toAnimeStatus(anime.status),
            released_on: anime.released_on,
            cover_image_large: anime.cover_image_large,
            cover_image_medium: anime.cover_image_medium,
          }))}
          isLoading={isLoading}
          emptyMessage={
            showMediaTabs && activeMediaTab !== 'all'
              ? `${mediaTabLabelMap[activeMediaTab]}が見つかりませんでした`
              : emptyMessage
          }
          columns={columns}
        />

        {/* ページネーション */}
        {totalPages > 1 && (
          <div className="mt-8">
            <Pagination
              currentPage={currentPage}
              totalPages={totalPages}
              onPageChange={setCurrentPage}
            />
          </div>
        )}
      </div>
    </Container>
  )
}

export default AnimeListPage
