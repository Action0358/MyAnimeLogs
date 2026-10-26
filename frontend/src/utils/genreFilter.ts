import type { DBAnime } from '@/lib/api/annict'

/**
 * 選択中ジャンルのいずれかを持つ作品のみを残す（OR）。
 * ANDにすると該当が激減し複数選択の体験が悪くなるため採らない（検索ページと同じ方針）。
 */
export const filterAnimesByGenres = (animes: DBAnime[], selectedGenres: string[]): DBAnime[] => {
  if (selectedGenres.length === 0) return animes

  return animes.filter((anime) => anime.genres?.some((genre) => selectedGenres.includes(genre)))
}
