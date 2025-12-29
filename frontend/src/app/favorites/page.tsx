'use client'

import React, { useState, useMemo, useCallback } from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import { Container } from '@/components/layout'
import { Button } from '@/components/ui/Button'
import { FavoriteTabs, FavoriteCard, TabKey, TAB_LABELS } from '@/components/favorites'
import { useFavorites, FavoriteAnime } from '@/hooks/useFavorites'
import { FavoriteStatus } from '@/lib/api/generated/myanimelogs'

// Helper to filter favorites with valid id
const filterValidFavorites = (favorites: FavoriteAnime[]): (FavoriteAnime & { id: number })[] => {
  return favorites.filter((anime): anime is FavoriteAnime & { id: number } => 
    typeof anime.id === 'number'
  )
}

const FavoritesPage: React.FC = () => {
  const [activeTab, setActiveTab] = useState<TabKey>('all')
  const { favorites, isLoading, error, updateStatus, removeFavorite } = useFavorites()
  const [updatingIds, setUpdatingIds] = useState<Set<number>>(new Set())

  // Filter out favorites without valid id
  const validFavorites = useMemo(() => filterValidFavorites(favorites), [favorites])

  const filteredFavorites = useMemo(() => {
    if (activeTab === 'all') return validFavorites
    return validFavorites.filter(anime => anime.user_favorite?.status === activeTab)
  }, [validFavorites, activeTab])

  const counts = useMemo(() => {
    const result: Record<TabKey, number> = {
      all: validFavorites.length,
      watching: 0,
      completed: 0,
      planned: 0,
      dropped: 0,
    }
    validFavorites.forEach(anime => {
      const status = anime.user_favorite?.status
      if (status && status in result) {
        result[status as FavoriteStatus]++
      }
    })
    return result
  }, [validFavorites])

  // Utility function to manage updating state
  const withUpdatingState = useCallback(async (animeId: number, action: () => Promise<void>) => {
    setUpdatingIds(prev => new Set(prev).add(animeId))
    try {
      await action()
    } finally {
      setUpdatingIds(prev => {
        const next = new Set(prev)
        next.delete(animeId)
        return next
      })
    }
  }, [])

  const handleStatusChange = useCallback(async (animeId: number, status: FavoriteStatus) => {
    await withUpdatingState(animeId, () => updateStatus(animeId, status))
  }, [withUpdatingState, updateStatus])

  const handleRemove = useCallback(async (animeId: number) => {
    await withUpdatingState(animeId, () => removeFavorite(animeId))
  }, [withUpdatingState, removeFavorite])

  if (error) {
    return (
      <Container size="lg" className="py-12">
        <div className="text-center py-16">
          <div className="text-6xl mb-4" aria-hidden="true">⚠️</div>
          <h2 className="text-xl font-semibold text-gray-700 mb-2">
            エラーが発生しました
          </h2>
          <p className="text-gray-500 mb-6">{error.message}</p>
          <Button variant="primary" onClick={() => window.location.reload()}>
            再読み込み
          </Button>
        </div>
      </Container>
    )
  }

  return (
    <Container size="lg" className="py-12">
      <h1 className="text-3xl font-bold text-gray-900 mb-6">
        お気に入りアニメ
      </h1>

      <FavoriteTabs
        activeTab={activeTab}
        onTabChange={setActiveTab}
        counts={counts}
      />

      {isLoading ? (
        <div className="py-16 text-center">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-4 border-blue-500 border-t-transparent" />
          <p className="mt-4 text-gray-500">読み込み中...</p>
        </div>
      ) : filteredFavorites.length === 0 ? (
        <div className="text-center py-16">
          <div className="text-6xl mb-4" aria-hidden="true">⭐</div>
          <h2 className="text-xl font-semibold text-gray-700 mb-2">
            {activeTab === 'all'
              ? 'お気に入りがまだありません'
              : `「${TAB_LABELS[activeTab]}」のアニメがありません`}
          </h2>
          <p className="text-gray-500 mb-6">
            気になるアニメを見つけたら、お気に入りに追加してみましょう
          </p>
          <Link href={"/" as Route}>
            <Button variant="primary" size="md">
              アニメを探す
            </Button>
          </Link>
        </div>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4 mt-6">
          {filteredFavorites.map((anime) => (
            <FavoriteCard
              key={anime.id}
              anime={anime}
              status={anime.user_favorite?.status || 'planned'}
              onStatusChange={(status) => handleStatusChange(anime.id, status)}
              onRemove={() => handleRemove(anime.id)}
              isUpdating={updatingIds.has(anime.id)}
            />
          ))}
        </div>
      )}
    </Container>
  )
}

export default FavoritesPage
