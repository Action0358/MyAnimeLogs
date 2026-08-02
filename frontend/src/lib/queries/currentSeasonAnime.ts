import { animeApi } from '@/lib/api/annict'
import { HOME_CURRENT_SEASON_ANIME_LIMIT } from '@/lib/constants'

/**
 * トップページの「今期アニメ」取得クエリ。
 *
 * ヒーロー背景（HeroBackdrop）とカルーセル（CurrentSeasonAnime）が
 * 同じキャッシュを共有して追加リクエストを発生させないための定義。
 * TanStack Query は queryKey が同じなら先にマウントした側の queryFn だけを実行するため、
 * キーと取得条件が別々に書かれていると、片方の limit を変えたときに
 * もう片方が意図しない件数を受け取る。両者を必ず一緒に使うこと。
 */
export const currentSeasonAnimeQuery = {
  queryKey: ['currentSeasonAnime'] as const,
  queryFn: () => animeApi.getCurrentSeasonAnimes(HOME_CURRENT_SEASON_ANIME_LIMIT),
  staleTime: 1000 * 60 * 15,
  gcTime: 1000 * 60 * 30,
}
