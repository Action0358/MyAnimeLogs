'use client'

import React, { Suspense, useState, useCallback, useRef } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import { SlidersHorizontal, X } from 'lucide-react'
import { Container } from '@/components/layout'
import { AnimeGrid } from '@/components/anime'
import { Pagination } from '@/components/ui'
import { SearchForm, FilterPanel, SortDropdown, sortOptions } from '@/components/search'
import { useAnimes } from '@/hooks/useAnimes'
import { useSearch } from '@/hooks/useSearch'
import { useOnClickOutside } from '@/hooks/useOnClickOutside'

const DEFAULT_SORT = 'year'

/** モバイル用フィルターシート（下からスライドイン） */
function FilterSheet({
  isOpen,
  onClose,
  ...filterProps
}: {
  isOpen: boolean
  onClose: () => void
} & Omit<React.ComponentProps<typeof FilterPanel>, 'variant'>) {
  const sheetRef = useRef<HTMLDivElement>(null)
  useOnClickOutside(sheetRef, onClose)

  if (!isOpen) return null

  return (
    <>
      <div className="fixed inset-0 bg-black/50 z-40 lg:hidden" aria-hidden="true" />
      <div
        ref={sheetRef}
        className="fixed inset-x-0 bottom-0 z-50 bg-background rounded-t-2xl shadow-xl lg:hidden max-h-[80vh] overflow-y-auto"
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
          <FilterPanel {...filterProps} variant="sheet" />
        </div>
      </div>
    </>
  )
}

function SearchContent() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const [showFilters, setShowFilters] = useState(false)

  const { filters, updateFilters, clearFilters, searchHistory, addToHistory, clearHistory } =
    useSearch()

  const currentPage = Number(searchParams.get('page')) || 1

  const { animes, isLoading, totalPages, total, setPage, isFetching } = useAnimes({
    filters: {
      keyword: filters.keyword,
      season: filters.season,
      status: filters.status,
      genre: filters.genre,
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

  const hasActiveFilters = filters.keyword || filters.season || filters.genre || filters.status

  const filterPanelProps = {
    selectedSeason: filters.season,
    selectedGenre: filters.genre,
    selectedStatus: filters.status as 'airing' | 'finished' | 'not_yet_aired' | undefined,
    onSeasonChange: (season: string | undefined) => updateFilters({ season }),
    onGenreChange: (genre: string | undefined) => updateFilters({ genre }),
    onStatusChange: (status: 'airing' | 'finished' | 'not_yet_aired' | undefined) =>
      updateFilters({ status }),
    onClear: clearFilters,
  }

  return (
    <Container className="py-8">
      <div className="max-w-7xl mx-auto">
        {/* ヘッダー */}
        <div className="mb-5">
          <h1 className="text-3xl font-bold text-foreground mb-5">アニメ検索</h1>
          <SearchForm
            onSearch={handleSearch}
            placeholder="タイトルで検索..."
            recentSearches={searchHistory}
            onClearHistory={clearHistory}
            initialValue={filters.keyword || ''}
            className="max-w-2xl lg:max-w-[650px]"
          />
        </div>

        {/* ツールバー: フィルター + 並び替え + 件数 */}
        <div className="flex items-center gap-2.5 flex-wrap pb-4 border-b border-border mb-6">
          {/* デスクトップ: インラインフィルターチップ + 並び替え（操作系を左に集約、検索ボックスと幅を揃える） */}
          <div className="hidden lg:flex items-center gap-2.5 flex-wrap max-w-2xl">
            <FilterPanel {...filterPanelProps} variant="inline" />
            <SortDropdown
              options={sortOptions}
              value={filters.sort ?? DEFAULT_SORT}
              onChange={handleSortChange}
            />
          </div>

          {/* モバイル: フィルターボタン */}
          <button
            onClick={() => setShowFilters(true)}
            className="lg:hidden inline-flex items-center gap-2 h-[38px] px-3.5 border border-input bg-background rounded-lg text-sm font-medium text-foreground hover:bg-muted focus:outline-none focus:ring-2 focus:ring-ring"
          >
            <SlidersHorizontal className="w-4 h-4" />
            フィルター
            {hasActiveFilters && <span className="w-1.5 h-1.5 bg-primary rounded-full" />}
          </button>

          {/* モバイル: 並び替え（件数と同じ行の左側にまとめる） */}
          <SortDropdown
            options={sortOptions}
            value={filters.sort ?? DEFAULT_SORT}
            onChange={handleSortChange}
            className="lg:hidden"
          />

          <div className="hidden lg:block lg:flex-1" />

          {/* 件数: 結果を説明するラベルとして右端に独立、数字のみ強調 */}
          <span className="text-sm text-muted-foreground whitespace-nowrap flex items-center gap-2">
            {total > 0 ? (
              <>
                <span className="text-base font-bold text-foreground">{total.toLocaleString()}</span>
                件
              </>
            ) : (
              !isLoading && '検索結果がありません'
            )}
            {isFetching && !isLoading && (
              <span className="inline-block animate-spin rounded-full h-3.5 w-3.5 border-b-2 border-primary" />
            )}
          </span>
        </div>

        {/* モバイルフィルターシート */}
        <FilterSheet
          isOpen={showFilters}
          onClose={() => setShowFilters(false)}
          {...filterPanelProps}
        />

        {/* 検索条件表示 */}
        {hasActiveFilters && (
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
            {filters.status && (
              <span className="inline-flex items-center gap-1 px-3 py-1 bg-primary/10 text-primary rounded-full text-sm">
                {filters.status === 'airing' ? '放送中' : filters.status === 'not_yet_aired' ? '放送予定' : '完結'}
                <button
                  onClick={() => updateFilters({ status: undefined })}
                  className="ml-1 hover:text-primary/70"
                >
                  ×
                </button>
              </span>
            )}
          </div>
        )}

        {/* アニメグリッド（一覧ページと統一の全幅グリッド） */}
        <AnimeGrid
          animes={animes}
          isLoading={isLoading}
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
