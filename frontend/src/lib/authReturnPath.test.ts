import {
  sanitizeReturnPath,
  buildSignInHref,
  saveReturnPath,
  consumeReturnPath,
} from './authReturnPath'

describe('sanitizeReturnPath', () => {
  it.each(['/anime/123', '/anime/123?from=current', '/search?q=%E3%81%82', '/mypage', '/news#2026-10-01-release'])(
    'サイト内のパス %s はそのまま通す',
    (path) => {
      expect(sanitizeReturnPath(path)).toBe(path)
    },
  )

  // 外部サイトへ飛ばすオープンリダイレクトを防ぐ
  it.each([
    'https://evil.example.com',
    '//evil.example.com',
    '/\\evil.example.com',
    '/\t/evil.example.com',
    'javascript:alert(1)',
    'anime/123',
  ])('外部や不正な値 %s は拒否する', (path) => {
    expect(sanitizeReturnPath(path)).toBeNull()
  })

  // 戻っても意味がない、またはログイン画面へ戻ってループするパス
  it.each(['/', '/signin', '/signin?reason=timeout', '/login', '/register', '/auth/callback', '/account-deleted'])(
    '%s は戻り先にしない',
    (path) => {
      expect(sanitizeReturnPath(path)).toBeNull()
    },
  )

  it('除外パスと前方一致するだけの別ページは通す', () => {
    expect(sanitizeReturnPath('/authors')).toBe('/authors')
  })

  it.each([null, undefined, ''])('空の値 %p は null', (value) => {
    expect(sanitizeReturnPath(value)).toBeNull()
  })
})

describe('buildSignInHref', () => {
  it('戻り先をエンコードして next に載せる', () => {
    expect(buildSignInHref('/anime/123?from=current')).toBe(
      '/signin?next=%2Fanime%2F123%3Ffrom%3Dcurrent',
    )
  })

  it('戻り先が無効なら素の /signin', () => {
    expect(buildSignInHref('/')).toBe('/signin')
    expect(buildSignInHref('//evil.example.com')).toBe('/signin')
  })
})

describe('saveReturnPath / consumeReturnPath', () => {
  beforeEach(() => sessionStorage.clear())

  it('退避した戻り先を1度だけ取り出せる', () => {
    saveReturnPath('/anime/123')
    expect(consumeReturnPath()).toBe('/anime/123')
    expect(consumeReturnPath()).toBe('/')
  })

  it('戻り先が無いときはトップ', () => {
    expect(consumeReturnPath()).toBe('/')
  })

  // 途中でやめた前回のログインの戻り先を持ち越さない
  it('無効な戻り先で保存すると前回の値を消す', () => {
    saveReturnPath('/anime/1')
    saveReturnPath(null)
    expect(consumeReturnPath()).toBe('/')
  })

  // sessionStorage は書き換えられうるため、取り出し時にも検証する
  it('ストレージに不正な値が入っていてもトップへ戻す', () => {
    sessionStorage.setItem('auth_return_path', '//evil.example.com')
    expect(consumeReturnPath()).toBe('/')
  })
})
