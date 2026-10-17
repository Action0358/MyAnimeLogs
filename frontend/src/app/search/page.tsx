'use client'

import React, { Suspense, useState, useCallback } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import { ChevronDown, Filter as FilterIcon } from 'lucide-react'
import { Container } from '@/components/layout'
import { AnimeGrid } from '@/components/anime'
import { Pagination } from '@/components/ui'
import { SearchForm, FilterSheet, sortOptions } from '@/components/search'
import { MediaTabs, type MediaTabDefinition } from '@/components/anime/MediaTabs'
import type { MediaType } from '@/utils/mediaFilter'
import { useAnimes } from '@/hooks/useAnimes'
import { useSearch } from '@/hooks/useSearch'

const DEFAULT_SORT = 'year'

// 今期アニメ・人気アニメページと同じメディア種別タブ（#812以降の検索絞り込み整理の続き）
const MEDIA_TABS: MediaTabDefinition[] = [
  { id: 'all', label: '全て' },
  { id: 'tv', label: 'TV' },
  { id: 'web', label: 'Web' },
  { id: 'movie', label: '映画' },
]

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
  // チップ行を出すかどうかは「実際にチップが1つ以上出るか」で判断する。
  // hasActiveFilters には media が含まれるが media に対応するチップはないため、
  // これで判定すると中身のない div が mb-6 ぶんの余白だけを作り、
  // メディアタブを切り替えるたびに以降の要素が24pxずれる（#838）。
  const hasFilterChips = Boolean(
    filters.keyword || filters.season || selectedGenres.length > 0 || (!isSortDefault && sortLabel)
  )

  const filterPanelProps = {
    selectedSeason: filters.season,
    selectedGenres: filters.genre,
    sort: filters.sort ?? DEFAULT_SORT,
    onSeasonChange: (season: string | undefined) => updateFilters({ season }),
    onGenresChange: (genre: string[]) => updateFilters({ genre }),
    onSortChange: handleSortChange,
  }

  return (
    // 幅は今期アニメ／人気アニメ（AnimeListPage）と同じ共通コンテナ（xl）に揃える
    <Container size="xl" className="py-8">
      {/* ヘッダー。見出し・件数の体裁は今期アニメ／人気アニメ（AnimeListPage）に揃える。
          件数は見出しと同じ行にベースライン揃えで併記し、視線の流れを一本化する */}
      <div className="mb-5">
        <div className="mb-4 flex items-baseline gap-2 md:gap-3 min-w-0">
          <h1 className="text-xl md:text-[26px] font-bold text-foreground tracking-tight">
            アニメ検索
          </h1>
          <span className="flex flex-shrink-0 items-baseline gap-1 text-sm text-muted-foreground">
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
            <span
              className="h-4 w-4 flex-shrink-0 animate-spin rounded-full border-b-2 border-primary"
              role="status"
              aria-label="更新中"
            />
          )}
        </div>

        {/* 検索ボックスは全幅。フィルターはメディアタブと同じ行へ移した */}
        <SearchForm
          onSearch={handleSearch}
          placeholder="タイトルで検索..."
          recentSearches={searchHistory}
          onClearHistory={clearHistory}
          initialValue={filters.keyword || ''}
          className="w-full"
        />
      </div>

      {/* モバイルフィルターシート */}
      <FilterSheet
        isOpen={showFilters}
        onClose={() => setShowFilters(false)}
        {...filterPanelProps}
      />

      {/* 検索条件表示 */}
      {hasFilterChips && (
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

      {/* メディア種別タブ + フィルター。件数は見出し横へ移したので、
          空いたこの位置にフィルターを置く。PC・SPで位置を揃えるため
          折り返さず1行に収める（タブ側が横スクロールで縮む） */}
      <div className="flex items-center justify-between gap-2.5 pb-4 border-b border-border mb-6">
        <MediaTabs tabs={MEDIA_TABS} activeTab={activeMediaTab} onTabChange={handleMediaTabChange} />
        <button
          onClick={() => setShowFilters(true)}
          className="flex flex-shrink-0 items-center justify-center gap-2 h-10 px-4 border border-input bg-background rounded-full text-sm font-medium text-foreground hover:bg-muted focus:outline-none focus:ring-2 focus:ring-ring"
        >
          <FilterIcon className="w-4 h-4" />
          {/* 狭い画面ではアイコンとバッジで用途が伝わるため、ラベルを省いて幅を稼ぐ */}
          <span className="max-sm:sr-only">フィルター</span>
          {activeFilterCount > 0 && (
            <span className="inline-flex min-w-[18px] h-[18px] items-center justify-center rounded-full bg-primary px-1 text-[11px] font-bold text-primary-foreground">
              {activeFilterCount}
            </span>
          )}
          <ChevronDown className="w-3.5 h-3.5 text-muted-foreground" />
        </button>
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
    </Container>
  )
}

const SearchPage: React.FC = () => {
  return (
    <Suspense
      fallback={
        <Container size="xl" className="py-12">
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
