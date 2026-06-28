'use client'

import React from 'react'
import Link from 'next/link'
import Image from 'next/image'
import type { Route } from 'next'
import { cn } from '@/lib/utils'
import { formatSeason } from '@/utils/formatSeason'
import { WatchStatusStatus, Anime } from '@/lib/api/generated/myanimelogs'
import { STATUS_CONFIG } from './StatusSelector'

interface WatchStatusCardProps {
  anime: Anime & { id: number } // id is required
  status: WatchStatusStatus
  isUpdating?: boolean
}

// 視聴ステータス付きのカード（マイページ用）
// デザインは今期/人気アニメの AnimeCard（medium）に合わせる
export const WatchStatusCard: React.FC<WatchStatusCardProps> = ({
  anime,
  status,
  isUpdating = false,
}) => {
  const imageUrl = anime.cover_image_extra_large || anime.cover_image_large || anime.cover_image_medium || anime.image_url || '/placeholder-anime.png'
  const formattedSeason = formatSeason(anime.season)

  return (
    <Link
      href={`/anime/${anime.id}` as Route}
      prefetch={false}
      className={cn('block', isUpdating && 'pointer-events-none')}
      aria-disabled={isUpdating}
      tabIndex={isUpdating ? -1 : undefined}
      onClick={(e) => {
        if (isUpdating) e.preventDefault()
      }}
    >
      <div className={cn(
        'group relative bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden',
        'hover:shadow-lg hover:border-gray-300 transition-all duration-200',
        'cursor-pointer',
        isUpdating && 'opacity-70'
      )}>
        <div className="relative aspect-[2/3] bg-gray-100">
          <Image
            src={imageUrl}
            alt=""
            fill
            className="object-cover group-hover:scale-105 transition-transform duration-300"
            sizes="(max-width: 767px) 33vw, (max-width: 1023px) 25vw, 16vw"
          />
          <div className="absolute inset-0 bg-black/0 group-hover:bg-black/10 transition-colors duration-200" />

          <div className="absolute top-2 left-2">
            <span className={cn(
              'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium text-white',
              STATUS_CONFIG[status].color
            )}>
              {STATUS_CONFIG[status].label}
            </span>
          </div>
        </div>

        <div className="p-2 sm:p-3 h-[72px] sm:h-[100px] flex flex-col">
          <h3 className="font-medium text-gray-900 group-hover:text-blue-600 transition-colors line-clamp-2 text-xs sm:text-sm">
            {anime.title || 'タイトル不明'}
          </h3>
          {formattedSeason && (
            <p className="text-xs text-gray-500 mt-auto">{formattedSeason}</p>
          )}
        </div>
      </div>
    </Link>
  )
}
