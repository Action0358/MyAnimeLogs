/**
 * 検索関連の定数
 */

/**
 * シーズン定義（時系列順）
 * 日本のアニメ放送クールに対応:
 *   冬(1-3月) → 春(4-6月) → 夏(7-9月) → 秋(10-12月)
 */
const SEASON_NAMES = ['冬', '春', '夏', '秋'] as const;

/** 同期開始: 1979年春（Annict API: 初代ガンダム） */
const START_YEAR = 1979;
const START_SEASON_INDEX = 1; // 春

/**
 * 現在の月から放送年・シーズンインデックスを取得
 *
 * 注意: 冬クール(1-3月)はその年の冬として扱う
 * 例: 2026年3月 → 2026年冬
 */
const getCurrentSeason = (): { year: number; seasonIndex: number } => {
  const now = new Date();
  const year = now.getFullYear();
  const month = now.getMonth() + 1;

  if (month <= 3) return { year, seasonIndex: 0 };  // 冬
  if (month <= 6) return { year, seasonIndex: 1 };  // 春
  if (month <= 9) return { year, seasonIndex: 2 };  // 夏
  return { year, seasonIndex: 3 };                    // 秋
};

/**
 * 放送時期の選択肢を動的生成（現在のシーズン〜1979年春まで新しい順）
 *
 * 年+シーズンを通し番号(year*4+seasonIndex)に変換し降順ループすることで、
 * 未来のシーズンや開始年以前のシーズンを確実に除外する
 */
const generateSeasons = (): string[] => {
  const { year: currentYear, seasonIndex: currentSeasonIdx } = getCurrentSeason();

  const currentSerial = currentYear * 4 + currentSeasonIdx;
  const startSerial = START_YEAR * 4 + START_SEASON_INDEX;

  const seasons: string[] = [];

  for (let serial = currentSerial; serial >= startSerial; serial--) {
    const year = Math.floor(serial / 4);
    const seasonIdx = serial % 4;
    seasons.push(`${year}年${SEASON_NAMES[seasonIdx]}`);
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
