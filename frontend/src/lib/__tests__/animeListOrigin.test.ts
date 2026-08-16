import { resolveBackTarget, withBackParam } from '../animeListOrigin'

describe('resolveBackTarget', () => {
  it('既知の一覧パスをラベル付きで解決する', () => {
    expect(resolveBackTarget('/anime/current')).toEqual({
      href: '/anime/current',
      label: '今期アニメ',
    })
    expect(resolveBackTarget('/anime/popular')).toEqual({
      href: '/anime/popular',
      label: '人気アニメ',
    })
  })

  it('エンコードされた値を復元する', () => {
    expect(resolveBackTarget('%2Fanime%2Fcurrent').label).toBe('今期アニメ')
  })

  it('検索条件をクエリごと保持する', () => {
    expect(resolveBackTarget('/search?keyword=%E5%91%AA%E8%A1%93')).toEqual({
      href: '/search?keyword=呪術',
      label: '検索結果',
    })
  })

  it('from が無ければホームにフォールバックする', () => {
    expect(resolveBackTarget(null)).toEqual({ href: '/', label: 'ホーム' })
    expect(resolveBackTarget('')).toEqual({ href: '/', label: 'ホーム' })
  })

  it('カードを置いている画面をすべて解決できる', () => {
    // 一覧以外にカードを持つ画面。ここが漏れるとホームに戻ってしまう
    expect(resolveBackTarget('/favorites').label).toBe('お気に入り')
    expect(resolveBackTarget('/mypage').label).toBe('マイページ')
  })

  // 本番ビルドのみ trailingSlash が有効なため、開発サーバーでは再現しない差分。
  // ここで両形式を固定しておかないと本番でだけホームに戻る。
  it('末尾スラッシュ付きでも解決できる', () => {
    expect(resolveBackTarget('/anime/current/')).toEqual({
      href: '/anime/current/',
      label: '今期アニメ',
    })
    expect(resolveBackTarget('%2Fanime%2Fpopular%2F').label).toBe('人気アニメ')
    expect(resolveBackTarget('/search/?keyword=abc')).toEqual({
      href: '/search/?keyword=abc',
      label: '検索結果',
    })
  })

  it('未知のパスはホームにフォールバックする', () => {
    expect(resolveBackTarget('/register')).toEqual({ href: '/', label: 'ホーム' })
  })

  it('外部URLへ誘導されない', () => {
    expect(resolveBackTarget('https://evil.example.com').href).toBe('/')
    // プロトコル相対URL
    expect(resolveBackTarget('//evil.example.com').href).toBe('/')
    expect(resolveBackTarget('/\/evil.example.com').href).toBe('/')
  })

  it('不正なパーセントエンコーディングでも例外を投げない', () => {
    expect(resolveBackTarget('%E0%A4%A')).toEqual({ href: '/', label: 'ホーム' })
  })
})

describe('withBackParam', () => {
  it('既知の一覧からの遷移には from を付ける', () => {
    expect(withBackParam('/anime/1', '/anime/current')).toBe(
      '/anime/1?from=%2Fanime%2Fcurrent'
    )
  })

  it('クエリ付きの遷移元をエンコードして載せる', () => {
    expect(withBackParam('/anime/1', '/search?keyword=abc')).toBe(
      '/anime/1?from=%2Fsearch%3Fkeyword%3Dabc'
    )
  })

  it('末尾スラッシュ付きの遷移元にも付ける', () => {
    expect(withBackParam('/anime/1', '/anime/current/')).toBe(
      '/anime/1?from=%2Fanime%2Fcurrent%2F'
    )
  })

  it('ホームなど一覧以外からは付けない', () => {
    expect(withBackParam('/anime/1', '/')).toBe('/anime/1')
    expect(withBackParam('/anime/1', undefined)).toBe('/anime/1')
  })
})
