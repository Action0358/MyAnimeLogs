'use client'

import React from 'react'
import { cn } from '@/lib/utils'
import { AnimeCard } from './AnimeCard'
import type { Anime } from '@/lib/api/generated/myanimelogs'

export interface AnimeGridProps {
  animes: (Anime & { id: number })[]
  isLoading?: boolean
  hasMore?: boolean
  onLoadMore?: () => void
  emptyMessage?: string
  columns?: {
    mobile?: number
    tablet?: number
    desktop?: number
    wide?: number
  }
  cardSize?: 'small' | 'medium' | 'large'
  className?: string
}

const SkeletonCard: React.FC = () => (
  <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden animate-pulse">
    <div className="aspect-[3/4] bg-gray-200" />
    <div className="p-3 space-y-2">
      <div className="h-4 bg-gray-200 rounded w-3/4" />
      <div className="h-3 bg-gray-200 rounded w-1/2" />
    </div>
  </div>
)

export const AnimeGrid: React.FC<AnimeGridProps> = ({
  animes,
  isLoading = false,
  hasMore = false,
  onLoadMore,
  emptyMessage = 'アニメが見つかりませんでした',
  columns = { mobile: 2, tablet: 3, desktop: 4, wide: 6 },
  cardSize = 'medium',
  className,
}) => {
  const gridCols = cn(
    columns.mobile === 1 && 'grid-cols-1',
    columns.mobile === 2 && 'grid-cols-2',
    columns.mobile === 3 && 'grid-cols-3',
    columns.tablet === 3 && 'sm:grid-cols-3',
    columns.tablet === 4 && 'sm:grid-cols-4',
    columns.desktop === 4 && 'md:grid-cols-4',
    columns.desktop === 5 && 'md:grid-cols-5',
    columns.wide === 5 && 'lg:grid-cols-5',
    columns.wide === 6 && 'lg:grid-cols-6'
  )

  if (isLoading && animes.length === 0) {
    return (
      <div className={cn('grid gap-4', gridCols, className)}>
        {Array.from({ length: 12 }).map((_, i) => (
          <SkeletonCard key={i} />
        ))}
      </div>
    )
  }

  if (!isLoading && animes.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <span className="text-6xl mb-4">🔍</span>
        <p className="text-gray-500 text-lg">{emptyMessage}</p>
      </div>
    )
  }

  return (
    <div className={className}>
      <div className={cn('grid gap-4', gridCols)}>
        {animes.map((anime) => (
          <AnimeCard key={anime.id} anime={anime} size={cardSize} />
        ))}
        {isLoading &&
          Array.from({ length: 6 }).map((_, i) => <SkeletonCard key={`skeleton-${i}`} />)}
      </div>

      {hasMore && onLoadMore && !isLoading && (
        <div className="flex justify-center mt-8">
          <button
            onClick={onLoadMore}
            className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            もっと見る
          </button>
        </div>
      )}
    </div>
  )
}

export default AnimeGrid
