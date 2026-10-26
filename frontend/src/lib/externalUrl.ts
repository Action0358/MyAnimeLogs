/**
 * 外部データ由来のURLを、リンク先として安全な場合だけ返す。
 *
 * 公式サイトURLなどは Annict 等の第三者が編集できるデータから来る。
 * React 18 は href の javascript: URL を止めず、CSP も 'unsafe-inline' のため
 * 止められない。http/https 以外や解釈できない値は null を返し、リンクを出さない。
 * バックエンドでも同じ検証をしているが、表示側でも重ねて確かめる。
 */
export const toSafeExternalUrl = (raw: string | null | undefined): URL | null => {
  if (!raw) return null
  try {
    const url = new URL(raw.trim())
    return url.protocol === 'http:' || url.protocol === 'https:' ? url : null
  } catch {
    return null
  }
}
