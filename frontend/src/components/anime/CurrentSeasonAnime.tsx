'use client'

import type React from 'react'
import { currentSeasonAnimeQuery } from '@/lib/queries/currentSeasonAnime'
import AnimeSection from './AnimeSection'

interface CurrentSeasonAnimeProps {
  className?: string
}

const CurrentSeasonAnime: React.FC<CurrentSeasonAnimeProps> = ({ className = '' }) => {
  return (
    <AnimeSection
      title="今期アニメ"
      queryKey={[...currentSeasonAnimeQuery.queryKey]}
      queryFn={currentSeasonAnimeQuery.queryFn}
      linkHref="/anime/current"
      emptyMessage="今期アニメが見つかりませんでした"
      displayMode="carousel"
      className={className}
    />
  )
}

export default CurrentSeasonAnime
