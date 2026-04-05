import { isValidAnimeStatus, toAnimeStatus } from '../constants'
import type { AnimeStatus } from '../constants'

describe('isValidAnimeStatus', () => {
  it('should return true for valid statuses', () => {
    expect(isValidAnimeStatus('airing')).toBe(true)
    expect(isValidAnimeStatus('finished')).toBe(true)
    expect(isValidAnimeStatus('upcoming')).toBe(true)
  })

  it('should return false for invalid statuses', () => {
    expect(isValidAnimeStatus('currently_airing')).toBe(false)
    expect(isValidAnimeStatus('not_yet_aired')).toBe(false)
    expect(isValidAnimeStatus('')).toBe(false)
    expect(isValidAnimeStatus(null)).toBe(false)
    expect(isValidAnimeStatus(undefined)).toBe(false)
    expect(isValidAnimeStatus(123)).toBe(false)
  })
})

describe('toAnimeStatus', () => {
  it('should return AnimeStatus for valid values', () => {
    expect(toAnimeStatus('airing')).toBe('airing')
    expect(toAnimeStatus('finished')).toBe('finished')
    expect(toAnimeStatus('upcoming')).toBe('upcoming')
  })

  it('should return undefined for invalid values', () => {
    expect(toAnimeStatus('currently_airing')).toBeUndefined()
    expect(toAnimeStatus('invalid')).toBeUndefined()
    expect(toAnimeStatus(null)).toBeUndefined()
    expect(toAnimeStatus(undefined)).toBeUndefined()
  })
})
