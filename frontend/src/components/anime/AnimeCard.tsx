'use client'

import React, { useState } from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import { cn } from '@/lib/utils'
import { formatSeason } from '@/utils/formatSeason'
import { getAnimeStatusLabel } from '@/utils/animeStatus'
import type { Anime } from '@/lib/api/generated/myanimelogs'

export interface AnimeCardProps {
  anime: Anime & { id: number }
  showProgress?: boolean
  showFavoriteStatus?: boolean
  size?: 'small' | 'medium' | 'large'
  onClick?: (anime: Anime & { id: number }) => void
  className?: string
}

const sizeConfig = {
  small: {
    aspectRatio: 'aspect-[3/4]',
    titleClass: 'text-xs',
    padding: 'p-2',
    contentHeight: 'h-[80px]',
  },
  medium: {
    aspectRatio: 'aspect-[3/4]',
    titleClass: 'text-sm',
    padding: 'p-3',
    contentHeight: 'h-[100px]',
  },
  large: {
    aspectRatio: 'aspect-[3/4]',
    titleClass: 'text-base',
    padding: 'p-4',
    contentHeight: 'h-[116px]',
  },
}

export const AnimeCard: React.FC<AnimeCardProps> = ({
  anime,
  size = 'medium',
  onClick,
  className,
}) => {
  const [imageError, setImageError] = useState(false)
  const config = sizeConfig[size]

  const imageUrl =
    anime.cover_image_extra_large ||
    anime.cover_image_large ||
    anime.cover_image_medium ||
    anime.image_url

  const handleClick = (e: React.MouseEvent) => {
    if (onClick) {
      e.preventDefault()
      onClick(anime)
    }
  }

  const CardContent = (
    <div
      className={cn(
        'group relative bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden',
        'hover:shadow-lg hover:border-gray-300 transition-all duration-200',
        'cursor-pointer',
        className
      )}
      onClick={onClick ? handleClick : undefined}
    >
      <div className={cn('relative bg-gray-100', config.aspectRatio)}>
        {imageUrl && !imageError ? (
          // 外部画像URLは様々なドメインから取得されるため、next/imageではなくimgタグを使用
          // next.config.jsにすべてのドメインを登録することは現実的ではない
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={imageUrl}
            alt={anime.title || 'アニメ画像'}
            className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
            onError={() => setImageError(true)}
            loading="lazy"
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center bg-gradient-to-br from-blue-100 to-purple-100">
            <span className="text-4xl">📺</span>
          </div>
        )}
        <div className="absolute inset-0 bg-black/0 group-hover:bg-black/10 transition-colors duration-200" />
        
        {anime.status && (
          <div className="absolute top-2 left-2">
            {(() => {
              const { label, className } = getAnimeStatusLabel(anime.status, anime.released_on)
              if (!label) return null
              return (
                <span
                  className={cn(
                    'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium',
                    className
                  )}
                >
                  {label}
                </span>
              )
            })()}
          </div>
        )}
      </div>

      <div className={cn(config.padding, config.contentHeight, 'flex flex-col')}>
        <h3
          className={cn(
            'font-medium text-gray-900 group-hover:text-blue-600 transition-colors',
            'line-clamp-2',
            config.titleClass
          )}
        >
          {anime.title}
        </h3>
        {anime.season && (
          <p className="text-xs text-gray-500 mt-auto">{formatSeason(anime.season)}</p>
        )}
      </div>
    </div>
  )

  if (onClick) {
    return CardContent
  }

  return (
    <Link href={`/anime/${anime.id}` as Route} className="block">
      {CardContent}
    </Link>
  )
}

export default AnimeCard
