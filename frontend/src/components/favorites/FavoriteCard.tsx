'use client'

import React from 'react'
import Link from 'next/link'
import Image from 'next/image'
import type { Route } from 'next'
import { cn } from '@/lib/utils'
import { Anime } from '@/lib/api/generated/myanimelogs'

interface FavoriteCardProps {
  anime: Anime & { id: number } // id is required
  onRemove: () => void
  isUpdating?: boolean
}

// お気に入り（ハート）専用カード。ステータスは持たず、ハートで解除できる。
export const FavoriteCard: React.FC<FavoriteCardProps> = ({
  anime,
  onRemove,
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
            sizes="(max-width: 640px) 50vw, (max-width: 1024px) 33vw, 25vw"
          />
        </div>
      </Link>

      {/* お気に入り解除ボタン（ハート） */}
      <button
        type="button"
        onClick={onRemove}
        disabled={isUpdating}
        aria-label="お気に入りから削除"
        className={cn(
          'absolute top-2 right-2 flex items-center justify-center w-9 h-9 rounded-full shadow',
          'bg-pink-500 text-white hover:bg-pink-600 transition-colors',
          'focus:outline-none focus:ring-2 focus:ring-pink-400 focus:ring-offset-1',
          'disabled:opacity-50 disabled:cursor-not-allowed'
        )}
      >
        <svg className="w-5 h-5" fill="currentColor" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
        </svg>
      </button>

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
