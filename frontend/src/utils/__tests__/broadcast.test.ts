import { translateBroadcastDay } from '../broadcast'

describe('translateBroadcastDay', () => {
  it('should translate English day names to Japanese', () => {
    expect(translateBroadcastDay('Mondays')).toBe('月曜日')
    expect(translateBroadcastDay('Tuesdays')).toBe('火曜日')
    expect(translateBroadcastDay('Wednesdays')).toBe('水曜日')
    expect(translateBroadcastDay('Thursdays')).toBe('木曜日')
    expect(translateBroadcastDay('Fridays')).toBe('金曜日')
    expect(translateBroadcastDay('Saturdays')).toBe('土曜日')
    expect(translateBroadcastDay('Sundays')).toBe('日曜日')
  })

  it('should return original string for unknown day', () => {
    expect(translateBroadcastDay('Unknown')).toBe('Unknown')
    expect(translateBroadcastDay('')).toBe('')
  })
})
