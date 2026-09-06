import fs from 'fs'
import path from 'path'
import { onRequest, STATIC_ANIME_PATHS } from './[id].js'

/**
 * /anime/ 配下のリライトを検証する（#842）。
 *
 * この関数は /anime/ 配下の単一セグメントすべてにマッチするため、
 * 除外を忘れると /anime/popular/ /anime/current/ が詳細ページのシェルへ
 * リライトされ、ローディング表示から抜けられなくなる。
 * サイト内リンクでは再現しないため、テストで固定する。
 */
// jsdom には fetch API の Request がないため、URLだけ持つ最小限のスタブを使う。
// Pages Function は new Request(url, request) で元のリクエストを引き継ぐだけなので、
// 検証に必要なのは最終的なURL。
class RequestStub {
  constructor(input) {
    // 文字列・URLインスタンス・別のRequestのいずれで渡されてもURLを取り出す
    if (typeof input === 'string') this.url = input
    else if (input instanceof URL) this.url = input.href
    else this.url = input.url
  }
}

describe('/anime/ 配下のリライト', () => {
  const originalRequest = global.Request

  beforeAll(() => {
    global.Request = RequestStub
  })

  afterAll(() => {
    global.Request = originalRequest
  })

  const callWith = async (pathname) => {
    const next = jest.fn().mockResolvedValue('NEXT')
    const assetsFetch = jest.fn().mockResolvedValue('SHELL')

    const result = await onRequest({
      request: new Request(`https://myanimelogs.com${pathname}`),
      env: { ASSETS: { fetch: assetsFetch } },
      next,
    })

    return {
      result,
      next,
      assetsFetch,
      rewrittenTo: assetsFetch.mock.calls[0]
        ? new URL(assetsFetch.mock.calls[0][0].url).pathname
        : null,
    }
  }

  describe('静的ページはリライトしない', () => {
    // 末尾スラッシュあり／なしの両方で来る（next.config は trailingSlash: true）
    const staticPaths = STATIC_ANIME_PATHS.flatMap((p) => [`/anime/${p}`, `/anime/${p}/`])

    it.each(staticPaths)('%s はそのまま通す', async (pathname) => {
      const { result, next, assetsFetch } = await callWith(pathname)

      expect(next).toHaveBeenCalled()
      expect(assetsFetch).not.toHaveBeenCalled()
      expect(result).toBe('NEXT')
    })
  })

  describe('作品IDは詳細シェルへリライトする', () => {
    it.each(['/anime/12345', '/anime/12345/', '/anime/1'])(
      '%s は /anime/_/ へ',
      async (pathname) => {
        const { result, next, rewrittenTo } = await callWith(pathname)

        expect(rewrittenTo).toBe('/anime/_/')
        expect(next).not.toHaveBeenCalled()
        expect(result).toBe('SHELL')
      }
    )

    it('クエリ文字列を保ったままリライトする', async () => {
      const { assetsFetch } = await callWith('/anime/12345/?from=/search')

      // 詳細ページは ?from= で「戻る」先を決めるため、失うと導線が壊れる
      const rewritten = new URL(assetsFetch.mock.calls[0][0].url)
      expect(rewritten.pathname).toBe('/anime/_/')
      expect(rewritten.searchParams.get('from')).toBe('/search')
    })
  })

  it('詳細シェル自身は無限にリライトしない', async () => {
    for (const pathname of ['/anime/_', '/anime/_/']) {
      const { next, assetsFetch } = await callWith(pathname)
      expect(next).toHaveBeenCalled()
      expect(assetsFetch).not.toHaveBeenCalled()
    }
  })

  /**
   * 除外リストの追随漏れを防ぐ。src/app/anime/ に静的ページを足したのに
   * STATIC_ANIME_PATHS へ追加し忘れると #842 が再発する。
   */
  it('src/app/anime/ の静的ページがすべて除外リストに入っている', () => {
    const appAnimeDir = path.join(__dirname, '../../src/app/anime')

    const staticDirs = fs
      .readdirSync(appAnimeDir, { withFileTypes: true })
      .filter((entry) => entry.isDirectory())
      // [id] は動的ルートなのでリライト対象
      .filter((entry) => !entry.name.startsWith('['))
      .map((entry) => entry.name)

    expect(staticDirs.length).toBeGreaterThan(0)
    expect(STATIC_ANIME_PATHS.slice().sort()).toEqual(staticDirs.sort())
  })
})
