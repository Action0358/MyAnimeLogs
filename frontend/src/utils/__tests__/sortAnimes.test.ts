import { sortAnimes } from '../sortAnimes'
import type { DBAnime } from '@/lib/api/annict'

describe('sortAnimes', () => {
  it('should sort by released_on ascending for "year"', () => {
    const animes = [
      { id: 1, title: 'A', released_on: '2026-03-01' },
      { id: 2, title: 'B', released_on: '2026-01-01' },
      { id: 3, title: 'C', released_on: '2026-02-01' },
    ] as DBAnime[]

    const result = sortAnimes(animes, 'year')
    expect(result.map((a) => a.id)).toEqual([2, 3, 1])
  })

  it('should put animes without released_on last', () => {
    const animes = [
      { id: 1, title: 'A', released_on: undefined },
      { id: 2, title: 'B', released_on: '2026-01-01' },
    ] as DBAnime[]

    const result = sortAnimes(animes, 'year')
    expect(result.map((a) => a.id)).toEqual([2, 1])
  })

  it('should not reorder for "popularity" (API already sorts by watchers_count)', () => {
    const animes = [
      { id: 1, title: 'A', watchers_count: 10 },
      { id: 2, title: 'B', watchers_count: 50 },
    ] as DBAnime[]

    const result = sortAnimes(animes, 'popularity')
    expect(result.map((a) => a.id)).toEqual([1, 2])
  })
})
