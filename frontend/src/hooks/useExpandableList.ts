import { useState } from 'react'

/**
 * 件数が多い一覧を初期表示件数で折りたたみ、ボタンで展開するための共通フック。
 * 関連作品（RelatedAnimeGrid）と同じしきい値調整ロジックを再利用する。
 *
 * @param oneRowAtWidest 残りがこの件数以内なら、クリックさせるほどの量ではないため
 *   折りたたまずそのまま全件表示する（グリッドの最大列数に合わせる）
 */
export function useExpandableList<T>(items: T[], initialCount: number, oneRowAtWidest = 6) {
  const [isExpanded, setIsExpanded] = useState(false)

  const rawHiddenCount = items.length - initialCount
  const fitsInOneMoreRow = rawHiddenCount > 0 && rawHiddenCount <= oneRowAtWidest
  const effectiveInitialCount = fitsInOneMoreRow ? items.length : initialCount
  const hiddenCount = items.length - effectiveInitialCount
  const visible = isExpanded ? items : items.slice(0, effectiveInitialCount)

  return {
    visible,
    hiddenCount,
    isExpanded,
    toggle: () => setIsExpanded((prev) => !prev),
    reset: () => setIsExpanded(false),
  }
}
