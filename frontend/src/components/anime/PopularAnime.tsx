'use client'

import type React from 'react'
import { animeApi } from '@/lib/api/annict'
import AnimeSection from './AnimeSection'

interface PopularAnimeProps {
  className?: string
}

const PopularAnime: React.FC<PopularAnimeProps> = ({ className = '' }) => {
  return (
    <AnimeSection
      title="人気アニメ"
      queryKey={['popularAnime']}
      queryFn={() => animeApi.getPopularAnimes(20)}
      linkHref="/anime/popular"
      emptyMessage="人気アニメが見つかりませんでした"
      errorMessage="人気アニメの取得に失敗しました"
      displayMode="carousel"
      className={className}
    />
  )
}

export default PopularAnime
