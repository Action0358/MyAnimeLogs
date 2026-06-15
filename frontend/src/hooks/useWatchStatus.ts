'use client'

import { useCallback } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetUsersMeWatchStatus,
  usePutUsersMeWatchStatusAnimeId,
  useDeleteUsersMeWatchStatusAnimeId,
  getGetUsersMeWatchStatusQueryKey,
  WatchStatusStatus,
} from '@/lib/api/generated/myanimelogs'
import type { Anime, Favorite, WatchStatus } from '@/lib/api/generated/myanimelogs'

export type { WatchStatusStatus }

// バックエンドから返される AnimeWithUserData 型（視聴ステータス一覧用）
export interface WatchStatusAnime extends Anime {
  user_favorite?: Favorite
  user_watch_status?: WatchStatus
}

export interface UseWatchStatusResult {
  watchStatuses: WatchStatusAnime[]
  isLoading: boolean
  error: Error | null
  updateStatus: (animeId: number, status: WatchStatusStatus) => Promise<void>
  removeStatus: (animeId: number) => Promise<void>
  getStatus: (animeId: number) => WatchStatusStatus | null
}

// 視聴ステータス（視聴予定/視聴中/完了）を扱うフック。お気に入りとは独立。
export const useWatchStatus = (options: { enabled?: boolean } = {}): UseWatchStatusResult => {
  const { enabled = true } = options
  const queryClient = useQueryClient()

  // API は Anime[] を返すが、実際には AnimeWithUserData[] が返される
  const { data, isLoading, error } = useGetUsersMeWatchStatus(undefined, { query: { enabled } })

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetUsersMeWatchStatusQueryKey() })
  }

  const putMutation = usePutUsersMeWatchStatusAnimeId({
    mutation: { onSuccess: invalidate },
  })

  const deleteMutation = useDeleteUsersMeWatchStatusAnimeId({
    mutation: { onSuccess: invalidate },
  })

  const watchStatuses: WatchStatusAnime[] = (data?.data as WatchStatusAnime[]) ?? []

  const updateStatus = useCallback(async (animeId: number, status: WatchStatusStatus) => {
    await putMutation.mutateAsync({ animeId, data: { status } })
  }, [putMutation])

  const removeStatus = useCallback(async (animeId: number) => {
    await deleteMutation.mutateAsync({ animeId })
  }, [deleteMutation])

  const getStatus = useCallback((animeId: number): WatchStatusStatus | null => {
    const anime = watchStatuses.find(a => a.id === animeId)
    return anime?.user_watch_status?.status ?? null
  }, [watchStatuses])

  return {
    watchStatuses,
    isLoading,
    error: error as Error | null,
    updateStatus,
    removeStatus,
    getStatus,
  }
}
