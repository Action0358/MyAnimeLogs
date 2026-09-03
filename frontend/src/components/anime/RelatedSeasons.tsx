'use client'

import React, { useState } from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import { formatSeason } from '@/utils/formatSeason'
import { Card, CardContent, SectionHeading } from '@/components/ui'
import type { RelatedAnime } from '@/lib/api/generated/myanimelogs'

interface RelatedSeasonsProps {
  items: RelatedAnime[]
}

const RelatedSeasonCard: React.FC<{ anime: RelatedAnime }> = ({ anime }) => {
  const [imageError, setImageError] = useState(false)
  const imageUrl = anime.cover_image_medium || anime.image_url

  return (
    <Link href={`/anime/${anime.id}` as Route} className="group w-[104px] flex-shrink-0 sm:w-[128px]">
      <div className="relative aspect-[2/3] overflow-hidden rounded-lg bg-muted shadow-[0_4px_14px_-6px_rgba(15,23,42,0.18)] transition-shadow duration-200 group-hover:shadow-[0_8px_20px_-6px_rgba(15,23,42,0.28)]">
        {imageUrl && !imageError ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={imageUrl}
            alt={anime.title}
            className="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
            onError={() => setImageError(true)}
            loading="lazy"
          />
        ) : (
          <div className="flex h-full w-full items-center justify-center bg-gradient-to-br from-primary/10 to-primary/20">
            <span className="text-3xl">📺</span>
          </div>
        )}
      </div>
      <h4 className="mt-2 line-clamp-2 min-h-[32px] text-[11.5px] font-semibold leading-[1.4] text-foreground transition-colors group-hover:text-primary sm:text-[13.5px]">
        {anime.title}
      </h4>
      {anime.season && (
        <p className="mt-0.5 text-[10.5px] text-muted-foreground sm:text-xs">{formatSeason(anime.season)}</p>
      )}
    </Link>
  )
}

/**
 * 詳細ページに表示する「同シリーズの他シーズン」セクション。
 * バックエンドが series_title の一致で放送時期順に返した一覧をそのまま並べる。
 */
export const RelatedSeasons: React.FC<RelatedSeasonsProps> = ({ items }) => {
  if (!items || items.length === 0) {
    return null
  }

  return (
    <Card>
      <CardContent className="p-5 md:px-6 md:py-5">
        <SectionHeading as="h2" size="sm">
          同シリーズの他シーズン
        </SectionHeading>
        <div className="flex gap-3 overflow-x-auto pb-1 md:gap-4">
          {items.map((anime) => (
            <RelatedSeasonCard key={anime.id} anime={anime} />
          ))}
        </div>
      </CardContent>
    </Card>
  )
}

export default RelatedSeasons
