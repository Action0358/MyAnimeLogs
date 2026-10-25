#!/usr/bin/env node

/**
 * エッジ（functions/_middleware.js）が使う設定を out/_edge-config.json に書き出す。
 * メンテナンス判定でバックエンドの GET /api/maintenance を呼ぶため、API の URL を渡す。
 *
 * 値はビルドに使う NEXT_PUBLIC_API_URL と同じ（フロントの JS に含まれる公開情報）。
 * Cloudflare 側に環境変数を持たせず、URL の正本を GitHub Actions の secrets に一本化する。
 * 未設定なら書き出さない（エッジは /maintenance.json だけで判定する）。
 */

const fs = require('fs')
const path = require('path')

const outDir = path.join(__dirname, '..', 'out')
const apiUrl = process.env.NEXT_PUBLIC_API_URL

if (!fs.existsSync(outDir)) {
  console.error('エラー: out/ がありません。先に "next build" を実行してください。')
  process.exit(1)
}

if (!apiUrl) {
  console.warn('警告: NEXT_PUBLIC_API_URL が未設定のため _edge-config.json を書き出しません。')
  process.exit(0)
}

fs.writeFileSync(path.join(outDir, '_edge-config.json'), `${JSON.stringify({ apiUrl })}\n`)
console.log(`_edge-config.json を書き出しました（apiUrl: ${apiUrl}）`)
