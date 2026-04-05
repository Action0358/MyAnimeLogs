import { getAnimeStatusLabel, getAnimeBroadcastStatusConfig } from '../animeStatus'

// 固定時刻: 2025年5月15日 12:00:00 UTC
const FIXED_NOW = new Date('2025-05-15T12:00:00Z')

beforeEach(() => {
  jest.useFakeTimers()
  jest.setSystemTime(FIXED_NOW)
})

afterEach(() => {
  jest.useRealTimers()
})

describe('getAnimeStatusLabel', () => {
  it('should return null label for null/undefined status', () => {
    expect(getAnimeStatusLabel(null)).toEqual({ label: null, className: '' })
    expect(getAnimeStatusLabel(undefined)).toEqual({ label: null, className: '' })
  })

  it('should return "放送中" for airing status', () => {
    const result = getAnimeStatusLabel('airing')
    expect(result.label).toBe('放送中')
  })

  it('should return "完結" for finished status', () => {
    const result = getAnimeStatusLabel('finished')
    expect(result.label).toBe('完結')
  })

  it('should return "放送予定" when released_on is in the future', () => {
    const result = getAnimeStatusLabel('airing', '2025-06-15')
    expect(result.label).toBe('放送予定')
  })

  it('should return "放送中" when released_on is in the past and status is airing', () => {
    const result = getAnimeStatusLabel('airing', '2020-01-01')
    expect(result.label).toBe('放送中')
  })

  it('should return "放送予定" when released_on is in the future regardless of status', () => {
    const result = getAnimeStatusLabel('finished', '2025-06-15')
    expect(result.label).toBe('放送予定')
  })

  it('should return "放送中" when released_on is today', () => {
    const result = getAnimeStatusLabel('airing', '2025-05-15')
    expect(result.label).toBe('放送中')
  })
})

describe('getAnimeBroadcastStatusConfig', () => {
  it('should return config with label, bgColor, textColor', () => {
    const result = getAnimeBroadcastStatusConfig('airing')
    expect(result.label).toBe('放送中')
    expect(result.bgColor).toBeDefined()
    expect(result.textColor).toBeDefined()
  })

  it('should return "不明" for unknown status', () => {
    const result = getAnimeBroadcastStatusConfig(null)
    expect(result.label).toBe('不明')
  })
})
