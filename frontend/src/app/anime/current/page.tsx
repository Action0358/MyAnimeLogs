'use client'

import { AnimeListPage } from '@/components/anime'
import { animeApi } from '@/lib/api/annict'
import { CURRENT_SEASON_ANIME_LIMIT } from '@/lib/constants'

export default function CurrentSeasonPage() {
  return (
    <AnimeListPage
      title="今期アニメ"
      queryKey={['currentSeasonAnimes']}
      queryFn={() => animeApi.getCurrentSeasonAnimes(CURRENT_SEASON_ANIME_LIMIT)}
      emptyMessage="今期アニメが見つかりませんでした"
      errorMessage="今期アニメの取得に失敗しました"
    />
  )
}
