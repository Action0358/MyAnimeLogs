'use client'

import React, { Suspense, useState, useCallback } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import { Container } from '@/components/layout'
import { AnimeGrid } from '@/components/anime'
import { Pagination } from '@/components/ui'
import { SearchForm, FilterPanel, SortDropdown, sortOptions } from '@/components/search'
import { useAnimes } from '@/hooks/useAnimes'
import { useSearch } from '@/hooks/useSearch'

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

  return (
    <Container className="py-8">
      <div className="max-w-7xl mx-auto">
        {/* ヘッダー */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-4">アニメ検索</h1>
          <SearchForm
            onSearch={handleSearch}
            placeholder="タイトルで検索..."
            recentSearches={searchHistory}
            onClearHistory={clearHistory}
            initialValue={filters.keyword || ''}
            className="max-w-2xl"
          />
        </div>

        <div className="flex flex-col lg:flex-row gap-6">
          {/* サイドバー（デスクトップ） */}
          <aside className="hidden lg:block w-64 flex-shrink-0">
            <FilterPanel
              selectedSeason={filters.season}
              selectedGenre={filters.genre}
              selectedStatus={filters.status}
              onSeasonChange={(season) => updateFilters({ season })}
              onGenreChange={(genre) => updateFilters({ genre })}
              onStatusChange={(status) => updateFilters({ status })}
              onClear={clearFilters}
            />
          </aside>

          {/* メインコンテンツ */}
          <main className="flex-1">
            {/* モバイルフィルターボタン */}
            <div className="lg:hidden mb-4">
              <button
                onClick={() => setShowFilters(!showFilters)}
                className="inline-flex items-center gap-2 px-4 py-2 bg-white border border-gray-300 rounded-lg text-sm text-gray-700 hover:bg-gray-50"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"
                  />
                </svg>
                フィルター
                {hasActiveFilters && (
                  <span className="w-2 h-2 bg-blue-600 rounded-full" />
                )}
              </button>
            </div>

            {/* モバイルフィルターパネル */}
            {showFilters && (
              <div className="lg:hidden mb-6">
                <FilterPanel
                  selectedSeason={filters.season}
                  selectedGenre={filters.genre}
                  selectedStatus={filters.status}
                  onSeasonChange={(season) => updateFilters({ season })}
                  onGenreChange={(genre) => updateFilters({ genre })}
                  onStatusChange={(status) => updateFilters({ status })}
                  onClear={clearFilters}
                />
              </div>
            )}

            {/* ツールバー */}
            <div className="flex items-center justify-between mb-6">
              <div className="flex items-center gap-2 text-sm text-gray-600">
                {total > 0 ? (
                  <span>
                    {total}件のアニメが見つかりました
                  </span>
                ) : (
                  !isLoading && <span>検索結果がありません</span>
                )}
                {isFetching && !isLoading && (
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600"></div>
                )}
              </div>
              <SortDropdown
                options={sortOptions}
                value={filters.sort}
                onChange={handleSortChange}
              />
            </div>

            {/* 検索条件表示 */}
            {hasActiveFilters && (
              <div className="flex flex-wrap gap-2 mb-6">
                {filters.keyword && (
                  <span className="inline-flex items-center gap-1 px-3 py-1 bg-blue-100 text-blue-800 rounded-full text-sm">
                    キーワード: {filters.keyword}
                    <button
                      onClick={() => updateFilters({ keyword: undefined })}
                      className="ml-1 hover:text-blue-600"
                    >
                      ×
                    </button>
                  </span>
                )}
                {filters.season && (
                  <span className="inline-flex items-center gap-1 px-3 py-1 bg-green-100 text-green-800 rounded-full text-sm">
                    {filters.season}
                    <button
                      onClick={() => updateFilters({ season: undefined })}
                      className="ml-1 hover:text-green-600"
                    >
                      ×
                    </button>
                  </span>
                )}
                {filters.status && (
                  <span className="inline-flex items-center gap-1 px-3 py-1 bg-purple-100 text-purple-800 rounded-full text-sm">
                    {filters.status === 'airing' ? '放送中' : '完結'}
                    <button
                      onClick={() => updateFilters({ status: undefined })}
                      className="ml-1 hover:text-purple-600"
                    >
                      ×
                    </button>
                  </span>
                )}
              </div>
            )}

            {/* アニメグリッド */}
            <AnimeGrid
              animes={animes}
              isLoading={isLoading}
              emptyMessage={
                hasActiveFilters
                  ? '条件に一致するアニメが見つかりませんでした'
                  : 'アニメを検索してください'
              }
              columns={{ mobile: 2, tablet: 3, desktop: 4, wide: 4 }}
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
          </main>
        </div>
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
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
            <p className="mt-4 text-gray-600">読み込み中...</p>
          </div>
        </Container>
      }
    >
      <SearchContent />
    </Suspense>
  )
}

export default SearchPage
