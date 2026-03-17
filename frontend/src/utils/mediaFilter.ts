import type { DBAnime } from '@/lib/api/annict'

export type MediaType = 'all' | 'tv' | 'web' | 'movie'

/** メディアタイプでアニメをフィルタリング */
export const filterAnimesByMedia = (
  animes: DBAnime[],
  mediaType: MediaType
): DBAnime[] => {
  if (mediaType === 'all') return animes

  return animes.filter((anime) => anime.media?.toLowerCase() === mediaType)
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
