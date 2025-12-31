'use client'

import type React from 'react'
import { annictApi } from '@/lib/api/annict'
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
      queryFn={annictApi.getCurrentSeasonAnime}
      linkHref="/anime?season=current"
      emptyMessage="今期アニメが見つかりませんでした"
      errorMessage="今期アニメの取得に失敗しました"
      icon="📺"
      gradientFrom="from-blue-100"
      gradientTo="to-purple-100"
      className={className}
    />
  )
}

export default CurrentSeasonAnime
