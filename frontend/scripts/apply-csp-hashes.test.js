/**
 * @jest-environment node
 */
const crypto = require('crypto')
const {
  collectInlineScriptHashes,
  injectCspMeta,
  readScriptSources,
} = require('./apply-csp-hashes')

const sha256 = (text) =>
  `'sha256-${crypto.createHash('sha256').update(text, 'utf8').digest('base64')}'`

describe('collectInlineScriptHashes', () => {
  it('インラインスクリプトの本文をハッシュする', () => {
    expect(
      collectInlineScriptHashes('<script>a()</script><script id="x">日本語()</script>')
    ).toEqual([sha256('a()'), sha256('日本語()')])
  })

  it('src 付きのスクリプトは対象外にする', () => {
    expect(collectInlineScriptHashes('<script async="" src="/_next/a.js"></script>')).toEqual([])
  })

  it('data-src はインラインスクリプトとして扱う', () => {
    expect(collectInlineScriptHashes('<script data-src="x">a()</script>')).toEqual([sha256('a()')])
  })

  it('同じ本文は1件にまとめる', () => {
    expect(collectInlineScriptHashes('<script>a()</script><script>a()</script>')).toEqual([
      sha256('a()'),
    ])
  })

  it('ブラウザと同じく改行を LF に揃えてハッシュする', () => {
    expect(collectInlineScriptHashes('<script>a()\r\nb()\rc()</script>')).toEqual([
      sha256('a()\nb()\nc()'),
    ])
  })
})

describe('readScriptSources', () => {
  it("script-src から 'unsafe-inline' だけを除く", () => {
    const headers =
      "  Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline' https://static.cloudflareinsights.com; style-src 'self' 'unsafe-inline';"
    expect(readScriptSources(headers)).toEqual(["'self'", 'https://static.cloudflareinsights.com'])
  })

  it('script-src がなければエラーにする', () => {
    expect(() => readScriptSources("Content-Security-Policy: default-src 'self';")).toThrow(
      'script-src'
    )
  })
})

describe('injectCspMeta', () => {
  it('meta charset の直後、インラインスクリプトより前に入れる', () => {
    const html =
      '<!DOCTYPE html><html><head><meta charSet="utf-8"/><title>t</title><script>a()</script></head></html>'
    const result = injectCspMeta(html, ["'self'"])
    expect(result).toBe(
      '<!DOCTYPE html><html><head><meta charSet="utf-8"/>' +
        `<meta http-equiv="Content-Security-Policy" content="script-src 'self' ${sha256('a()')}"/>` +
        '<title>t</title><script>a()</script></head></html>'
    )
  })

  it('meta charset がなければ <head> の直後に入れる', () => {
    const result = injectCspMeta('<head lang="ja"><title>t</title></head>', ["'self'"])
    expect(result).toBe(
      `<head lang="ja"><meta http-equiv="Content-Security-Policy" content="script-src 'self'"/><title>t</title></head>`
    )
  })

  it('<head> がなければエラーにする（保護が外れたまま公開しない）', () => {
    expect(() => injectCspMeta('<html><body></body></html>', ["'self'"])).toThrow('<head>')
  })
})
