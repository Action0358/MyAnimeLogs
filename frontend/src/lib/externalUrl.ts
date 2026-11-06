/**
 * 外部データ由来のURLを、リンク先として安全な場合だけ返す。
 *
 * 公式サイトURLなどは Annict 等の第三者が編集できるデータから来る。
 *
 * React 19 は href の javascript: URL を、クリックすると例外を投げるだけの URL に
 * 置き換えて止める（大文字・前後の空白・タブ入りも同様）。また STG・本番のビルドでは
 * ページごとの CSP（scripts/apply-csp-hashes.js）が javascript: URL の実行を止める。
 * ただし React は data: や vbscript: などをそのまま出力し、ローカルの next dev には
 * ページごとの CSP がない。React や CSP だけに頼らず、http/https 以外や解釈できない
 * 値は null を返してリンクを出さない。
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
