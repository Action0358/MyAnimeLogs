import { filterAnimesByMedia, countAnimesByMedia } from '../mediaFilter'
import type { DBAnime } from '@/lib/api/annict'

const mockAnimes: DBAnime[] = [
  { id: 1, title: 'TV Anime', media: 'tv' } as DBAnime,
  { id: 2, title: 'Web Anime', media: 'web' } as DBAnime,
  { id: 3, title: 'Movie', media: 'movie' } as DBAnime,
  { id: 4, title: 'TV Anime 2', media: 'tv' } as DBAnime,
  { id: 5, title: 'No Media', media: undefined } as DBAnime,
]

describe('filterAnimesByMedia', () => {
  it('should return all animes for "all" type', () => {
    expect(filterAnimesByMedia(mockAnimes, 'all')).toHaveLength(5)
  })

  it('should filter TV animes', () => {
    const result = filterAnimesByMedia(mockAnimes, 'tv')
    expect(result).toHaveLength(2)
    expect(result.every((a) => a.media === 'tv')).toBe(true)
  })

  it('should filter web animes', () => {
    const result = filterAnimesByMedia(mockAnimes, 'web')
    expect(result).toHaveLength(1)
    expect(result[0].media).toBe('web')
  })

  it('should filter movies', () => {
    const result = filterAnimesByMedia(mockAnimes, 'movie')
    expect(result).toHaveLength(1)
    expect(result[0].media).toBe('movie')
  })

  it('should return empty array when no matches', () => {
    const tvOnly = [{ id: 1, title: 'TV', media: 'tv' } as DBAnime]
    expect(filterAnimesByMedia(tvOnly, 'movie')).toHaveLength(0)
  })
})

describe('countAnimesByMedia', () => {
  it('should count animes by media type', () => {
    const counts = countAnimesByMedia(mockAnimes)
    expect(counts.all).toBe(5)
    expect(counts.tv).toBe(2)
    expect(counts.web).toBe(1)
    expect(counts.movie).toBe(1)
  })

  it('should return all zeros for empty array', () => {
    const counts = countAnimesByMedia([])
    expect(counts.all).toBe(0)
    expect(counts.tv).toBe(0)
    expect(counts.web).toBe(0)
    expect(counts.movie).toBe(0)
  })
})
