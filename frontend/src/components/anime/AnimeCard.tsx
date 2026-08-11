'use client'

import React, { useState, useEffect } from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import type { Route } from 'next'
import { cn } from '@/lib/utils'
import { withBackParam } from '@/lib/animeListOrigin'
import { formatSeason } from '@/utils/formatSeason'
import { getAnimeStatusLabel, type BroadcastStatusLabel } from '@/utils/animeStatus'
import type { Anime } from '@/lib/api/generated/myanimelogs'

/** AnimeCard用の共有型定義 */
export type AnimeCardData = Anime & { id: number }

export interface AnimeCardProps {
  anime: AnimeCardData
  showProgress?: boolean
  showFavoriteStatus?: boolean
  size?: 'small' | 'medium' | 'large'
  onClick?: (anime: AnimeCardData) => void
  className?: string
  /**
   * 詳細ページの「戻る」導線に載せる遷移元パス。
   * 省略時は現在のパスを使う。検索ページのようにクエリごと保持したい場合に明示する。
   */
  from?: string
}

/**
 * カードは枠を持たず、ポスター画像の下にテキストを置く構成。
 * タイトルは2行ぶんの高さを確保し、タイトル長がばらついてもグリッドの縦リズムを揃える。
 */
const sizeConfig = {
  small: {
    aspectRatio: 'aspect-[2/3]',
    titleClass: 'text-[11.5px] leading-[1.4]',
    titleHeight: 'min-h-[32px]',
    seasonClass: 'text-[10.5px] mt-0.5',
  },
  medium: {
    aspectRatio: 'aspect-[2/3]',
    titleClass: 'text-[11.5px] sm:text-[13.5px] leading-[1.4]',
    titleHeight: 'min-h-[32px] sm:min-h-[38px]',
    seasonClass: 'text-[10.5px] sm:text-xs mt-0.5 sm:mt-1',
  },
  large: {
    aspectRatio: 'aspect-[2/3]',
    titleClass: 'text-sm sm:text-base leading-[1.4]',
    titleHeight: 'min-h-[40px] sm:min-h-[46px]',
    seasonClass: 'text-xs mt-1',
  },
}

export const AnimeCard: React.FC<AnimeCardProps> = ({
  anime,
  size = 'medium',
  onClick,
  className,
  from,
}) => {
  const pathname = usePathname()
  const [imageError, setImageError] = useState(false)
  // ハイドレーションミスマッチを回避するため、クライアントサイドでのみ放送状態を計算
  const [statusInfo, setStatusInfo] = useState<BroadcastStatusLabel>({ label: null, className: '' })
  const config = sizeConfig[size]

  useEffect(() => {
    // クライアントサイドでのみ実行（new Date()を使用するため）
    setStatusInfo(getAnimeStatusLabel(anime.status, anime.released_on))
  }, [anime.status, anime.released_on])

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
      className={cn('group cursor-pointer', className)}
      onClick={onClick ? handleClick : undefined}
    >
      <div
        className={cn(
          'relative bg-muted rounded-lg overflow-hidden',
          'shadow-[0_4px_14px_-6px_rgba(15,23,42,0.18)]',
          'transition-shadow duration-200 group-hover:shadow-[0_8px_20px_-6px_rgba(15,23,42,0.28)]',
          config.aspectRatio
        )}
      >
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
          <div className="w-full h-full flex items-center justify-center bg-gradient-to-br from-primary/10 to-primary/20">
            <span className="text-4xl">📺</span>
          </div>
        )}
        <div className="absolute inset-0 bg-black/0 group-hover:bg-black/10 transition-colors duration-200" />

        {statusInfo.label && (
          <div className="absolute top-1.5 left-1.5 sm:top-2 sm:left-2">
            <span
              className={cn(
                'inline-flex items-center px-2 py-0.5 rounded-full text-[9.5px] sm:text-[11px] font-bold',
                statusInfo.className
              )}
            >
              {statusInfo.label}
            </span>
          </div>
        )}
      </div>

      <h3
        className={cn(
          'mt-2 sm:mt-2.5 font-semibold text-foreground group-hover:text-primary transition-colors',
          'line-clamp-2',
          config.titleClass,
          config.titleHeight
        )}
      >
        {anime.title}
      </h3>
      {anime.season && (
        <p className={cn('text-muted-foreground', config.seasonClass)}>
          {formatSeason(anime.season)}
        </p>
      )}
    </div>
  )

  if (onClick) {
    return CardContent
  }

  return (
    <Link
      href={withBackParam(`/anime/${anime.id}`, from ?? pathname ?? undefined) as Route}
      prefetch={false}
      className="block rounded-lg focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
    >
      {CardContent}
    </Link>
  )
}

export default AnimeCard
