'use client'

import { useState, useCallback, useEffect } from 'react'
import { useRouter, useSearchParams, usePathname } from 'next/navigation'

export interface SearchFilters {
  keyword?: string
  season?: string
  /** 複数指定できる。URLではカンマ区切りの1パラメータ（例: genre=アクション,ドラマ） */
  genre?: string[]
  year?: number
  status?: 'airing' | 'finished' | 'not_yet_aired'
  sort?: string
}

export interface UseSearchResult {
  filters: SearchFilters
  updateFilters: (newFilters: Partial<SearchFilters>) => void
  searchHistory: string[]
  addToHistory: (query: string) => void
  clearHistory: () => void
}

const SEARCH_HISTORY_KEY = 'anime_search_history'
const MAX_HISTORY_ITEMS = 10

/**
 * カンマ区切りの genre パラメータを配列に戻す。
 * 手書きURLで空要素や余分な空白が混じっても落とさずに拾う。
 */
const parseGenreParam = (raw: string | null): string[] | undefined => {
  if (!raw) return undefined

  const genres = raw
    .split(',')
    .map((genre) => genre.trim())
    .filter(Boolean)

  return genres.length > 0 ? genres : undefined
}

export const useSearch = (): UseSearchResult => {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()

  const [searchHistory, setSearchHistory] = useState<string[]>([])

  // URLパラメータからフィルターを取得
  const filters: SearchFilters = {
    keyword: searchParams.get('keyword') || undefined,
    season: searchParams.get('season') || undefined,
    genre: parseGenreParam(searchParams.get('genre')),
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
        // 配列（ジャンル）はカンマ区切りの1パラメータにまとめる。
        // 空配列は「絞り込みなし」なのでパラメータごと消す
        if (Array.isArray(value)) {
          if (value.length > 0) {
            params.set(key, value.join(','))
          } else {
            params.delete(key)
          }
          return
        }

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
    searchHistory,
    addToHistory,
    clearHistory,
  }
}

export default useSearch
