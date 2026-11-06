import fs from 'fs'
import path from 'path'
import { ALLOWED_IMAGE_HOSTS, isAllowedImageUrl, pickAllowedImageUrl } from '../allowedImageUrl'

describe('isAllowedImageUrl', () => {
  it('許可されたホストの https URL は true', () => {
    expect(isAllowedImageUrl('https://cdn.myanimelist.net/images/anime/1/1.jpg')).toBe(true)
    expect(isAllowedImageUrl('https://image.tmdb.org/t/p/w500/a.jpg')).toBe(true)
  })

  it('ワイルドカードはサブドメインに一致する', () => {
    expect(isAllowedImageUrl('https://s4.anilist.co/file/anilistcdn/media/a.jpg')).toBe(true)
    expect(isAllowedImageUrl('https://pbs.twimg.com/a.jpg')).toBe(true)
  })

  it('ワイルドカードはドメインそのものや、末尾が同じだけの別ドメインには一致しない', () => {
    expect(isAllowedImageUrl('https://anilist.co/a.jpg')).toBe(false)
    expect(isAllowedImageUrl('https://evilanilist.co/a.jpg')).toBe(false)
  })

  it('許可されていないホストは false', () => {
    expect(isAllowedImageUrl('https://sekiro-anime.jp/og-image4.jpg')).toBe(false)
    expect(isAllowedImageUrl('https://poleprincess.jp/img/ogp.jpg')).toBe(false)
    expect(isAllowedImageUrl('https://cdn.myanimelist.net.evil.example/a.jpg')).toBe(false)
  })

  it('https 以外・解釈できない値・空は false', () => {
    expect(isAllowedImageUrl('http://cdn.myanimelist.net/a.jpg')).toBe(false)
    expect(isAllowedImageUrl('data:image/png;base64,AAAA')).toBe(false)
    expect(isAllowedImageUrl('/images/a.jpg')).toBe(false)
    expect(isAllowedImageUrl('not a url')).toBe(false)
    expect(isAllowedImageUrl('')).toBe(false)
    expect(isAllowedImageUrl(null)).toBe(false)
    expect(isAllowedImageUrl(undefined)).toBe(false)
  })
})

describe('pickAllowedImageUrl', () => {
  it('許可される最初の候補を返す', () => {
    expect(
      pickAllowedImageUrl(
        null,
        'https://s4.anilist.co/large.jpg',
        'https://s4.anilist.co/medium.jpg'
      )
    ).toBe('https://s4.anilist.co/large.jpg')
  })

  it('許可されない候補は飛ばす', () => {
    expect(
      pickAllowedImageUrl('https://sekiro-anime.jp/og.jpg', 'https://cdn.myanimelist.net/a.jpg')
    ).toBe('https://cdn.myanimelist.net/a.jpg')
  })

  it('許可される候補がなければ undefined', () => {
    expect(pickAllowedImageUrl(undefined, null, 'https://sekiro-anime.jp/og.jpg')).toBeUndefined()
    expect(pickAllowedImageUrl()).toBeUndefined()
  })
})

describe('ALLOWED_IMAGE_HOSTS と _headers の img-src の一致', () => {
  const publicDir = path.join(__dirname, '..', '..', '..', 'public')

  it.each(['_headers', '_headers.staging', '_headers.production'])('%s', (file) => {
    const content = fs.readFileSync(path.join(publicDir, file), 'utf8')
    const imgSrc = content.match(/img-src ([^;\n]+)/)?.[1]
    expect(imgSrc).toBeDefined()

    const hosts = imgSrc!
      .trim()
      .split(/\s+/)
      .filter((source) => source.startsWith('https://'))
      .map((source) => source.replace('https://', ''))
      // API のドメインは環境ごとに違い、アニメ画像には使わないため比較から除く
      .filter((host) => !host.endsWith('.fly.dev'))

    expect([...hosts].sort()).toEqual([...ALLOWED_IMAGE_HOSTS].sort())
  })
})
