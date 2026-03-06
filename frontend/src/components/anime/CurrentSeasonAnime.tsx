'use client'

import type React from 'react'
import { animeApi } from '@/lib/api/annict'
import AnimeSection from './AnimeSection'

interface CurrentSeasonAnimeProps {
  className?: string
}

const CurrentSeasonAnime: React.FC<CurrentSeasonAnimeProps> = ({
  className = '',
}) => {
  return (
    <AnimeSection
      title="今期アニメ"
      queryKey={['currentSeasonAnime']}
      queryFn={() => animeApi.getCurrentSeasonAnimes(20)}
      linkHref="/anime/current"
      emptyMessage="今期アニメが見つかりませんでした"
      errorMessage="今期アニメの取得に失敗しました"
      displayMode="carousel"
      className={className}
    />
  )
}

export default CurrentSeasonAnime
