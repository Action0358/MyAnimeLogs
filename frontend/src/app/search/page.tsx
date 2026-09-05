'use client'

import React, { Suspense, useState, useCallback, useRef } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import { ChevronDown, Filter as FilterIcon, X } from 'lucide-react'
import { Container } from '@/components/layout'
import { AnimeGrid } from '@/components/anime'
import { Pagination } from '@/components/ui'
import { SearchForm, FilterPanel, sortOptions } from '@/components/search'
import { MediaTabs, type MediaTabDefinition } from '@/components/anime/MediaTabs'
import type { MediaType } from '@/utils/mediaFilter'
import { useAnimes } from '@/hooks/useAnimes'
import { useSearch } from '@/hooks/useSearch'
import { useOnClickOutside } from '@/hooks/useOnClickOutside'

const DEFAULT_SORT = 'year'

// 今期アニメ・人気アニメページと同じメディア種別タブ（#812以降の検索絞り込み整理の続き）
const MEDIA_TABS: MediaTabDefinition[] = [
  { id: 'all', label: '全て' },
  { id: 'tv', label: 'TV' },
  { id: 'web', label: 'Web' },
  { id: 'movie', label: '映画' },
]

/**
 * フィルターパネル（画面幅を問わず同じ「ボタン→開閉パネル」の体験に揃える）。
 *
 * モバイルは下からのスライドシート、デスクトップ（lg以上）は画面中央の
 * ダイアログとして表示する。デスクトップでも常時展開のチップ列を廃止し、
 * ジャンルが18種類まで増えた現状でもツールバーが崩れないようにする。
 */
function FilterSheet({
  isOpen,
  onClose,
  ...filterProps
}: {
  isOpen: boolean
  onClose: () => void
} & React.ComponentProps<typeof FilterPanel>) {
  const sheetRef = useRef<HTMLDivElement>(null)
  useOnClickOutside(sheetRef, onClose)

  if (!isOpen) return null

  return (
    <>
      <div className="fixed inset-0 bg-black/50 z-40" aria-hidden="true" />
      <div
        ref={sheetRef}
        className="fixed inset-x-0 bottom-0 z-50 bg-background rounded-t-2xl shadow-xl max-h-[80vh] overflow-y-auto
          lg:inset-x-auto lg:bottom-auto lg:left-1/2 lg:top-1/2 lg:w-full lg:max-w-md
          lg:-translate-x-1/2 lg:-translate-y-1/2 lg:rounded-2xl"
        role="dialog"
        aria-modal="true"
        aria-label="フィルター"
      >
        <div className="flex items-center justify-between px-5 py-4 border-b border-border">
          <h2 className="text-base font-semibold text-foreground">フィルター</h2>
          <button
            onClick={onClose}
            className="p-1.5 rounded-md text-muted-foreground hover:bg-muted focus:outline-none focus:ring-2 focus:ring-ring"
            aria-label="閉じる"
          >
            <X className="h-5 w-5" />
          </button>
        </div>
        <div className="p-5">
          <FilterPanel {...filterProps} />
        </div>
      </div>
    </>
  )
}

