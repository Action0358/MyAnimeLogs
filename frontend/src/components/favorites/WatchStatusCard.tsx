'use client'

import React from 'react'
import Link from 'next/link'
import Image from 'next/image'
import type { Route } from 'next'
import { cn } from '@/lib/utils'
import { WatchStatusStatus, Anime } from '@/lib/api/generated/myanimelogs'
import { STATUS_CONFIG } from './StatusSelector'

interface WatchStatusCardProps {
  anime: Anime & { id: number } // id is required
  status: WatchStatusStatus
  isUpdating?: boolean
}

// 視聴ステータス付きのカード（マイページ用）
export const WatchStatusCard: React.FC<WatchStatusCardProps> = ({
  anime,
  status,
  isUpdating = false,
}) => {
  const imageUrl = anime.cover_image_extra_large || anime.cover_image_large || anime.cover_image_medium || anime.image_url || '/placeholder-anime.png'

  return (
    <div className={cn(
      'group relative bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden',
      'hover:shadow-md transition-shadow',
      isUpdating && 'opacity-70'
    )}>
      <Link href={`/anime/${anime.id}` as Route} className="block">
        <div className="relative aspect-[3/4] bg-gray-100">
          <Image
            src={imageUrl}
            alt={anime.title || 'アニメ画像'}
            fill
            className="object-cover"
            sizes="(max-width: 767px) 33vw, (max-width: 1023px) 25vw, 16vw"
          />
          <div className="absolute top-2 left-2">
            <span className={cn(
              'inline-flex items-center px-2 py-1 rounded-full text-xs font-medium text-white',
              STATUS_CONFIG[status].color
            )}>
              {STATUS_CONFIG[status].label}
            </span>
          </div>
        </div>
      </Link>

      <div className="p-3">
        <Link href={`/anime/${anime.id}` as Route}>
          <h3 className="font-medium text-gray-900 line-clamp-2 hover:text-blue-600 transition-colors">
            {anime.title}
          </h3>
        </Link>

        {anime.season && (
          <p className="text-sm text-gray-500 mt-1">{anime.season}</p>
        )}
      </div>
    </div>
  )
}
