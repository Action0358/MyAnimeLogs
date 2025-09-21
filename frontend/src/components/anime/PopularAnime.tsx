'use client'

import type React from 'react'
import { annictApi } from '@/lib/api/annict'
import AnimeSection from './AnimeSection'

interface PopularAnimeProps {
  className?: string
}

const PopularAnime: React.FC<PopularAnimeProps> = ({ className = '' }) => {
  return (
    <AnimeSection
      title="人気アニメ"
      queryKey={['popularAnime']}
      queryFn={annictApi.getPopularAnime}
      linkHref="/anime?sort=popular"
      emptyMessage="人気アニメが見つかりませんでした"
      errorMessage="人気アニメの取得に失敗しました"
      icon="🎬"
      gradientFrom="from-green-100"
      gradientTo="to-green-300"
      className={className}
    />
  )
}

export default PopularAnime
