/**
 * ユーザーの視聴ステータスを判定してラベルとスタイルを返すユーティリティ
 * watching, completed, planned, dropped などのユーザー視聴状態を扱う
 * 
 * 注意: 放送状態（airing, finished, upcoming）は animeStatus.ts を使用してください
 */

export interface UserWatchStatusConfig {
  label: string
  bgColor: string
  textColor: string
}

export const USER_WATCH_STATUS_CONFIG: Record<string, UserWatchStatusConfig> = {
  watching: {
    label: '視聴中',
    bgColor: 'bg-blue-100',
    textColor: 'text-blue-700',
  },
  completed: {
    label: '完了',
    bgColor: 'bg-green-100',
    textColor: 'text-green-700',
  },
  planned: {
    label: '予定',
    bgColor: 'bg-gray-100',
    textColor: 'text-gray-700',
  },
  dropped: {
    label: '中断',
    bgColor: 'bg-orange-100',
    textColor: 'text-orange-700',
  },
}

/**
 * ユーザーの視聴ステータスを判定してラベルとスタイルを返す
 * @param status - ユーザーの視聴ステータス（watching, completed, planned, dropped）
 * @returns ラベルとCSSクラス名（bgColor/textColor形式）
 */
export const getUserWatchStatusConfig = (status: string): UserWatchStatusConfig => {
  return USER_WATCH_STATUS_CONFIG[status] || {
    label: status,
    bgColor: 'bg-gray-100',
    textColor: 'text-gray-700',
  }
}