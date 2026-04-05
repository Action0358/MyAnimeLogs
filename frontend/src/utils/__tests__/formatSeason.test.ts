import { translateSeason, formatSeason } from '../formatSeason'

describe('translateSeason', () => {
  it('should translate season names to Japanese', () => {
    expect(translateSeason('winter')).toBe('冬')
    expect(translateSeason('spring')).toBe('春')
    expect(translateSeason('summer')).toBe('夏')
    expect(translateSeason('autumn')).toBe('秋')
    expect(translateSeason('fall')).toBe('秋')
  })

  it('should be case insensitive', () => {
    expect(translateSeason('Winter')).toBe('冬')
    expect(translateSeason('SPRING')).toBe('春')
  })

  it('should return original string for unknown season', () => {
    expect(translateSeason('unknown')).toBe('unknown')
  })
})

describe('formatSeason', () => {
  it('should return undefined for null/undefined/empty', () => {
    expect(formatSeason(null)).toBeUndefined()
    expect(formatSeason(undefined)).toBeUndefined()
    expect(formatSeason('')).toBeUndefined()
  })

  it('should format "YYYY-season" pattern', () => {
    expect(formatSeason('2026-winter')).toBe('2026年冬')
    expect(formatSeason('2025-spring')).toBe('2025年春')
    expect(formatSeason('2024-summer')).toBe('2024年夏')
    expect(formatSeason('2023-autumn')).toBe('2023年秋')
  })

  it('should format "YYYY season" pattern', () => {
    expect(formatSeason('2026 winter')).toBe('2026年冬')
  })

  it('should format season name only', () => {
    expect(formatSeason('winter')).toBe('冬')
    expect(formatSeason('spring')).toBe('春')
  })

  it('should return original string for unknown format', () => {
    expect(formatSeason('something')).toBe('something')
  })
})
