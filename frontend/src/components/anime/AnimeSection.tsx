'use client'

import type React from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import { useQuery } from '@tanstack/react-query'
import { Button } from '@/components/ui'
import { AnimeCard } from './AnimeCard'
import { AnimeCarousel, CarouselSkeleton } from './AnimeCarousel'
import { SkeletonCard } from './SkeletonCard'
import type { DBAnime } from '@/lib/api/annict'
import type { Anime, AnimeStatus } from '@/lib/api/generated/myanimelogs'

interface AnimeSectionProps {
  title: string
  queryKey: string[]
  queryFn: () => Promise<{ animes: DBAnime[]; total: number }>
  linkHref: string
  linkText?: string
  emptyMessage: string
  errorMessage: string
  displayMode?: 'grid' | 'carousel'
  className?: string
}

// DBAnimeをAnimeCard用の型に変換
const convertToAnimeCardFormat = (anime: DBAnime): Anime & { id: number } => {
  let status: AnimeStatus | undefined = undefined
  if (anime.status === 'airing' || anime.status === 'finished') {
    status = anime.status
  }

  return {
    id: anime.id,
    title: anime.title,
    description: undefined,
    image_url: anime.image_url,
    season: anime.season,
    status,
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
  errorMessage,
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

  const SectionHeader = (
    <div className="flex items-center justify-between mb-6">
      <h2 className="text-2xl font-bold text-gray-900">{title}</h2>
      <Button variant="outline-gray" size="sm" asChild className="font-bold">
        <Link href={linkHref as Route}>{linkText}</Link>
      </Button>
    </div>
  )

  if (isLoading) {
    return (
      <div className={className}>
        {SectionHeader}
        {isCarousel ? (
          <CarouselSkeleton />
        ) : (
          <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
            {Array.from({ length: 6 }).map((_, i) => (
              <SkeletonCard key={i} />
            ))}
          </div>
        )}
      </div>
    )
  }

  if (error) {
    return (
      <div className={className}>
        {SectionHeader}
        <div className="text-center py-12">
          <p className="text-red-600 mb-4">
            {error instanceof Error ? error.message : errorMessage}
          </p>
          <Button variant="outline" onClick={() => refetch()}>
            再試行
          </Button>
        </div>
      </div>
    )
  }

  const cardAnimes = animes.map(convertToAnimeCardFormat)

  return (
    <div className={className}>
      {SectionHeader}

      {isCarousel ? (
        <AnimeCarousel animes={cardAnimes} />
      ) : (
        <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
          {cardAnimes.slice(0, 6).map((anime) => (
            <AnimeCard key={anime.id} anime={anime} size="medium" />
          ))}
        </div>
      )}

      {animes.length === 0 && (
        <div className="text-center py-12">
          <p className="text-gray-500">{emptyMessage}</p>
        </div>
      )}
    </div>
  )
}

export default AnimeSection
