'use client'

import React from 'react'
import Link from 'next/link'
import Image from 'next/image'
import type { Route } from 'next'
import { cn } from '@/lib/utils'
import { formatSeason } from '@/utils/formatSeason'
import { Anime } from '@/lib/api/generated/myanimelogs'
import { withBackParam } from '@/lib/animeListOrigin'

interface WatchStatusCardProps {
  anime: Anime & { id: number } // id is required
  isUpdating?: boolean
}

/**
 * 視聴ステータス付きのカード（マイページ用）。
 *
 * 枠と白背景を持たず、ポスターを主役にする構成は AnimeCard（medium）と共通。
 * 一覧・検索ページと縦のリズムを揃えるため、タイトルは2行ぶんの高さを確保する。
 */
export const WatchStatusCard: React.FC<WatchStatusCardProps> = ({
  anime,
  isUpdating = false,
}) => {
  const imageUrl =
    anime.cover_image_extra_large ||
    anime.cover_image_large ||
    anime.cover_image_medium ||
    anime.image_url ||
    '/placeholder-anime.png'
  const formattedSeason = formatSeason(anime.season)

  return (
    <Link
      href={withBackParam(`/anime/${anime.id}`, '/mypage') as Route}
      prefetch={false}
      className={cn(
        'block rounded-lg focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
        isUpdating && 'pointer-events-none'
      )}
      aria-disabled={isUpdating}
      tabIndex={isUpdating ? -1 : undefined}
      onClick={(e) => {
        if (isUpdating) e.preventDefault()
      }}
    >
      <div className={cn('group cursor-pointer', isUpdating && 'opacity-70')}>
        <div
          className={cn(
            'relative aspect-[2/3] bg-muted rounded-lg overflow-hidden',
            'shadow-[0_4px_14px_-6px_rgba(15,23,42,0.18)]',
            'transition-shadow duration-200 group-hover:shadow-[0_8px_20px_-6px_rgba(15,23,42,0.28)]'
          )}
        >
          <Image
            src={imageUrl}
            alt=""
            fill
            className="object-cover group-hover:scale-105 transition-transform duration-300"
            sizes="(max-width: 767px) 33vw, (max-width: 1023px) 25vw, 16vw"
          />
          <div className="absolute inset-0 bg-black/0 group-hover:bg-black/10 transition-colors duration-200" />
        </div>

        <h3
          className={cn(
            'mt-2 sm:mt-2.5 font-semibold text-foreground group-hover:text-primary transition-colors',
            'line-clamp-2 text-[11.5px] sm:text-[13.5px] leading-[1.4]',
            'min-h-[32px] sm:min-h-[38px]'
          )}
        >
          {anime.title || 'タイトル不明'}
        </h3>
        {formattedSeason && (
          <p className="text-muted-foreground text-[10.5px] sm:text-xs mt-0.5 sm:mt-1">
            {formattedSeason}
          </p>
        )}
      </div>
    </Link>
  )
}
