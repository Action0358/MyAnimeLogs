'use client'

import React, { Suspense, useEffect, useMemo, useRef, useState } from 'react'
import { usePathname, useRouter, useSearchParams } from 'next/navigation'
import type { Route } from 'next'
import { useQuery } from '@tanstack/react-query'
import { ChevronDown, Filter as FilterIcon } from 'lucide-react'
import { Container } from '@/components/layout'
import { AnimeGrid } from '@/components/anime'
import { MediaTabs, type MediaTabDefinition } from '@/components/anime/MediaTabs'
import { ErrorState, Pagination, getErrorStateType } from '@/components/ui'
import { FilterSheet, sortOptions } from '@/components/search'
import type { DBAnime, DBAnimeResponse } from '@/lib/api/annict'
import { toAnimeStatus } from '@/lib/constants'
import { filterAnimesByMedia, type MediaType } from '@/utils/mediaFilter'
import { filterAnimesByGenres } from '@/utils/genreFilter'
import { sortAnimes } from '@/utils/sortAnimes'

// キャッシュ時間の定数（バックエンドのCurrentSeasonAnimeCacheTTL/PopularAnimeCacheTTLと同期）
const STALE_TIME = 1000 * 60 * 30 // 30分
const GC_TIME = 1000 * 60 * 60 // 60分

const DEFAULT_SORT = 'popularity'

// メディアタイプタブの定義（タブ表示・空メッセージの両方で使用）
const MEDIA_TABS: MediaTabDefinition[] = [
  { id: 'all', label: '全て' },
  { id: 'tv', label: 'TV' },
  { id: 'web', label: 'Web' },
  { id: 'movie', label: '映画' },
]

// タブIDからラベルを取得するマップ
const mediaTabLabelMap = Object.fromEntries(MEDIA_TABS.map((tab) => [tab.id, tab.label])) as Record<
  MediaType,
  string
>

// 1ページあたりの表示件数
const PAGE_SIZE = 24

const MEDIA_TYPES = MEDIA_TABS.map((tab) => tab.id as MediaType)

/** 絞り込み条件。詳細ページから戻ったときに復元できるよう、URLのクエリに持たせる */
interface ListFilters {
  media: MediaType
  genres: string[]
  sort: string
  page: number
}

/**
 * URLのクエリを絞り込み条件に戻す。手書きURLなどで不正な値が来ても既定値に落とす。
 * パラメータ名と形式（genre はカンマ区切り）は作品検索（useSearch）と揃える。
 */
export const parseListFilters = (params: URLSearchParams): ListFilters => {
  const media = params.get('media') as MediaType | null
  const sort = params.get('sort')
  const page = Number(params.get('page'))
  return {
    media: media && MEDIA_TYPES.includes(media) ? media : 'all',
    genres: (params.get('genre') ?? '')
      .split(',')
      .map((genre) => genre.trim())
      .filter(Boolean),
    sort: sort && sortOptions.some((option) => option.value === sort) ? sort : DEFAULT_SORT,
    page: Number.isInteger(page) && page > 0 ? page : 1,
  }
}

/** 絞り込み条件をクエリ文字列にする。既定値のものは省き、URLを短く保つ */
export const buildListQuery = ({ media, genres, sort, page }: ListFilters): string => {
  const params = new URLSearchParams()
  if (media !== 'all') params.set('media', media)
  if (genres.length > 0) params.set('genre', genres.join(','))
  if (sort !== DEFAULT_SORT) params.set('sort', sort)
  if (page > 1) params.set('page', String(page))
  return params.toString()
}

interface AnimeListPageProps {
  title: string
  queryKey: string[]
  queryFn: () => Promise<DBAnimeResponse>
  emptyMessage: string
  showMediaTabs?: boolean
  /** 並び替え・ジャンルの絞り込みを出す（今期・人気アニメ向け。放送時期は固定/無関係なため出さない） */
  showGenreSort?: boolean
  columns?: {
    mobile?: number
    tablet?: number
    desktop?: number
    wide?: number
  }
}