function SearchContent() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const [showFilters, setShowFilters] = useState(false)

  const { filters, updateFilters, searchHistory, addToHistory, clearHistory } = useSearch()

  const currentPage = Number(searchParams.get('page')) || 1

  const { animes, isLoading, totalPages, total, setPage, isFetching } = useAnimes({
    filters: {
      keyword: filters.keyword,
      season: filters.season,
      genre: filters.genre,
      media: filters.media,
      sort: filters.sort ?? DEFAULT_SORT,
    },
    pageSize: 24,
    initialPage: currentPage,
    prefetchNextPage: true,
  })

  const handleSearch = useCallback(
    (query: string) => {
      addToHistory(query)
      updateFilters({ keyword: query })
    },
    [addToHistory, updateFilters]
  )

  const handlePageChange = useCallback(
    (page: number) => {
      setPage(page)
      const params = new URLSearchParams(searchParams.toString())
      params.set('page', String(page))
      router.push(`/search?${params.toString()}`)
      window.scrollTo({ top: 0, behavior: 'smooth' })
    },
    [setPage, searchParams, router]
  )

  const handleSortChange = useCallback(
    (sort: string) => {
      updateFilters({ sort })
    },
    [updateFilters]
  )

  const activeMediaTab = (filters.media as MediaType) || 'all'
  const handleMediaTabChange = useCallback(
    (tab: MediaType) => {
      updateFilters({ media: tab === 'all' ? undefined : tab })
    },
    [updateFilters]
  )

  const selectedGenres = filters.genre ?? []
  const sort = filters.sort ?? DEFAULT_SORT
  const isSortDefault = sort === DEFAULT_SORT
  const hasActiveFilters =
    filters.keyword || filters.season || selectedGenres.length > 0 || filters.media
  // フィルターボタンのバッジ件数。並び替えは絞り込みではないため0件表示には影響しないが、
  // 「適用中の設定」としてバッジ・チップには含める
  const activeFilterCount =
    (filters.season ? 1 : 0) + selectedGenres.length + (isSortDefault ? 0 : 1)
  const sortLabel = sortOptions.find((option) => option.value === sort)?.label

  const filterPanelProps = {
    selectedSeason: filters.season,
    selectedGenres: filters.genre,
    sort: filters.sort ?? DEFAULT_SORT,
    onSeasonChange: (season: string | undefined) => updateFilters({ season }),
    onGenresChange: (genre: string[]) => updateFilters({ genre }),
    onSortChange: handleSortChange,
  }

  return (
    <Container className="py-8">
      <div className="max-w-7xl mx-auto">
        {/* ヘッダー */}
        <div className="mb-5">
          {/* 狭い画面では検索フォーム自体が用途を示すため見出しを出さない。
              ページの見出し構造は保つため、読み上げには残す */}
          <h1 className="sr-only md:not-sr-only text-3xl font-bold text-foreground md:mb-5">
            アニメ検索
          </h1>

          {/* 検索ボックス + フィルターボタン。
              狭い画面では縦積み、lg以上では横並びにする（並び替えは
              フィルターパネル内に統合済みのため、ここには置かない） */}
          <div className="flex flex-col gap-2.5 lg:flex-row lg:items-center">
            <SearchForm
              onSearch={handleSearch}
              placeholder="タイトルで検索..."
              recentSearches={searchHistory}
              onClearHistory={clearHistory}
              initialValue={filters.keyword || ''}
              className="w-full lg:max-w-[440px]"
            />
            <button
              onClick={() => setShowFilters(true)}
              className="flex w-full lg:w-auto flex-shrink-0 items-center justify-center gap-2 h-10 px-4 border border-input bg-background rounded-full text-sm font-medium text-foreground hover:bg-muted focus:outline-none focus:ring-2 focus:ring-ring"
            >
              <FilterIcon className="w-4 h-4" />
              フィルター
              {activeFilterCount > 0 && (
                <span className="inline-flex min-w-[18px] h-[18px] items-center justify-center rounded-full bg-primary px-1 text-[11px] font-bold text-primary-foreground">
                  {activeFilterCount}
                </span>
              )}
              <ChevronDown className="w-3.5 h-3.5 text-muted-foreground" />
            </button>
          </div>
        </div>

        {/* モバイルフィルターシート */}
        <FilterSheet
          isOpen={showFilters}
          onClose={() => setShowFilters(false)}
          {...filterPanelProps}
        />

        {/* 検索条件表示 */}
        {(hasActiveFilters || !isSortDefault) && (
          <div className="flex flex-wrap gap-2 mb-6">
            {filters.keyword && (
              <span className="inline-flex items-center gap-1 px-3 py-1 bg-primary/10 text-primary rounded-full text-sm">
                キーワード: {filters.keyword}
                <button
                  onClick={() => updateFilters({ keyword: undefined })}
                  className="ml-1 hover:text-primary/70"
                >
                  ×
                </button>
              </span>
            )}
            {filters.season && (
              <span className="inline-flex items-center gap-1 px-3 py-1 bg-primary/10 text-primary rounded-full text-sm">
                {filters.season}
                <button
                  onClick={() => updateFilters({ season: undefined })}
                  className="ml-1 hover:text-primary/70"
                >
                  ×
                </button>
              </span>
            )}
            {/* ジャンルは複数選べるので1つずつチップにし、個別に外せるようにする */}
            {selectedGenres.map((genre) => (
              <span
                key={genre}
                className="inline-flex items-center gap-1 px-3 py-1 bg-primary/10 text-primary rounded-full text-sm"
              >
                {genre}
                <button
                  onClick={() =>
                    updateFilters({ genre: selectedGenres.filter((g) => g !== genre) })
                  }
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
                  onClick={() => updateFilters({ sort: undefined })}
                  className="ml-1 hover:text-primary/70"
                  aria-label="並び替えを既定に戻す"
                >
                  ×
                </button>
              </span>
            )}
          </div>
        )}

        {/* メディア種別タブ + 件数。今期アニメ・人気アニメページと同じ並び（タブ左・件数右）で
            区切り線を共有する */}
        <div className="flex items-center justify-between gap-2.5 pb-4 border-b border-border mb-6 text-sm text-muted-foreground">
          <MediaTabs tabs={MEDIA_TABS} activeTab={activeMediaTab} onTabChange={handleMediaTabChange} />
          <span className="flex flex-shrink-0 items-center gap-1.5">
            {!isLoading && (
              <span className="flex items-center gap-1.5 whitespace-nowrap">
                <span className="text-base font-bold text-foreground">{total.toLocaleString()}</span>
                件のアニメ
              </span>
            )}
            {isFetching && !isLoading && (
              <span className="inline-block h-3.5 w-3.5 flex-shrink-0 animate-spin rounded-full border-b-2 border-primary" />
            )}
          </span>
        </div>

        {/* アニメグリッド（一覧ページと統一の全幅グリッド） */}
        <AnimeGrid
          animes={animes}
          isLoading={isLoading}
          // 詳細から戻ったときに検索条件を失わないよう、クエリごと遷移元として渡す
          from={
            searchParams.toString()
              ? `/search?${searchParams.toString()}`
              : '/search'
          }
          emptyMessage={
            hasActiveFilters
              ? '条件に一致するアニメが見つかりませんでした'
              : 'アニメを検索してください'
          }
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

const SearchPage: React.FC = () => {
  return (
    <Suspense
      fallback={
        <Container className="py-12">
          <div className="text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
            <p className="mt-4 text-muted-foreground">読み込み中...</p>
          </div>
        </Container>
      }
    >
      <SearchContent />
    </Suspense>
  )
}

export default SearchPage
