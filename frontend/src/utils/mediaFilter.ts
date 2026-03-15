import type { DBAnime } from '@/lib/api/annict'

export type MediaType = 'all' | 'tv' | 'web' | 'movie'

/** メディアタイプでアニメをフィルタリング */
export const filterAnimesByMedia = (
  animes: DBAnime[],
  mediaType: MediaType
): DBAnime[] => {
  if (mediaType === 'all') return animes

  return animes.filter((anime) => {
    const media = anime.media?.toLowerCase()
    switch (mediaType) {
      case 'tv': return media === 'tv'
      case 'web': return media === 'web'
      case 'movie': return media === 'movie'
      default: return true
    }
  })
}

/** メディアタイプ別の件数を集計 */
export const countAnimesByMedia = (animes: DBAnime[]) => {
  const counts = { all: animes.length, tv: 0, web: 0, movie: 0 }

  animes.forEach((anime) => {
    const media = anime.media?.toLowerCase()
    if (media === 'tv') counts.tv++
    else if (media === 'web') counts.web++
    else if (media === 'movie') counts.movie++
  })

  return counts
}
