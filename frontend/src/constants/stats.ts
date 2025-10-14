/**
 * ユーザー統計情報の型定義
 */
export interface UserStats {
  totalEpisodesWatched: number;
  totalAnimesCompleted: number;
  totalWatchTimeHours: number;
  favoriteGenres: string[];
  thisMonthEpisodes: number;
}

/**
 * デフォルトのユーザー統計情報
 * APIが404を返した場合（未実装）や、データが存在しない場合に使用
 */
export const DEFAULT_USER_STATS: UserStats = {
  totalEpisodesWatched: 0,
  totalAnimesCompleted: 0,
  totalWatchTimeHours: 0,
  favoriteGenres: [],
  thisMonthEpisodes: 0,
};

/**
 * エピソードあたりの時間定数
 */
export const MINUTES_PER_EPISODE = 24;
export const HOURS_PER_EPISODE = MINUTES_PER_EPISODE / 60; // 0.4時間
