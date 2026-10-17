import { MAX_DETAIL_BACK_DEPTH, resolveBackTarget, withBackParam } from '../animeListOrigin'

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
    // マイページは表示中のタブを含めて戻す
    expect(resolveBackTarget(encodeURIComponent('/mypage?tab=watching'))).toEqual({
      href: '/mypage?tab=watching',
      label: 'マイページ',
    })
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

describe('関連作品からの遷移（前の作品に戻る）', () => {
  // 今期アニメ → 作品A → 作品B と進んだときの、作品Bの URL に載る from を作る
  const fromOfA = '/anime/10?from=%2Fanime%2Fcurrent'
  const hrefOfB = withBackParam('/anime/20', fromOfA)
  const fromParamOfB = (href: string) => new URLSearchParams(href.split('?')[1]).get('from')

  it('作品詳細からの遷移には、その作品自身の戻り先ごと from に載せる', () => {
    expect(fromParamOfB(hrefOfB)).toBe(fromOfA)
  })

  it('作品詳細の戻り先は「前の作品」として、作品自身の戻り先を保ったまま解決する', () => {
    expect(resolveBackTarget(fromParamOfB(hrefOfB))).toEqual({ href: fromOfA, label: '前の作品' })
    // 作品Aに戻ったあと、さらに今期アニメへ戻れる
    expect(resolveBackTarget(new URLSearchParams(fromOfA.split('?')[1]).get('from')).label).toBe(
      '今期アニメ'
    )
  })

  it('作品自身の戻り先に含まれる & や ? を壊さない', () => {
    const fromOfSearchA = withBackParam('/anime/10', '/search?keyword=a&genre=b')
    const target = resolveBackTarget(fromParamOfB(withBackParam('/anime/20', fromOfSearchA)))
    expect(target.href).toBe(fromOfSearchA)
    expect(new URLSearchParams(target.href.split('?')[1]).get('from')).toBe(
      '/search?keyword=a&genre=b'
    )
  })

  it('末尾スラッシュ付きの作品詳細パスも解決できる', () => {
    expect(resolveBackTarget('/anime/10/').label).toBe('前の作品')
    expect(withBackParam('/anime/20', '/anime/10/')).toBe('/anime/20?from=%2Fanime%2F10%2F')
  })

  it('作品詳細以外のアニメ配下のパスは前の作品として扱わない', () => {
    expect(resolveBackTarget('/anime/abc').label).toBe('ホーム')
  })

  it(`さかのぼれる作品は ${MAX_DETAIL_BACK_DEPTH} 段までで、古い戻り先から捨てる`, () => {
    // 今期アニメ → 作品1 → 作品2 → … と関連作品をたどる
    let from = '/anime/current'
    for (let id = 1; id <= MAX_DETAIL_BACK_DEPTH + 2; id++) {
      const href = withBackParam(`/anime/${id}`, from)
      from = href
    }
    // 最後の作品の詳細から、戻り先を順にたどる
    const chain: string[] = []
    let current: string | null = new URLSearchParams(
      withBackParam('/anime/999', from).split('?')[1]
    ).get('from')
    while (current) {
      const target = resolveBackTarget(current)
      chain.push(target.label)
      if (target.label !== '前の作品') break
      current = new URLSearchParams(target.href.split('?')[1] ?? '').get('from')
    }
    expect(chain).toEqual(Array(MAX_DETAIL_BACK_DEPTH).fill('前の作品'))
  })
})
