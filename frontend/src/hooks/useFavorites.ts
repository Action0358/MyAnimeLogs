'use client'

import { useCallback, useMemo } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetUsersMeFavorites,
  usePutUsersMeFavoritesAnimeId,
  useDeleteUsersMeFavoritesAnimeId,
  getGetUsersMeFavoritesQueryKey,
} from '@/lib/api/generated/myanimelogs'
import type { Anime, Favorite, WatchStatus } from '@/lib/api/generated/myanimelogs'

// バックエンドから返される AnimeWithUserData 型。
// お気に入り（ハート）と視聴ステータスは独立した別データとして同梱される。
export interface FavoriteAnime extends Anime {
  user_favorite?: Favorite
  user_watch_status?: WatchStatus
}

export interface UseFavoritesResult {
  favorites: FavoriteAnime[]
  isLoading: boolean
  error: Error | null
  addFavorite: (animeId: number) => Promise<void>
  removeFavorite: (animeId: number) => Promise<void>
  isFavorite: (animeId: number) => boolean
}

// お気に入り（ハート）のみを扱うフック。視聴ステータスは useWatchStatus を使用する。
export const useFavorites = (options: { enabled?: boolean } = {}): UseFavoritesResult => {
  const { enabled = true } = options
  const queryClient = useQueryClient()

  // API は Anime[] を返すが、実際には AnimeWithUserData[] が返される
  const { data, isLoading, error } = useGetUsersMeFavorites({ query: { enabled } })

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetUsersMeFavoritesQueryKey() })
  }

  const putMutation = usePutUsersMeFavoritesAnimeId({
    mutation: { onSuccess: invalidate },
  })

  const deleteMutation = useDeleteUsersMeFavoritesAnimeId({
    mutation: { onSuccess: invalidate },
  })

  // ?? [] は毎レンダー新しい配列を作り、下の useCallback を毎回作り直させるためメモ化する
  const favorites: FavoriteAnime[] = useMemo(() => (data?.data as FavoriteAnime[]) ?? [], [data])

  const addFavorite = useCallback(async (animeId: number) => {
    await putMutation.mutateAsync({ animeId })
  }, [putMutation])

  const removeFavorite = useCallback(async (animeId: number) => {
    await deleteMutation.mutateAsync({ animeId })
  }, [deleteMutation])

  const isFavorite = useCallback((animeId: number) => {
    return favorites.some(anime => anime.id === animeId)
  }, [favorites])

  return {
    favorites,
    isLoading,
    error: error as Error | null,
    addFavorite,
    removeFavorite,
    isFavorite,
  }
}
