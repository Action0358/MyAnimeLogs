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
  if (!status) {
    return { label: null, className: '' }
  }

  const today = new Date()
  const releasedDate = releasedOn ? new Date(releasedOn) : null
  const isUpcoming = releasedDate && releasedDate > today

  if (isUpcoming) {
    return { label: '放送予定', className: 'bg-amber-100 text-amber-800' }
  } else if (status === 'airing') {
    return { label: '放送中', className: 'bg-green-100 text-green-800' }
  } else if (status === 'finished') {
    return { label: '完結', className: 'bg-gray-100 text-gray-800' }
  }
  return { label: null, className: '' }
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
  const today = new Date()
  const releasedDate = releasedOn ? new Date(releasedOn) : null
  const isUpcoming = releasedDate && releasedDate > today

  if (isUpcoming) {
    return { label: '放送予定', bgColor: 'bg-amber-100', textColor: 'text-amber-800' }
  } else if (status === 'airing') {
    return { label: '放送中', bgColor: 'bg-green-100', textColor: 'text-green-800' }
  } else if (status === 'finished') {
    return { label: '完結', bgColor: 'bg-gray-100', textColor: 'text-gray-800' }
  }
  return { label: '不明', bgColor: 'bg-gray-100', textColor: 'text-gray-800' }
}
