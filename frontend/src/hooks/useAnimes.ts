'use client'

import { useState, useCallback, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getAnimes, type GetAnimesParams, type Anime } from '@/lib/api/generated/myanimelogs'

export interface SearchFilters {
  keyword?: string
  season?: string
  status?: 'airing' | 'finished'
  genre?: string
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
}

export const useAnimes = (options: UseAnimesOptions = {}): UseAnimesResult => {
  const { filters = {}, pageSize = 20, enabled = true, initialPage = 1 } = options
  const [currentPage, setCurrentPage] = useState(initialPage)

  const queryParams: GetAnimesParams = useMemo(
    () => ({
      keyword: filters.keyword || undefined,
      season: filters.season || undefined,
      status: filters.status || undefined,
      genre: filters.genre || undefined,
      page: currentPage,
      per_page: pageSize,
    }),
    [filters.keyword, filters.season, filters.status, filters.genre, currentPage, pageSize]
  )

  const { data, isLoading, error, refetch } = useQuery({
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
    staleTime: 1000 * 60 * 5, // 5分間キャッシュ
  })

  const animes = useMemo(() => {
    return (data?.data || []).filter((anime): anime is Anime & { id: number } => anime.id !== undefined)
  }, [data?.data])

  const totalPages = data?.meta?.total_pages || 1
  const total = data?.meta?.total || 0
  const hasMore = currentPage < totalPages

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
  }
}

export default useAnimes
