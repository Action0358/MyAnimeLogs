import { formatDuration, formatEpisodeDuration } from '../duration'

describe('formatDuration', () => {
  it('should return null for null/undefined/0', () => {
    expect(formatDuration(null)).toBeNull()
    expect(formatDuration(undefined)).toBeNull()
    expect(formatDuration(0)).toBeNull()
    expect(formatDuration(-1)).toBeNull()
  })

  it('should format minutes only', () => {
    expect(formatDuration(24)).toBe('24分')
    expect(formatDuration(1)).toBe('1分')
    expect(formatDuration(59)).toBe('59分')
  })

  it('should format hours only', () => {
    expect(formatDuration(60)).toBe('1時間')
    expect(formatDuration(120)).toBe('2時間')
  })

  it('should format hours and minutes', () => {
    expect(formatDuration(90)).toBe('1時間30分')
    expect(formatDuration(123)).toBe('2時間3分')
    expect(formatDuration(61)).toBe('1時間1分')
  })
})

describe('formatEpisodeDuration', () => {
  it('should return null for null/undefined/0', () => {
    expect(formatEpisodeDuration(null)).toBeNull()
    expect(formatEpisodeDuration(undefined)).toBeNull()
    expect(formatEpisodeDuration(0)).toBeNull()
  })

  it('should format minutes', () => {
    expect(formatEpisodeDuration(24)).toBe('24分')
    expect(formatEpisodeDuration(1)).toBe('1分')
  })
})
