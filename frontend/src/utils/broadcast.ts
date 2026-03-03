/**
 * 放送曜日の翻訳ユーティリティ
 */

const dayMap: Record<string, string> = {
  'Mondays': '月曜日',
  'Tuesdays': '火曜日',
  'Wednesdays': '水曜日',
  'Thursdays': '木曜日',
  'Fridays': '金曜日',
  'Saturdays': '土曜日',
  'Sundays': '日曜日',
}

export const translateBroadcastDay = (day: string): string => {
  return dayMap[day] || day
}
