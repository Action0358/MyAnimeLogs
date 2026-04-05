import { getUserWatchStatusConfig, USER_WATCH_STATUS_CONFIG } from '../userWatchStatus'

describe('getUserWatchStatusConfig', () => {
  it('should return config for watching', () => {
    const result = getUserWatchStatusConfig('watching')
    expect(result.label).toBe('視聴中')
    expect(result.bgColor).toBe('bg-blue-100')
    expect(result.textColor).toBe('text-blue-700')
  })

  it('should return config for completed', () => {
    const result = getUserWatchStatusConfig('completed')
    expect(result.label).toBe('完了')
  })

  it('should return config for planned', () => {
    const result = getUserWatchStatusConfig('planned')
    expect(result.label).toBe('予定')
  })

  it('should return config for dropped', () => {
    const result = getUserWatchStatusConfig('dropped')
    expect(result.label).toBe('中断')
  })

  it('should return default config for unknown status', () => {
    const result = getUserWatchStatusConfig('unknown')
    expect(result.label).toBe('unknown')
    expect(result.bgColor).toBe('bg-gray-100')
  })
})

describe('USER_WATCH_STATUS_CONFIG', () => {
  it('should have all 4 statuses defined', () => {
    expect(Object.keys(USER_WATCH_STATUS_CONFIG)).toEqual([
      'watching', 'completed', 'planned', 'dropped',
    ])
  })
})
