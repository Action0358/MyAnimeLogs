import { useMediaQuery } from './useMediaQuery'

/** Tailwindの`md`ブレークポイント（デフォルト768px）をPC/SPの境とする */
const PC_MEDIA_QUERY = '(min-width: 768px)'

/**
 * 一覧の初期表示件数をSP/PCで分けるためのフック。
 * グリッドの列数がSP3列・PC6列（マイページ/お気に入り）であることに合わせ、
 * 「1行ぶん以内の残りは折りたたまない」しきい値もSP/PCで変える。
 */
export function useResponsiveInitialCount(spCount: number, pcCount: number) {
  const isPc = useMediaQuery(PC_MEDIA_QUERY)
  return {
    initialCount: isPc ? pcCount : spCount,
    oneRowAtWidest: isPc ? 6 : 3,
  }
}
