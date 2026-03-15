'use client'

import { AnimeListPage } from '@/components/anime'
import { animeApi } from '@/lib/api/annict'
import { POPULAR_ANIME_LIMIT } from '@/lib/constants'

export default function PopularAnimePage() {
  return (
    <AnimeListPage
      title="人気アニメ"
      queryKey={['popularAnimes']}
      queryFn={() => animeApi.getPopularAnimes(POPULAR_ANIME_LIMIT)}
      emptyMessage="人気アニメが見つかりませんでした"
      errorMessage="人気アニメの取得に失敗しました"
      showMediaTabs
    />
  )
}
