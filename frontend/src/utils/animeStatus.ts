/**
 * アニメの放送状態を判定してラベルとスタイルを返すユーティリティ
 */

export interface AnimeStatusLabel {
  label: string | null
  className: string
}

export interface AnimeStatusConfig {
  label: string
  bgColor: string
  textColor: string
}

// 放送状態の種類
type BroadcastStatus = 'upcoming' | 'airing' | 'finished' | 'unknown'

/**
 * 放送状態を判定する共通ヘルパー関数
 * タイムゾーンの問題を回避するため、日付文字列を直接比較
 * @param status - アニメのステータス（airing, finished など）
 * @param releasedOn - 放送開始日（YYYY-MM-DD形式）
 * @returns 放送状態
 */
const determineBroadcastStatus = (
  status: string | undefined | null,
  releasedOn?: string | undefined | null
): BroadcastStatus => {
  if (!status) {
    return 'unknown'
  }

  // YYYY-MM-DD形式の今日の日付を取得（タイムゾーン問題を回避）
  const today = new Date()
  const todayStr = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`

  // 日付文字列を直接比較（releasedOnがYYYY-MM-DD形式の場合）
  const isUpcoming = releasedOn && releasedOn > todayStr

  if (isUpcoming) {
    return 'upcoming'
  } else if (status === 'airing') {
    return 'airing'
  } else if (status === 'finished') {
    return 'finished'
  }
  return 'unknown'
}

/**
 * アニメの放送状態を判定してラベルとスタイルを返す
 * @param status - アニメのステータス（airing, finished など）
 * @param releasedOn - 放送開始日（YYYY-MM-DD形式）
 * @returns ラベルとCSSクラス名
 */
export const getAnimeStatusLabel = (
  status: string | undefined | null,
  releasedOn?: string | undefined | null
): AnimeStatusLabel => {
  const broadcastStatus = determineBroadcastStatus(status, releasedOn)

  switch (broadcastStatus) {
    case 'upcoming':
      return { label: '放送予定', className: 'bg-amber-100 text-amber-800' }
    case 'airing':
      return { label: '放送中', className: 'bg-green-100 text-green-800' }
    case 'finished':
      return { label: '完結', className: 'bg-gray-100 text-gray-800' }
    default:
      return { label: null, className: '' }
  }
}

/**
 * アニメの放送状態を判定してラベルとスタイルを返す（bgColor/textColor形式）
 * MyPageContentなどで使用
 * @param status - アニメのステータス（airing, finished など）
 * @param releasedOn - 放送開始日（YYYY-MM-DD形式）
 * @returns ラベルとCSSクラス名（bgColor/textColor形式）
 */
export const getAnimeStatusConfig = (
  status: string | undefined | null,
  releasedOn?: string | undefined | null
): AnimeStatusConfig => {
  const broadcastStatus = determineBroadcastStatus(status, releasedOn)

  switch (broadcastStatus) {
    case 'upcoming':
      return { label: '放送予定', bgColor: 'bg-amber-100', textColor: 'text-amber-800' }
    case 'airing':
      return { label: '放送中', bgColor: 'bg-green-100', textColor: 'text-green-800' }
    case 'finished':
      return { label: '完結', bgColor: 'bg-gray-100', textColor: 'text-gray-800' }
    default:
      return { label: '不明', bgColor: 'bg-gray-100', textColor: 'text-gray-800' }
  }
}
