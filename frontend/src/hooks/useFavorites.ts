'use client'

import { useCallback } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetUsersMeFavorites,
  usePutUsersMeFavoritesAnimeId,
  useDeleteUsersMeFavoritesAnimeId,
  getGetUsersMeFavoritesQueryKey,
  FavoriteStatus,
} from '@/lib/api/generated/myanimelogs'
import type { Anime, Favorite } from '@/lib/api/generated/myanimelogs'

export type { FavoriteStatus }

// バックエンドから返されるAnimeWithUserData型
export interface FavoriteAnime extends Anime {
  user_favorite?: Favorite
}

export interface UseFavoritesResult {
  favorites: FavoriteAnime[]
  isLoading: boolean
  error: Error | null
  updateStatus: (animeId: number, status: FavoriteStatus) => Promise<void>
  removeFavorite: (animeId: number) => Promise<void>
  isFavorite: (animeId: number) => boolean
  getFavoriteStatus: (animeId: number) => FavoriteStatus | null
}

export const useFavorites = (statusFilter?: FavoriteStatus): UseFavoritesResult => {
  const queryClient = useQueryClient()
  
  // APIはAnime[]を返すが、実際にはAnimeWithUserData[]が返される
  const { data, isLoading, error } = useGetUsersMeFavorites(
    statusFilter ? { status: statusFilter } : undefined
  )
  
  const putMutation = usePutUsersMeFavoritesAnimeId({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: getGetUsersMeFavoritesQueryKey() })
      },
    },
  })
  
  const deleteMutation = useDeleteUsersMeFavoritesAnimeId({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: getGetUsersMeFavoritesQueryKey() })
      },
    },
  })
  
  const favorites: FavoriteAnime[] = (data?.data as FavoriteAnime[]) ?? []
  
  const updateStatus = useCallback(async (animeId: number, status: FavoriteStatus) => {
    await putMutation.mutateAsync({ animeId, data: { status } })
  }, [putMutation])
  
  const removeFavorite = useCallback(async (animeId: number) => {
    await deleteMutation.mutateAsync({ animeId })
  }, [deleteMutation])
  
  const isFavorite = useCallback((animeId: number) => {
    return favorites.some(anime => anime.id === animeId)
  }, [favorites])
  
  const getFavoriteStatus = useCallback((animeId: number): FavoriteStatus | null => {
    const anime = favorites.find(a => a.id === animeId)
    return anime?.user_favorite?.status ?? null
  }, [favorites])
  
  return {
    favorites,
    isLoading,
    error: error as Error | null,
    updateStatus,
    removeFavorite,
    isFavorite,
    getFavoriteStatus,
  }
}
