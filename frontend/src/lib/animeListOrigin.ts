/**
 * アニメ詳細ページの「戻る」導線。
 *
 * 一覧から詳細へ遷移するとき、遷移元のパスを `from` クエリに載せる。
 * 詳細ページはそれを解決して「◯◯に戻る」を表示する。
 * クエリ付きのパス（検索条件など）をそのまま保持できるよう、
 * ラベルの識別子ではなくパスそのものを渡す方式にしている。
 *
 * 関連作品から別の作品へ進んだ場合は、直前の作品の詳細ページ（その作品自身の
 * `from` を含む）を戻り先にする。今期アニメ → 作品A → 作品B と進んだとき、
 * B から A、A から今期アニメへと、たどった順に戻れる。
 */

/** 戻り先として認めるパスと、その表示名 */
const KNOWN_ORIGINS: { pathname: string; label: string }[] = [
  { pathname: '/anime/current', label: '今期アニメ' },
  { pathname: '/anime/popular', label: '人気アニメ' },
  { pathname: '/search', label: '検索結果' },
  { pathname: '/favorites', label: 'お気に入り' },
  { pathname: '/mypage', label: 'マイページ' },
]

const HOME_ORIGIN = { href: '/', label: 'ホーム' }

/** 作品の詳細ページ。遷移前の作品はユーザーが把握しているため、作品名ではなく共通の表示名にする */
const DETAIL_PATHNAME = /^\/anime\/\d+$/
const DETAIL_LABEL = '前の作品'

/**
 * 戻り先としてさかのぼれる作品詳細の段数。関連作品をたどるたびに from が入れ子になり
 * URLが伸びるため、これを超えた分は最も古い戻り先から捨てる
 * （捨てられた先の作品の「戻る」はホームになる）。
 */
export const MAX_DETAIL_BACK_DEPTH = 5

export interface BackTarget {
  href: string
  label: string
}

/**
 * 末尾スラッシュを取り除いて比較用のパスに揃える。
 *
 * 本番ビルドは next.config.js で trailingSlash を有効にしているため、
 * usePathname() は `/anime/current/` を返す。開発サーバーでは無効なので
 * `/anime/current` になる。この差を吸収しないと、本番でだけ既知の一覧に
 * 一致せずホームへフォールバックする（ローカルでは再現しない）。
 * ルート `/` は空文字にしないでそのまま残す。
 */
const normalizePathname = (pathname: string): string =>
  pathname.length > 1 ? pathname.replace(/\/+$/, '') : pathname

/**
 * `from` クエリの値を戻り先に解決する。
 *
 * 外部URLへ飛ばされないよう、ホスト付きURLやプロトコル相対URL（`//`）は
 * 受け付けず、既知のパスに一致したものだけを採用する。
 * 解決できない場合はホームにフォールバックする。
 */
export function resolveBackTarget(from: string | null | undefined): BackTarget {
  if (!from) return HOME_ORIGIN

  let decoded: string
  try {
    decoded = decodeURIComponent(from)
  } catch {
    // 不正なパーセントエンコーディング
    return HOME_ORIGIN
  }

  if (!decoded.startsWith('/') || decoded.startsWith('//')) return HOME_ORIGIN

  const pathname = normalizePathname(decoded.split('?')[0])
  if (DETAIL_PATHNAME.test(pathname)) {
    // 作品詳細の戻り先は、その作品自身の from（エンコード済み）を含む。
    // もう一度デコードすると入れ子の from に含まれる & や ? が露出して壊れるため、
    // URLSearchParams で取り出した値（先頭が / のもの）はそのまま使う
    const href = from.startsWith('/') ? from : decoded
    return { href, label: DETAIL_LABEL }
  }

  const origin = KNOWN_ORIGINS.find((o) => o.pathname === pathname)
  if (!origin) return HOME_ORIGIN

  return { href: decoded, label: origin.label }
}

/**
 * 詳細ページへのリンクに `from` を付与する。
 * 遷移元が既知の一覧でない場合（ホームなど）は付けない。
 */
export function withBackParam(href: string, from: string | undefined): string {
  if (!from) return href
  const pathname = normalizePathname(from.split('?')[0])
  if (DETAIL_PATHNAME.test(pathname)) {
    return `${href}?from=${encodeURIComponent(limitDetailDepth(from, 1))}`
  }
  if (!KNOWN_ORIGINS.some((o) => o.pathname === pathname)) return href
  return `${href}?from=${encodeURIComponent(from)}`
}

/**
 * 作品詳細の戻り先の入れ子を MAX_DETAIL_BACK_DEPTH 段までに切り詰める。
 * depth は path 自身が何段目の作品詳細か（1始まり）。
 */
function limitDetailDepth(path: string, depth: number): string {
  const queryStart = path.indexOf('?')
  const rawPathname = queryStart === -1 ? path : path.slice(0, queryStart)
  const query = queryStart === -1 ? '' : path.slice(queryStart + 1)
  if (!DETAIL_PATHNAME.test(normalizePathname(rawPathname))) return path

  const inner = new URLSearchParams(query).get('from')
  if (!inner || depth >= MAX_DETAIL_BACK_DEPTH) return rawPathname
  return `${rawPathname}?from=${encodeURIComponent(limitDetailDepth(inner, depth + 1))}`
}
