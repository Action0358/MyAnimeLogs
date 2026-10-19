'use client'

import type React from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import { useQuery } from '@tanstack/react-query'
import { Button, ErrorState, SectionHeading, getErrorStateType } from '@/components/ui'
import { AnimeCard, type AnimeCardData } from './AnimeCard'
import { AnimeCarousel, CarouselSkeleton } from './AnimeCarousel'
import { SkeletonCard } from './SkeletonCard'
import type { DBAnime } from '@/lib/api/annict'
import type { AnimeStatus } from '@/lib/api/generated/myanimelogs'

/** グリッド表示の定数 */
const GRID_ITEM_COUNT = 6

interface AnimeSectionProps {
  title: string
  queryKey: string[]
  queryFn: () => Promise<{ animes: DBAnime[]; total: number }>
  linkHref: string
  linkText?: string
  emptyMessage: string
  displayMode?: 'grid' | 'carousel'
  className?: string
}

// DBAnimeをAnimeCard用の型に変換
const convertToAnimeCardFormat = (anime: DBAnime): AnimeCardData => {
  let status: AnimeStatus | undefined = undefined
  if (anime.status === 'airing' || anime.status === 'finished' || anime.status === 'upcoming') {
    status = anime.status
  }

  return {
    id: anime.id,
    title: anime.title,
    description: undefined,
    image_url: anime.image_url,
    season: anime.season,
    status,
    released_on: anime.released_on,
    cover_image_large: anime.cover_image_large,
    cover_image_medium: anime.cover_image_medium,
  }
}

const AnimeSection: React.FC<AnimeSectionProps> = ({
  title,
  queryKey,
  queryFn,
  linkHref,
  linkText = 'もっと見る',
  emptyMessage,
  displayMode = 'grid',
  className = '',
}) => {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey,
    queryFn,
    staleTime: 1000 * 60 * 15,
    gcTime: 1000 * 60 * 30,
  })

  const animes = data?.animes ?? []
  const isCarousel = displayMode === 'carousel'

  // グリッド表示の場合は先にsliceしてパフォーマンス改善
  const displayAnimes = isCarousel ? animes : animes.slice(0, GRID_ITEM_COUNT)
  const cardAnimes = displayAnimes.map(convertToAnimeCardFormat)

  const SectionHeader = (
    <SectionHeading
      action={
        <Button variant="outline-gray" size="sm" asChild className="font-bold">
          <Link href={linkHref as Route}>{linkText}</Link>
        </Button>
      }
    >
      {title}
    </SectionHeading>
  )

  if (isLoading) {
    return (
      <div className={className}>
        {SectionHeader}
        {isCarousel ? (
          <CarouselSkeleton />
        ) : (
          <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
            {Array.from({ length: GRID_ITEM_COUNT }).map((_, i) => (
              <SkeletonCard key={i} />
            ))}
          </div>
        )}
      </div>
    )
  }

  if (error) {
    // 失敗した一覧への「もっと見る」は同じエラーになるため出さない。
    // 再読み込みはこのセクションのAPIだけを再取得する
    return (
      <div className={className}>
        <SectionHeading>{title}</SectionHeading>
        <ErrorState variant="compact" type={getErrorStateType(error)} onRetry={() => refetch()} />
      </div>
    )
  }

  return (
    <div className={className}>
      {SectionHeader}

      {isCarousel ? (
        <AnimeCarousel animes={cardAnimes} />
      ) : (
        <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
          {cardAnimes.map((anime) => (
            <AnimeCard key={anime.id} anime={anime} size="medium" />
          ))}
        </div>
      )}

      {animes.length === 0 && (
        <div className="text-center py-12">
          <p className="text-muted-foreground">{emptyMessage}</p>
        </div>
      )}
    </div>
  )
}

export default AnimeSection
