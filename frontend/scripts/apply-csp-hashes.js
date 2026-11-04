#!/usr/bin/env node

/**
 * out/ の各 HTML に、そのページのインラインスクリプトだけを許可する CSP の meta を入れる（V-08）。
 *
 * Next.js の静的出力はページごとに異なるインラインスクリプトを含むため、'unsafe-inline' を
 * 外すにはページごとのハッシュが要る。一方 _headers は全ページ共通でしか書けず、1行の長さにも
 * 上限がある。そこで _headers の CSP はそのまま残し、ページごとの制限を meta で重ねる。
 * ブラウザは両方の CSP を満たすものだけを実行するので、meta のハッシュ以外のインライン
 * スクリプト・イベントハンドラ・javascript: URL は止まる。
 *
 * copy-headers.js の後に実行する（out/_headers を読み、out/404.html も対象にするため）。
 */

const crypto = require('crypto')
const fs = require('fs')
const path = require('path')

const INLINE_SCRIPT_PATTERN = /<script\b([^>]*)>([\s\S]*?)<\/script\s*>/gi
const SRC_ATTRIBUTE_PATTERN = /(?:^|\s)src\s*=/i
const HEAD_START_PATTERN = /<head\b[^>]*>(?:\s*<meta\s+charset=["']?utf-8["']?\s*\/?>)?/i
const SCRIPT_SRC_PATTERN = /\bscript-src\s+([^;]+)/

/** インラインスクリプトの sha256 ソース（重複なし） */
const collectInlineScriptHashes = (html) => {
  const hashes = new Set()
  for (const [, attributes, content] of html.matchAll(INLINE_SCRIPT_PATTERN)) {
    if (SRC_ATTRIBUTE_PATTERN.test(attributes)) continue
    // ブラウザは HTML を解析する時点で改行を LF に揃えるので、揃えてからハッシュを取る
    const normalized = content.replace(/\r\n?/g, '\n')
    const digest = crypto.createHash('sha256').update(normalized, 'utf8').digest('base64')
    hashes.add(`'sha256-${digest}'`)
  }
  return [...hashes]
}

/** _headers の CSP から script-src の許可元を取り出し、'unsafe-inline' を除いて返す */
const readScriptSources = (headersText) => {
  const match = headersText.match(SCRIPT_SRC_PATTERN)
  if (!match) throw new Error('_headers に script-src がありません')
  return match[1]
    .trim()
    .split(/\s+/)
    .filter((source) => source !== "'unsafe-inline'")
}

/**
 * HTML に CSP の meta を入れる。インラインスクリプトより前に置かないと効かないので
 * <head> の先頭に入れる。ただし文字コードの指定は先頭 1024 バイト以内に必要なため、
 * <meta charset> があればその直後にする。
 */
const injectCspMeta = (html, scriptSources) => {
  const headStart = html.match(HEAD_START_PATTERN)
  if (!headStart) throw new Error('<head> が見つかりません')
  const sources = [...scriptSources, ...collectInlineScriptHashes(html)].join(' ')
  const meta = `<meta http-equiv="Content-Security-Policy" content="script-src ${sources}"/>`
  const insertAt = headStart.index + headStart[0].length
  return html.slice(0, insertAt) + meta + html.slice(insertAt)
}

const findHtmlFiles = (dir) =>
  fs.readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const entryPath = path.join(dir, entry.name)
    if (entry.isDirectory()) return findHtmlFiles(entryPath)
    return entry.name.endsWith('.html') ? [entryPath] : []
  })

const main = () => {
  const outDir = path.join(__dirname, '..', 'out')
  const scriptSources = readScriptSources(fs.readFileSync(path.join(outDir, '_headers'), 'utf8'))
  const htmlFiles = findHtmlFiles(outDir)
  if (htmlFiles.length === 0) throw new Error('out/ に HTML がありません')

  for (const file of htmlFiles) {
    fs.writeFileSync(file, injectCspMeta(fs.readFileSync(file, 'utf8'), scriptSources))
  }
  console.log(`✓ ${htmlFiles.length} 件の HTML に CSP のハッシュを入れました`)
}

if (require.main === module) {
  try {
    main()
  } catch (error) {
    console.error(`エラー: CSP のハッシュを入れられませんでした: ${error.message}`)
    process.exit(1)
  }
}

module.exports = { collectInlineScriptHashes, injectCspMeta, readScriptSources }
