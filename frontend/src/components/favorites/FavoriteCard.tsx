'use client'

import React from 'react'
import Link from 'next/link'
import Image from 'next/image'
import type { Route } from 'next'
import { cn } from '@/lib/utils'
import { formatSeason } from '@/utils/formatSeason'
import { Anime } from '@/lib/api/generated/myanimelogs'
import { withBackParam } from '@/lib/animeListOrigin'

interface FavoriteCardProps {
  anime: Anime & { id: number } // id is required
  onRemove: () => void
  isUpdating?: boolean
}

/**
 * お気に入り（ハート）専用カード。ステータスは持たず、ハートで解除できる。
 *
 * 枠と白背景を持たない構成・寸法・余白は AnimeCard（medium）と共通で、
 * 今期／人気アニメ・検索・マイページと同じ見た目になる。
 * 解除ボタンを重ねるため、カード全体ではなくポスターとタイトルを個別にリンクする。
 */
export const FavoriteCard: React.FC<FavoriteCardProps> = ({
  anime,
  onRemove,
  isUpdating = false,
}) => {
  const imageUrl =
    anime.cover_image_extra_large ||
    anime.cover_image_large ||
    anime.cover_image_medium ||
    anime.image_url ||
    '/placeholder-anime.png'
  const formattedSeason = formatSeason(anime.season)
  const href = withBackParam(`/anime/${anime.id}`, '/favorites') as Route

  return (
    <div className={cn('group', isUpdating && 'opacity-70')}>
      <div className="relative">
        <Link
          href={href}
          prefetch={false}
          className="block rounded-lg focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
        >
          <div
            className={cn(
              'relative aspect-[2/3] bg-muted rounded-lg overflow-hidden',
              'shadow-[0_4px_14px_-6px_rgba(15,23,42,0.18)]',
              'transition-shadow duration-200 group-hover:shadow-[0_8px_20px_-6px_rgba(15,23,42,0.28)]'
            )}
          >
            <Image
              src={imageUrl}
              alt={anime.title || 'アニメ画像'}
              fill
              className="object-cover group-hover:scale-105 transition-transform duration-300"
              sizes="(max-width: 767px) 33vw, (max-width: 1023px) 25vw, 16vw"
            />
            <div className="absolute inset-0 bg-black/0 group-hover:bg-black/10 transition-colors duration-200" />
          </div>
        </Link>

        {/* お気に入り解除ボタン（ハート） */}
        <button
          type="button"
          onClick={onRemove}
          disabled={isUpdating}
          aria-label="お気に入りから削除"
          className={cn(
            'absolute top-1.5 right-1.5 sm:top-2 sm:right-2 flex items-center justify-center',
            'w-7 h-7 sm:w-8 sm:h-8 rounded-full shadow',
            'bg-pink-500 text-white hover:bg-pink-600 transition-colors',
            'focus:outline-none focus-visible:ring-2 focus-visible:ring-pink-400 focus-visible:ring-offset-1',
            'disabled:opacity-50 disabled:cursor-not-allowed'
          )}
        >
          <svg
            className="w-4 h-4 sm:w-[18px] sm:h-[18px]"
            fill="currentColor"
            stroke="currentColor"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"
            />
          </svg>
        </button>
      </div>

      <Link href={href} prefetch={false} className="block">
        <h3
          className={cn(
            'mt-2 sm:mt-2.5 font-semibold text-foreground group-hover:text-primary transition-colors',
            'line-clamp-2 text-[11.5px] sm:text-[13.5px] leading-[1.4]',
            'min-h-[32px] sm:min-h-[38px]'
          )}
        >
          {anime.title}
        </h3>
      </Link>
      {formattedSeason && (
        <p className="text-muted-foreground text-[10.5px] sm:text-xs mt-0.5 sm:mt-1">
          {formattedSeason}
        </p>
      )}
    </div>
  )
}
