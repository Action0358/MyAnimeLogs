import { filterAnimesByGenres } from '../genreFilter'
import type { DBAnime } from '@/lib/api/annict'

const mockAnimes: DBAnime[] = [
  { id: 1, title: 'A', genres: ['アクション', 'コメディ'] } as DBAnime,
  { id: 2, title: 'B', genres: ['ドラマ'] } as DBAnime,
  { id: 3, title: 'C', genres: undefined } as DBAnime,
]

describe('filterAnimesByGenres', () => {
  it('should return all animes when no genre is selected', () => {
    expect(filterAnimesByGenres(mockAnimes, [])).toHaveLength(3)
  })

  it('should keep animes matching any selected genre (OR)', () => {
    const result = filterAnimesByGenres(mockAnimes, ['コメディ', 'ドラマ'])
    expect(result.map((a) => a.id)).toEqual([1, 2])
  })

  it('should exclude animes without genres data', () => {
    const result = filterAnimesByGenres(mockAnimes, ['アクション'])
    expect(result.map((a) => a.id)).toEqual([1])
  })
})
