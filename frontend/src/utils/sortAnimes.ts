import type { DBAnime } from '@/lib/api/annict'

/**
 * 今期・人気アニメ一覧の並び替え。検索ページの並び替え選択肢（year/popularity）と
 * 同じ値・ラベルを使うが、ここでは取得済みの一覧をクライアント側で並べ替えるだけなので、
 * 検索のSQL ORDER BYとは実装が異なる（放送年ではなくreleased_onの日付で比較する）。
 */
export const sortAnimes = (animes: DBAnime[], sort: string): DBAnime[] => {
  if (sort === 'year') {
    // released_onが無い作品は末尾に送る（古い順）
    return [...animes].sort((a, b) => {
      if (!a.released_on && !b.released_on) return 0
      if (!a.released_on) return 1
      if (!b.released_on) return -1
      return a.released_on.localeCompare(b.released_on)
    })
  }

  // popularity（デフォルト）: watchers_count降順。APIが既にこの順で返すため並べ替え不要
  return animes
}
