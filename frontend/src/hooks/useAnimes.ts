'use client'

import { useState, useCallback, useMemo, useEffect } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { getAnimes, type GetAnimesParams, type Anime } from '@/lib/api/generated/myanimelogs'

export interface SearchFilters {
  keyword?: string
  season?: string
  /** 複数指定できる。APIへはカンマ区切りの1パラメータとして送る */
  genre?: string[]
  sort?: string
  media?: string
}

export interface SortOption {
  value: string
  label: string
  direction: 'asc' | 'desc'
}

export interface UseAnimesOptions {
  filters?: SearchFilters
  pageSize?: number
  enabled?: boolean
  initialPage?: number
  prefetchNextPage?: boolean
}

export interface UseAnimesResult {
  animes: (Anime & { id: number })[]
  isLoading: boolean
  error: Error | null
  hasMore: boolean
  currentPage: number
  totalPages: number
  total: number
  setPage: (page: number) => void
  nextPage: () => void
  prevPage: () => void
  refetch: () => void
  isFetching: boolean
}

// キャッシュ時間の定数（バックエンドと同期）
const STALE_TIME = 1000 * 60 * 15 // 15分間キャッシュ
const GC_TIME = 1000 * 60 * 30 // 30分間ガベージコレクション猶予

export const useAnimes = (options: UseAnimesOptions = {}): UseAnimesResult => {
  const { 
    filters = {}, 
    pageSize = 20, 
    enabled = true, 
    initialPage = 1,
    prefetchNextPage = true 
  } = options
  const [currentPage, setCurrentPage] = useState(initialPage)
  const queryClient = useQueryClient()

  // initialPageは呼び出し側（検索ページ）がURLのpageパラメータから算出した値。
  // フィルター変更時にURLのpageが削除され1ページ目に戻るが、useState(initialPage)は
  // マウント時の値しか見ないため、このコンポーネントインスタンスが再マウントされない限り
  // 内部stateが追従せず、「ページネーション表示は1ページ目なのに表示される作品は
  // 直前のページのまま」という乖離が起きていた。initialPageの変化を検知して同期する
  useEffect(() => {
    setCurrentPage(initialPage)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [initialPage])

  // 配列のままでは毎レンダー参照が変わり useMemo が効かないため、
  // 依存にも使える文字列へ先に畳んでおく
  const genreParam = filters.genre?.length ? filters.genre.join(',') : undefined

  const queryParams: GetAnimesParams = useMemo(
    () => ({
      keyword: filters.keyword || undefined,
      season: filters.season || undefined,
      genre: genreParam,
      media: filters.media as GetAnimesParams['media'],
      sort: filters.sort as GetAnimesParams['sort'],
      page: currentPage,
      per_page: pageSize,
    }),
    [filters.keyword, filters.season, genreParam, filters.media, filters.sort, currentPage, pageSize]
  )

  const { data, isLoading, error, refetch, isFetching } = useQuery({
    queryKey: ['animes', queryParams],
    queryFn: async () => {
      try {
        const response = await getAnimes(queryParams)
        return response.data
      } catch (err) {
        console.error('Failed to fetch animes:', err)
        throw err
      }
    },
    enabled,
    staleTime: STALE_TIME,
    gcTime: GC_TIME,
  })

  const animes = useMemo(() => {
    return (data?.data || []).filter((anime): anime is Anime & { id: number } => anime.id !== undefined)
  }, [data?.data])

  const totalPages = data?.meta?.total_pages || 1
  const total = data?.meta?.total || 0
  const hasMore = currentPage < totalPages

  // 次ページをプリフェッチ
  useEffect(() => {
    if (prefetchNextPage && hasMore && !isLoading) {
      const nextPageParams: GetAnimesParams = {
        ...queryParams,
        page: currentPage + 1,
      }
      
      queryClient.prefetchQuery({
        queryKey: ['animes', nextPageParams],
        queryFn: async () => {
          const response = await getAnimes(nextPageParams)
          return response.data
        },
        staleTime: STALE_TIME,
      })
    }
  }, [queryClient, queryParams, currentPage, hasMore, isLoading, prefetchNextPage])

  const setPage = useCallback((page: number) => {
    if (page >= 1 && page <= totalPages) {
      setCurrentPage(page)
    }
  }, [totalPages])

  const nextPage = useCallback(() => {
    if (hasMore) {
      setCurrentPage((prev) => prev + 1)
    }
  }, [hasMore])

  const prevPage = useCallback(() => {
    if (currentPage > 1) {
      setCurrentPage((prev) => prev - 1)
    }
  }, [currentPage])

  return {
    animes,
    isLoading,
    error: error as Error | null,
    hasMore,
    currentPage,
    totalPages,
    total,
    setPage,
    nextPage,
    prevPage,
    refetch,
    isFetching,
  }
}

export default useAnimes
