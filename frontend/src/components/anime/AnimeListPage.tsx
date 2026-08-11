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
      <Container size="xl" className="max-w-[1180px] py-8">
        <div className="text-center py-12">
          <p className="text-destructive mb-4">{errorMessage}</p>
        </div>
      </Container>
    )
  }

  return (
    <Container size="xl" className="max-w-[1180px] py-8">
      <div>
        {/* ページヘッダー: タイトル・件数・フィルタタブを1ブロックにまとめ、
            下端の区切り線でグリッドと分離する。
            件数はタイトルと同じ行にベースライン揃えで併記し、視線の流れを一本化する。 */}
        <div className="pb-4 md:pb-[18px] border-b border-border mb-5 md:mb-6">
          <div className="flex flex-col gap-3 md:flex-row md:items-baseline md:justify-between md:gap-6">
            <div className="flex items-baseline gap-2 md:gap-3 min-w-0">
              <h1 className="text-xl md:text-[26px] font-bold text-foreground tracking-tight truncate">
                {title}
              </h1>
              <span className="text-sm text-muted-foreground flex-shrink-0 flex items-baseline gap-1">
                {isLoading ? (
                  '読み込み中...'
                ) : (
                  <>
                    <span className="text-base font-bold text-foreground">
                      {total.toLocaleString()}
                    </span>
                    件
                  </>
                )}
              </span>
              {isFetching && !isLoading && (
                <div
                  className="animate-spin rounded-full h-4 w-4 border-b-2 border-primary flex-shrink-0"
                  role="status"
                  aria-label="更新中"
                />
              )}
            </div>

            {/* メディアタイプタブ。デスクトップではヘッダー右側に置き、
                並び替えなどを将来足せる余白として左右に情報を分散させる */}
            {showMediaTabs && (
              <MediaTabs
                tabs={MEDIA_TABS}
                activeTab={activeMediaTab}
                onTabChange={handleMediaTabChange}
                className="md:flex-shrink-0"
              />
            )}
          </div>
        </div>

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
          <div className="mt-6 md:mt-9">
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
