/**
 * アニメ詳細ページの「戻る」導線。
 *
 * 一覧から詳細へ遷移するとき、遷移元のパスを `from` クエリに載せる。
 * 詳細ページはそれを解決して「◯◯に戻る」を表示する。
 * クエリ付きのパス（検索条件など）をそのまま保持できるよう、
 * ラベルの識別子ではなくパスそのものを渡す方式にしている。
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
  if (!KNOWN_ORIGINS.some((o) => o.pathname === pathname)) return href
  return `${href}?from=${encodeURIComponent(from)}`
}
