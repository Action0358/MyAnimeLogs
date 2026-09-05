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
 * SEASON_NAMES: ['冬', '春', '夏', '秋'] → index 0-3
 * 月→インデックス: Math.floor((month % 12) / 3)
 *   1-3月→0(冬), 4-6月→1(春), 7-9月→2(夏), 10-12月→3(秋)
 */
const getCurrentSeason = (): { year: number; seasonIndex: number } => {
  const now = new Date();
  return {
    year: now.getFullYear(),
    seasonIndex: Math.floor((now.getMonth() % 12) / 3),
  };
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

/** ジャンルの選択肢（五十音順、DBの実データに存在するもののみ。アダルト向けタグは除外） */
export const GENRES = [
  'アクション',
  'アドベンチャー',
  'エッチ',
  '音楽',
  'SF',
  'コメディ',
  '心理',
  'スポーツ',
  'スリラー',
  '超常現象',
  'ドラマ',
  '日常',
  'ファンタジー',
  'ホラー',
  '魔法少女',
  'ミステリー',
  'メカ',
  'ロマンス',
] as const;

export type Season = string;
export type Genre = (typeof GENRES)[number];
