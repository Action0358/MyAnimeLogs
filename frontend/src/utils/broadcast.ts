/**
 * 放送曜日・時間の翻訳・フォーマットユーティリティ
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

export const formatBroadcastInfo = (
  day?: string | null,
  time?: string | null
): string | null => {
  if (!day) return null
  const translatedDay = translateBroadcastDay(day)
  return time ? `${translatedDay} ${time}` : translatedDay
}
