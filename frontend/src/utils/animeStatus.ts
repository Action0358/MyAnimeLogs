/**
 * アニメの放送状態（放送予定/放送中/完結）を判定してラベルとスタイルを返すユーティリティ
 * airing, finished などのアニメ放送状態を扱う
 * 
 * ユーザーの視聴ステータス（watching, completed, planned）のラベル・スタイルは
 * components/favorites/StatusSelector.tsx の STATUS_CONFIG を使用してください
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
 * @param status - アニメのステータス（airing, finished など）
 * @param releasedOn - 放送開始日（YYYY-MM-DD形式またはISO 8601形式）
 * @returns 放送状態
 */
const determineBroadcastStatus = (
  status: string | undefined | null,
  releasedOn?: string | undefined | null
): BroadcastStatus => {
  if (!status) {
    return 'unknown'
  }

  // 放送開始日が未来かどうかを判定
  let isUpcoming = false
  if (releasedOn) {
    try {
      // releasedOnをDateオブジェクトに変換
      const releaseDate = new Date(releasedOn)
      
      // 今日の日付（時刻を00:00:00にリセット）
      const today = new Date()
      today.setHours(0, 0, 0, 0)
      
      // 放送開始日も時刻を00:00:00にリセット
      releaseDate.setHours(0, 0, 0, 0)
      
      // 日付を比較
      isUpcoming = releaseDate > today
    } catch (error) {
      // 日付のパースに失敗した場合は無視
      console.warn('Failed to parse releasedOn date:', releasedOn, error)
    }
  }

  if (isUpcoming) {
    return 'upcoming'
  } else if (status === 'upcoming') {
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
 * @param releasedOn - 放送開始日（YYYY-MM-DD形式またはISO 8601形式）
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
 * @param releasedOn - 放送開始日（YYYY-MM-DD形式またはISO 8601形式）
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