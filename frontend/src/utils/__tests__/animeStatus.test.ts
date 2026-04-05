import { getAnimeStatusLabel, getAnimeBroadcastStatusConfig } from '../animeStatus'

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
    const futureDate = new Date()
    futureDate.setDate(futureDate.getDate() + 30)
    const futureDateStr = futureDate.toISOString().split('T')[0]

    const result = getAnimeStatusLabel('airing', futureDateStr)
    expect(result.label).toBe('放送予定')
  })

  it('should return "放送中" when released_on is in the past and status is airing', () => {
    const result = getAnimeStatusLabel('airing', '2020-01-01')
    expect(result.label).toBe('放送中')
  })

  it('should return "放送予定" when released_on is in the future regardless of status', () => {
    const futureDate = new Date()
    futureDate.setDate(futureDate.getDate() + 30)
    const futureDateStr = futureDate.toISOString().split('T')[0]

    const result = getAnimeStatusLabel('finished', futureDateStr)
    expect(result.label).toBe('放送予定')
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
