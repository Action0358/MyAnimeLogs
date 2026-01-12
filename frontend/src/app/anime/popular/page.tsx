'use client'

import { AnimeListPage } from '@/components/anime'
import { animeApi } from '@/lib/api/annict'

export default function PopularAnimePage() {
  return (
    <AnimeListPage
      title="人気アニメ"
      queryKey={['popularAnimes']}
      queryFn={() => animeApi.getPopularAnimes(100)}
      emptyMessage="人気アニメが見つかりませんでした"
      errorMessage="人気アニメの取得に失敗しました"
      totalLabel={(total) => `人気Top${total}`}
    />
  )
}
