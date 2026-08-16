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
}

/**
 * お気に入り一覧のカード。
 *
 * 枠と白背景を持たない構成・寸法・余白は AnimeCard（medium）と共通で、
 * 今期／人気アニメ・検索・マイページと同じ見た目になる。
 * 解除は視聴ステータスと同様に詳細ページで行うため、カード上に操作は持たない。
 */
export const FavoriteCard: React.FC<FavoriteCardProps> = ({ anime }) => {
  const imageUrl =
    anime.cover_image_extra_large ||
    anime.cover_image_large ||
    anime.cover_image_medium ||
    anime.image_url ||
    '/placeholder-anime.png'
  const formattedSeason = formatSeason(anime.season)

  return (
    <Link
      href={withBackParam(`/anime/${anime.id}`, '/favorites') as Route}
      prefetch={false}
      className="block rounded-lg focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
    >
      <div className="group cursor-pointer">
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

        <h3
          className={cn(
            'mt-2 sm:mt-2.5 font-semibold text-foreground group-hover:text-primary transition-colors',
            'line-clamp-2 text-[11.5px] sm:text-[13.5px] leading-[1.4]',
            'min-h-[32px] sm:min-h-[38px]'
          )}
        >
          {anime.title}
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
