/**
 * メンテナンス判定でフロント本体（React）とエッジ（functions/_middleware.js）が共有する定義。
 * エッジからも import するため、React・Next.js・ブラウザ API に依存させないこと。
 */

/**
 * メンテナンス画面のページ。エッジはメンテナンス中、このページの静的 HTML を
 * どの URL に対しても返す（URL は変えない）。利用者がこの URL を開くことは想定しない。
 */
export const MAINTENANCE_PAGE_PATH = '/maintenance'

/**
 * メンテナンス中も開けるパス。お知らせ（/news）は静的生成で API に依存しないため、
 * メンテナンス画面の「お知らせを見る」の行き先として生かしておく。
 * 静的エクスポートは trailingSlash のため /news/ も来る。
 */
export const isMaintenanceExemptPath = (pathname: string | null): boolean =>
  pathname === '/news' || !!pathname?.startsWith('/news/')

/**
 * エッジがメンテナンス中の開放ページ（/news）に差し込む meta の name。
 * content は終了予定（ISO 8601、なければ空）。クライアントはこれを見て
 * API を呼ばずにメンテナンス状態になる。
 */
export const MAINTENANCE_META_NAME = 'myanimelogs-maintenance'

/**
 * エッジがメンテナンス中の開放ページ（/news）で <html> に付ける属性。
 * JS が動く前から、CSS（globals.css）でヘッダー・フッターをメンテナンス用の見た目にする。
 */
export const MAINTENANCE_HTML_ATTR = 'data-maintenance'

/** 終了予定の行。エッジは終了予定があるとき、この要素の hidden を外す */
export const ENDS_AT_ROW_ATTR = 'data-maintenance-ends-at-row'
/** 終了予定の文言を入れる要素。エッジは「終了予定：…」を差し込む */
export const ENDS_AT_TEXT_ATTR = 'data-maintenance-ends-at'

/** 「2026年10月25日 15:00ごろ」形式（JST）。解釈できない値は null */
export const formatEndsAt = (endsAt?: string | null): string | null => {
  if (!endsAt) return null
  const date = new Date(endsAt)
  if (Number.isNaN(date.getTime())) return null
  const parts = new Intl.DateTimeFormat('ja-JP', {
    timeZone: 'Asia/Tokyo',
    year: 'numeric',
    month: 'numeric',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).formatToParts(date)
  const get = (type: string) => parts.find((p) => p.type === type)?.value ?? ''
  // hour12: false でも環境により "24" が返ることがあるため 0 時に寄せる
  const hour = get('hour') === '24' ? '00' : get('hour')
  return `${get('year')}年${get('month')}月${get('day')}日 ${hour}:${get('minute')}ごろ`
}

/** 終了予定の行に出す文言。終了予定がなければ null */
export const endsAtText = (endsAt?: string | null): string | null => {
  const label = formatEndsAt(endsAt)
  return label ? `終了予定：${label}` : null
}
