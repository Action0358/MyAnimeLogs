/**
 * 検索関連の定数
 */

/** 放送時期の選択肢 */
export const SEASONS = [
  '2025年冬',
  '2024年秋',
  '2024年夏',
  '2024年春',
  '2024年冬',
  '2023年秋',
] as const;

/** ジャンルの選択肢（五十音順） */
export const GENRES = [
  'アクション',
  'アドベンチャー',
  'SF',
  'コメディ',
  'ドラマ',
  'ファンタジー',
  'ホラー',
  'ロマンス',
] as const;

export type Season = (typeof SEASONS)[number];
export type Genre = (typeof GENRES)[number];
