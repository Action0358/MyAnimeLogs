'use client'

import { useCallback } from 'react'
import { useRouter, useSearchParams, usePathname } from 'next/navigation'
import type { Route } from 'next'
import { useSearchHistory } from './useSearchHistory'

export interface SearchFilters {
  keyword?: string
  season?: string
  /** 複数指定できる。URLではカンマ区切りの1パラメータ（例: genre=アクション,ドラマ） */
  genre?: string[]
  year?: number
  sort?: string
  media?: string
}

export interface UseSearchResult {
  filters: SearchFilters
  updateFilters: (newFilters: Partial<SearchFilters>) => void
  searchHistory: string[]
  addToHistory: (query: string) => void
  removeFromHistory: (query: string) => void
  clearHistory: () => void
}

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

  const { searchHistory, addToHistory, removeFromHistory, clearHistory } = useSearchHistory()

  // URLパラメータからフィルターを取得
  const filters: SearchFilters = {
    keyword: searchParams.get('keyword') || undefined,
    season: searchParams.get('season') || undefined,
    genre: parseGenreParam(searchParams.get('genre')),
    year: searchParams.get('year') ? Number(searchParams.get('year')) : undefined,
    sort: searchParams.get('sort') || undefined,
    media: searchParams.get('media') || undefined,
  }

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
      // クエリを組み立てた動的URLのため Route 型には収まらない
      router.push((queryString ? `${pathname}?${queryString}` : pathname) as Route)
    },
    [router, pathname, searchParams]
  )

  return {
    filters,
    updateFilters,
    searchHistory,
    addToHistory,
    removeFromHistory,
    clearHistory,
  }
}

export default useSearch