const AnimeListContent: React.FC<AnimeListPageProps> = ({
  title,
  queryKey,
  queryFn,
  emptyMessage,
  showMediaTabs = false,
  showGenreSort = false,
  columns,
}) => {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()
  const queryString = searchParams.toString()
  const filters = useMemo(() => parseListFilters(new URLSearchParams(queryString)), [queryString])
  const { media: activeMediaTab, genres: selectedGenres, sort, page: currentPage } = filters
  const [showFilters, setShowFilters] = useState(false)
  const isInitialRender = useRef(true)

  // タブ・絞り込み変更時のスクロール処理（副作用をuseEffectで管理）
  useEffect(() => {
    // 初回レンダリング時はスクロールしない
    if (isInitialRender.current) {
      isInitialRender.current = false
      return
    }
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }, [activeMediaTab, currentPage])

  const { data, isLoading, error, isFetching, refetch } = useQuery({
    queryKey,
    queryFn,
    staleTime: STALE_TIME,
    gcTime: GC_TIME,
  })

  // メディアタイプ・ジャンルでフィルタリングし、並び替える
  const filteredAnimes = useMemo(() => {
    let result = data?.animes ?? []
    if (showMediaTabs) {
      result = filterAnimesByMedia(result, activeMediaTab)
    }
    if (showGenreSort) {
      result = filterAnimesByGenres(result, selectedGenres)
      result = sortAnimes(result, sort)
    }
    return result
  }, [data?.animes, activeMediaTab, showMediaTabs, selectedGenres, sort, showGenreSort])

  // クライアントサイドページネーション
  const { paginatedAnimes, totalPages, total } = useMemo(() => {
    const total = filteredAnimes.length
    const totalPages = Math.ceil(total / PAGE_SIZE)
    const startIndex = (currentPage - 1) * PAGE_SIZE
    const paginatedAnimes = filteredAnimes.slice(startIndex, startIndex + PAGE_SIZE)
    return { paginatedAnimes, totalPages, total }
  }, [filteredAnimes, currentPage])

  // 条件をURLへ反映する。スクロールは上の useEffect で扱うため、ルーター側では行わない
  const updateFilters = (next: Partial<ListFilters>) => {
    const query = buildListQuery({ ...filters, ...next })
    router.push((query ? `${pathname}?${query}` : pathname) as Route, { scroll: false })
  }

  // 条件を変えたら1ページ目に戻す
  const handleMediaTabChange = (tab: MediaType) => updateFilters({ media: tab, page: 1 })
  const handleGenresChange = (genres: string[]) => updateFilters({ genres, page: 1 })
  const handleSortChange = (nextSort: string) => updateFilters({ sort: nextSort, page: 1 })
  const setCurrentPage = (page: number) => updateFilters({ page })

  const isSortDefault = sort === DEFAULT_SORT
  const activeFilterCount = selectedGenres.length + (isSortDefault ? 0 : 1)
  const sortLabel = sortOptions.find((option) => option.value === sort)?.label

  if (error) {
    return (
      <Container size="xl" className="py-8">
        <h1 className="text-xl md:text-[26px] font-bold text-foreground tracking-tight">{title}</h1>
        <ErrorState type={getErrorStateType(error)} onRetry={() => refetch()} />
      </Container>
    )
  }

  return (
    <Container size="xl" className="py-8">
      <div>
        {/* ページヘッダー: タイトル・件数・フィルタタブを1ブロックにまとめ、
            下端の区切り線でグリッドと分離する。
            件数はタイトルと同じ行にベースライン揃えで併記し、視線の流れを一本化する。 */}
        <div className="pb-4 md:pb-[18px] border-b border-border mb-5 md:mb-6">
          <div className="flex flex-col gap-3 md:gap-4">
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

            {/* メディアタイプタブ + フィルター。全画面幅で見出しの下に置き、
                タブを左・フィルターを右端に揃える（検索ページと同じ配置） */}
            {(showMediaTabs || showGenreSort) && (
              <div className="flex items-center justify-between gap-2.5">
                {showMediaTabs && (
                  <MediaTabs
                    tabs={MEDIA_TABS}
                    activeTab={activeMediaTab}
                    onTabChange={handleMediaTabChange}
                  />
                )}
                {showGenreSort && (
                  <button
                    onClick={() => setShowFilters(true)}
                    className="flex flex-shrink-0 items-center justify-center gap-2 h-10 px-4 border border-input bg-background rounded-full text-sm font-medium text-foreground hover:bg-muted focus:outline-none focus:ring-2 focus:ring-ring"
                  >
                    <FilterIcon className="w-4 h-4" />
                    <span className="max-sm:sr-only">フィルター</span>
                    {activeFilterCount > 0 && (
                      <span className="inline-flex min-w-[18px] h-[18px] items-center justify-center rounded-full bg-primary px-1 text-[11px] font-bold text-primary-foreground">
                        {activeFilterCount}
                      </span>
                    )}
                    <ChevronDown className="w-3.5 h-3.5 text-muted-foreground" />
                  </button>
                )}
              </div>
            )}
          </div>

          {/* 絞り込み条件のチップ表示。ジャンルは複数選べるので1つずつ外せるようにする */}
          {showGenreSort && (selectedGenres.length > 0 || !isSortDefault) && (
            <div className="flex flex-wrap gap-2 mt-3">
              {selectedGenres.map((genre) => (
                <span
                  key={genre}
                  className="inline-flex items-center gap-1 px-3 py-1 bg-primary/10 text-primary rounded-full text-sm"
                >
                  {genre}
                  <button
                    onClick={() => handleGenresChange(selectedGenres.filter((g) => g !== genre))}
                    className="ml-1 hover:text-primary/70"
                    aria-label={`ジャンル「${genre}」の絞り込みを解除`}
                  >
                    ×
                  </button>
                </span>
              ))}
              {!isSortDefault && sortLabel && (
                <span className="inline-flex items-center gap-1 px-3 py-1 bg-primary/10 text-primary rounded-full text-sm">
                  {sortLabel}
                  <button
                    onClick={() => handleSortChange(DEFAULT_SORT)}
                    className="ml-1 hover:text-primary/70"
                    aria-label="並び替えを既定に戻す"
                  >
                    ×
                  </button>
                </span>
              )}
            </div>
          )}
        </div>

        {showGenreSort && (
          <FilterSheet
            isOpen={showFilters}
            onClose={() => setShowFilters(false)}
            showSeason={false}
            selectedGenres={selectedGenres}
            sort={sort}
            onGenresChange={handleGenresChange}
            onSortChange={handleSortChange}
          />
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
          // 詳細から戻ったときに絞り込み条件を失わないよう、クエリごと遷移元として渡す
          from={queryString ? `${pathname}?${queryString}` : pathname}
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

/** useSearchParams を使うため、静的エクスポートでは Suspense の内側に置く必要がある */
export const AnimeListPage: React.FC<AnimeListPageProps> = (props) => (
  <Suspense fallback={null}>
    <AnimeListContent {...props} />
  </Suspense>
)

export default AnimeListPage
