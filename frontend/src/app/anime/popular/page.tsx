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
      showMediaTabs
      showGenreSort
      columns={{ mobile: 3, tablet: 3, desktop: 4, wide: 6 }}
    />
  )
}
