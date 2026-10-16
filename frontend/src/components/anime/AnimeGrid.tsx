'use client'

import React from 'react'
import { SearchX } from 'lucide-react'
import { cn } from '@/lib/utils'
import { AnimeCard } from './AnimeCard'
import { SkeletonCard } from './SkeletonCard'
import { EmptyState } from '@/components/ui/EmptyState'
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
  /** 詳細ページの「戻る」導線に載せる遷移元パス。省略時はカードが現在のパスを使う */
  from?: string
}

export const AnimeGrid: React.FC<AnimeGridProps> = ({
  animes,
  isLoading = false,
  hasMore = false,
  onLoadMore,
  emptyMessage = 'アニメが見つかりませんでした',
  columns = { mobile: 3, tablet: 3, desktop: 4, wide: 6 },
  cardSize = 'medium',
  className,
  from,
}) => {
  const gridCols = cn(
    columns.mobile === 1 && 'grid-cols-1',
    columns.mobile === 2 && 'grid-cols-2',
    columns.mobile === 3 && 'grid-cols-3',
    columns.tablet === 3 && 'sm:grid-cols-3',
    columns.tablet === 4 && 'sm:grid-cols-4',
    columns.desktop === 4 && 'md:grid-cols-4',
    columns.desktop === 5 && 'md:grid-cols-5',
    columns.wide === 4 && 'lg:grid-cols-4',
    columns.wide === 5 && 'lg:grid-cols-5',
    columns.wide === 6 && 'lg:grid-cols-6'
  )

  if (isLoading && animes.length === 0) {
    return (
      <div className={className}>
        <div className={cn('grid gap-x-2.5 gap-y-3.5 md:gap-x-[18px] md:gap-y-5', gridCols)}>
          {Array.from({ length: 12 }).map((_, i) => (
            <SkeletonCard key={i} />
          ))}
        </div>
      </div>
    )
  }

  if (!isLoading && animes.length === 0) {
    return (
      <EmptyState
        title={emptyMessage}
        icon={<SearchX className="w-7 h-7 md:w-8 md:h-8" strokeWidth={1.75} aria-hidden="true" />}
      />
    )
  }

  return (
    <div className={className}>
      <div className={cn('grid gap-x-2.5 gap-y-3.5 md:gap-x-[18px] md:gap-y-5', gridCols)}>
        {animes.map((anime) => (
          <AnimeCard key={anime.id} anime={anime} size={cardSize} from={from} />
        ))}
        {isLoading &&
          Array.from({ length: 6 }).map((_, i) => <SkeletonCard key={`skeleton-${i}`} />)}
      </div>

      {hasMore && onLoadMore && !isLoading && (
        <div className="flex justify-center mt-8">
          <button
            onClick={onLoadMore}
            className="px-6 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors"
          >
            もっと見る
          </button>
        </div>
      )}
    </div>
  )
}

export default AnimeGrid
