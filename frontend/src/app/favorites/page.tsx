'use client'

import React, { useState, useMemo, useCallback } from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import { Container } from '@/components/layout'
import { Button } from '@/components/ui/Button'
import { FavoriteCard } from '@/components/favorites'
import { useFavorites, FavoriteAnime } from '@/hooks/useFavorites'

// id が number のお気に入りのみを抽出
const filterValidFavorites = (favorites: FavoriteAnime[]): (FavoriteAnime & { id: number })[] => {
  return favorites.filter((anime): anime is FavoriteAnime & { id: number } =>
    typeof anime.id === 'number'
  )
}

const FavoritesPage: React.FC = () => {
  const { favorites, isLoading, error, removeFavorite } = useFavorites()
  const [updatingIds, setUpdatingIds] = useState<Set<number>>(new Set())

  const validFavorites = useMemo(() => filterValidFavorites(favorites), [favorites])

  const handleRemove = useCallback(async (animeId: number) => {
    setUpdatingIds(prev => new Set(prev).add(animeId))
    try {
      await removeFavorite(animeId)
    } finally {
      setUpdatingIds(prev => {
        const next = new Set(prev)
        next.delete(animeId)
        return next
      })
    }
  }, [removeFavorite])

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

      {isLoading ? (
        <div className="py-16 text-center">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-4 border-blue-500 border-t-transparent" />
          <p className="mt-4 text-gray-500">読み込み中...</p>
        </div>
      ) : validFavorites.length === 0 ? (
        <div className="text-center py-16">
          <div className="text-6xl mb-4" aria-hidden="true">⭐</div>
          <h2 className="text-xl font-semibold text-gray-700 mb-2">
            お気に入りがまだありません
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
          {validFavorites.map((anime) => (
            <FavoriteCard
              key={anime.id}
              anime={anime}
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
