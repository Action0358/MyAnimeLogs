/**
 * 検索関連の定数
 */

/** シーズン順序（新しい順） */
const SEASON_NAMES = ['冬', '秋', '夏', '春'] as const;

/** 同期開始年（Annict API） */
const START_YEAR = 1979;

/**
 * 放送時期の選択肢を動的生成（現在のシーズン〜1979年春まで新しい順）
 */
const generateSeasons = (): string[] => {
  const now = new Date();
  const currentYear = now.getFullYear();
  const currentMonth = now.getMonth() + 1;

  // 現在の月からシーズンを判定
  // 1-3月: 冬, 4-6月: 春, 7-9月: 夏, 10-12月: 秋
  let currentSeasonIndex: number;
  if (currentMonth <= 3) currentSeasonIndex = 0;      // 冬
  else if (currentMonth <= 6) currentSeasonIndex = 3;  // 春
  else if (currentMonth <= 9) currentSeasonIndex = 2;  // 夏
  else currentSeasonIndex = 1;                          // 秋

  const seasons: string[] = [];

  for (let year = currentYear; year >= START_YEAR; year--) {
    const startIndex = year === currentYear ? currentSeasonIndex : 0;
    const endIndex = year === START_YEAR ? 3 : SEASON_NAMES.length - 1;

    for (let i = startIndex; i <= endIndex; i++) {
      seasons.push(`${year}年${SEASON_NAMES[i]}`);
    }
  }

  return seasons;
};

/** 放送時期の選択肢（動的生成） */
export const SEASONS = generateSeasons();

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

export type Season = string;
export type Genre = (typeof GENRES)[number];
