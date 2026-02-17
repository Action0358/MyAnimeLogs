'use client'

import { useState, useCallback, useEffect } from 'react'
import { useRouter, useSearchParams, usePathname } from 'next/navigation'

export interface SearchFilters {
  keyword?: string
  season?: string
  genre?: string
  year?: number
  status?: 'airing' | 'finished' | 'not_yet_aired'
  sort?: string
}

export interface UseSearchResult {
  filters: SearchFilters
  updateFilters: (newFilters: Partial<SearchFilters>) => void
  clearFilters: () => void
  searchHistory: string[]
  addToHistory: (query: string) => void
  clearHistory: () => void
}

const SEARCH_HISTORY_KEY = 'anime_search_history'
const MAX_HISTORY_ITEMS = 10

export const useSearch = (): UseSearchResult => {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()

  const [searchHistory, setSearchHistory] = useState<string[]>([])

  // URLパラメータからフィルターを取得
  const filters: SearchFilters = {
    keyword: searchParams.get('keyword') || undefined,
    season: searchParams.get('season') || undefined,
    genre: searchParams.get('genre') || undefined,
    year: searchParams.get('year') ? Number(searchParams.get('year')) : undefined,
    status: (searchParams.get('status') as 'airing' | 'finished' | 'not_yet_aired') || undefined,
    sort: searchParams.get('sort') || undefined,
  }

  // 検索履歴をlocalStorageから読み込み
  useEffect(() => {
    if (typeof window !== 'undefined') {
      try {
        const stored = localStorage.getItem(SEARCH_HISTORY_KEY)
        if (stored) {
          setSearchHistory(JSON.parse(stored))
        }
      } catch {
        // localStorageエラーは無視
      }
    }
  }, [])

  // フィルターを更新してURLを変更
  const updateFilters = useCallback(
    (newFilters: Partial<SearchFilters>) => {
      const params = new URLSearchParams(searchParams.toString())

      Object.entries(newFilters).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== '') {
          params.set(key, String(value))
        } else {
          params.delete(key)
        }
      })

      // ページをリセット
      params.delete('page')

      const queryString = params.toString()
      router.push(queryString ? `${pathname}?${queryString}` : pathname)
    },
    [router, pathname, searchParams]
  )

  // フィルターをクリア
  const clearFilters = useCallback(() => {
    router.push(pathname)
  }, [router, pathname])

  // 検索履歴に追加
  const addToHistory = useCallback((query: string) => {
    if (!query.trim()) return

    setSearchHistory((prev) => {
      const filtered = prev.filter((item) => item !== query)
      const newHistory = [query, ...filtered].slice(0, MAX_HISTORY_ITEMS)

      if (typeof window !== 'undefined') {
        try {
          localStorage.setItem(SEARCH_HISTORY_KEY, JSON.stringify(newHistory))
        } catch {
          // localStorageエラーは無視
        }
      }

      return newHistory
    })
  }, [])

  // 検索履歴をクリア
  const clearHistory = useCallback(() => {
    setSearchHistory([])
    if (typeof window !== 'undefined') {
      try {
        localStorage.removeItem(SEARCH_HISTORY_KEY)
      } catch {
        // localStorageエラーは無視
      }
    }
  }, [])

  return {
    filters,
    updateFilters,
    clearFilters,
    searchHistory,
    addToHistory,
    clearHistory,
  }
}

export default useSearch
