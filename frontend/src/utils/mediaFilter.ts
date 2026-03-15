import type { DBAnime } from '@/lib/api/annict'

export type MediaType = 'all' | 'tv' | 'web' | 'movie'

/** メディアタイプとDBのmedia値のマッピング */
const mediaValueMap: Record<Exclude<MediaType, 'all'>, string> = {
  tv: 'tv',
  web: 'web',
  movie: 'movie',
}

/** メディアタイプでアニメをフィルタリング */
export const filterAnimesByMedia = (
  animes: DBAnime[],
  mediaType: MediaType
): DBAnime[] => {
  if (mediaType === 'all') return animes

  const targetMedia = mediaValueMap[mediaType]
  return animes.filter((anime) => anime.media?.toLowerCase() === targetMedia)
}

/** メディアタイプ別の件数を集計 */
export const countAnimesByMedia = (animes: DBAnime[]) => {
  const counts = { all: animes.length, tv: 0, web: 0, movie: 0 }

  animes.forEach((anime) => {
    switch (anime.media?.toLowerCase()) {
      case 'tv': counts.tv++; break
      case 'web': counts.web++; break
      case 'movie': counts.movie++; break
    }
  })

  return counts
}
