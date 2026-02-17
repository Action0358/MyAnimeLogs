/**
 * アニメの放送状態（放送予定/放送中/完結）を判定してラベルとスタイルを返すユーティリティ
 * airing, finished などのアニメ放送状態を扱う
 * 
 * ユーザーの視聴ステータス（watching, completed, planned, dropped）は
 * userWatchStatus.ts を使用してください
 */

export interface BroadcastStatusLabel {
  label: string | null
  className: string
}

export interface BroadcastStatusConfig {
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
): BroadcastStatusLabel => {
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
 * @param status - アニメのステータス（airing, finished など）
 * @param releasedOn - 放送開始日（YYYY-MM-DD形式）
 * @returns ラベルとCSSクラス名（bgColor/textColor形式）
 */
export const getAnimeBroadcastStatusConfig = (
  status: string | undefined | null,
  releasedOn?: string | undefined | null
): BroadcastStatusConfig => {
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

// 後方互換性のためのエイリアス
// AnimeStatusLabel と AnimeStatusConfig は BroadcastStatusLabel と BroadcastStatusConfig のエイリアス
export type AnimeStatusLabel = BroadcastStatusLabel
export type AnimeStatusConfig = BroadcastStatusConfig