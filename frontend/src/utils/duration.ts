/**
 * 分を時間表記に変換
 * @param minutes - 分数
 * @returns フォーマットされた時間文字列（例: "2時間3分", "24分"）
 */
export const formatDuration = (
  minutes: number | null | undefined
): string | null => {
  if (!minutes || minutes <= 0) return null;

  const hours = Math.floor(minutes / 60);
  const mins = minutes % 60;

  if (hours > 0 && mins > 0) {
    return `${hours}時間${mins}分`;
  } else if (hours > 0) {
    return `${hours}時間`;
  } else {
    return `${mins}分`;
  }
};

/**
 * エピソードの放送時間を表示用にフォーマット
 * @param minutes - 分数
 * @returns フォーマットされた文字列（例: "24分"）
 */
export const formatEpisodeDuration = (
  minutes: number | null | undefined
): string | null => {
  if (!minutes || minutes <= 0) return null;
  return `${minutes}分`;
};
