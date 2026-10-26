/**
 * シーズン名を日本語に変換するユーティリティ
 */

const SEASON_MAP: Record<string, string> = {
  winter: '冬',
  spring: '春',
  summer: '夏',
  autumn: '秋',
  fall: '秋',
}

/**
 * 英語の季節名を日本語に変換
 */
export const translateSeason = (season: string): string => {
  return SEASON_MAP[season.toLowerCase()] || season
}

/**
 * シーズン文字列をフォーマット
 * 例: "2026-winter" → "2026年冬"
 *     "winter" → "冬"
 *     "2026年 winter" → "2026年冬"
 */
export const formatSeason = (seasonStr?: string | null): string | undefined => {
  if (!seasonStr) return undefined

  // "2026-winter" 形式
  const dashMatch = seasonStr.match(/^(\d{4})-(\w+)$/)
  if (dashMatch) {
    const [, year, season] = dashMatch
    return `${year}年${translateSeason(season)}`
  }

  // "2026年 winter" または "2026年winter" 形式
  const yearSeasonMatch = seasonStr.match(/^(\d{4})年\s*(\w+)$/)
  if (yearSeasonMatch) {
    const [, year, season] = yearSeasonMatch
    return `${year}年${translateSeason(season)}`
  }

  // "2026 winter" 形式
  const spaceMatch = seasonStr.match(/^(\d{4})\s+(\w+)$/)
  if (spaceMatch) {
    const [, year, season] = spaceMatch
    return `${year}年${translateSeason(season)}`
  }

  // 季節名のみの場合
  if (SEASON_MAP[seasonStr.toLowerCase()]) {
    return translateSeason(seasonStr)
  }

  // その他はそのまま返す
  return seasonStr
}
