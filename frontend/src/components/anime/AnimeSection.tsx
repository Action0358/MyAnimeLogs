'use client'

import type React from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import { useQuery } from '@tanstack/react-query'
import { Button } from '@/components/ui'
import { AnimeCard } from './AnimeCard'
import { formatSeason } from '@/utils/formatSeason'
import type { AnnictWork } from '@/lib/api/annict'

interface AnimeSectionProps {
  title: string
  queryKey: string[]
  queryFn: () => Promise<{ works: AnnictWork[] }>
  linkHref: string
  linkText?: string
  emptyMessage: string
  errorMessage: string
  className?: string
}

// AnnictWorkをAnimeCard用の型に変換
const convertToAnimeCardFormat = (work: AnnictWork) => {
  let season: string | undefined = undefined
  
  if (work.season_year && work.season_name) {
    season = `${work.season_year}年${work.season_name}`
  } else if (work.season_year) {
    season = `${work.season_year}年`
  } else if (work.season_name) {
    season = work.season_name
  }
  
  return {
    id: work.id,
    title: work.title,
    description: undefined,
    image_url: work.images?.twitter?.bigger_avatar_url || work.images?.twitter?.original_avatar_url,
    season: formatSeason(season),
    status: undefined,
    cover_image_large: work.cover_image_large,
    cover_image_medium: work.cover_image_medium,
  }
}

// スケルトンカード
const SkeletonCard: React.FC = () => (
  <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden animate-pulse">
    <div className="aspect-[3/4] bg-gray-200" />
    <div className="p-3 h-[100px] flex flex-col">
      <div className="h-4 bg-gray-200 rounded w-full mb-1" />
      <div className="h-4 bg-gray-200 rounded w-3/4 mb-2" />
      <div className="h-3 bg-gray-200 rounded w-1/2 mt-auto" />
    </div>
  </div>
)

const AnimeSection: React.FC<AnimeSectionProps> = ({
  title,
  queryKey,
  queryFn,
  linkHref,
  linkText = 'もっと見る',
  emptyMessage,
  errorMessage,
  className = '',
}) => {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey,
    queryFn,
    staleTime: 1000 * 60 * 5, // 5分間キャッシュを有効
  })

  const animes = data?.works ?? []

  if (isLoading) {
    return (
      <div className={className}>
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-2xl font-bold text-gray-900">{title}</h2>
        </div>
        <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
          {Array.from({ length: 6 }).map((_, i) => (
            <SkeletonCard key={i} />
          ))}
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className={className}>
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-2xl font-bold text-gray-900">{title}</h2>
        </div>
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

  return (
    <div className={className}>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold text-gray-900">{title}</h2>
        <Link
          href={linkHref as Route}
          className="inline-flex items-center justify-center h-8 px-3 text-sm font-medium rounded-md border border-gray-300 bg-transparent text-gray-700 hover:bg-gray-50 active:bg-gray-100 transition-colors"
        >
          {linkText}
        </Link>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
        {animes.slice(0, 6).map((anime) => (
          <AnimeCard
            key={anime.id}
            anime={convertToAnimeCardFormat(anime)}
            size="medium"
          />
        ))}
      </div>

      {animes.length === 0 && (
        <div className="text-center py-12">
          <p className="text-gray-500">{emptyMessage}</p>
        </div>
      )}
    </div>
  )
}

export default AnimeSection
