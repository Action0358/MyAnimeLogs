'use client'

import { useState, useCallback, useEffect } from 'react'

const SEARCH_HISTORY_KEY = 'anime_search_history'
const MAX_HISTORY_ITEMS = 10

export interface UseSearchHistoryResult {
  searchHistory: string[]
  addToHistory: (query: string) => void
  clearHistory: () => void
}

/**
 * キーワード検索の履歴（新しい順・最大10件）。localStorage に保存し、
 * トップページの検索欄と作品検索画面で同じ履歴を共有する。
 */
export const useSearchHistory = (): UseSearchHistoryResult => {
  const [searchHistory, setSearchHistory] = useState<string[]>([])

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

  return { searchHistory, addToHistory, clearHistory }
}

export default useSearchHistory
