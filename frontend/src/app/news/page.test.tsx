import React from 'react'
import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import NewsPage from './page'
import { announcementHref, getAnnouncements } from '@/constants/announcements'
import type { Announcement } from '@/constants/announcements'

jest.mock('@/constants/announcements', () => {
  const actual = jest.requireActual('@/constants/announcements')
  return { ...actual, getAnnouncements: jest.fn() }
})

const mockGetAnnouncements = getAnnouncements as jest.Mock

const build = (overrides: Partial<Announcement> = {}): Announcement => ({
  id: '2026-09-06-beta',
  date: '2026-09-06',
  category: 'release',
  title: 'ベータ版を公開しました',
  body: '本日よりベータ版を公開しました。\nご意見をお待ちしています。',
  ...overrides,
})

describe('お知らせ一覧ページ', () => {
  beforeEach(() => jest.clearAllMocks())

  it('件数を見出しの隣に出す', () => {
    mockGetAnnouncements.mockReturnValue([build(), build({ id: 'b', date: '2026-08-20' })])

    render(<NewsPage />)

    const heading = screen.getByRole('heading', { name: 'お知らせ', level: 1 })
    expect(heading.parentElement).toHaveTextContent('2')
    expect(heading.parentElement).toHaveTextContent('件')
  })

  it('日付・カテゴリ・タイトル・本文をその場に出す', () => {
    mockGetAnnouncements.mockReturnValue([build()])

    render(<NewsPage />)

    expect(screen.getByText('2026年9月6日')).toBeInTheDocument()
    expect(screen.getByText('リリース')).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'ベータ版を公開しました' })).toBeInTheDocument()
    // 詳細ページを持たないので、本文がここに出ることが要件
    expect(screen.getByText(/本日よりベータ版を公開しました/)).toBeInTheDocument()
  })

  it('日付を time 要素で持つ', () => {
    mockGetAnnouncements.mockReturnValue([build()])

    const { container } = render(<NewsPage />)

    expect(container.querySelector('time')).toHaveAttribute('dateTime', '2026-09-06')
  })

  it('リンクつきのお知らせは導線を出す', () => {
    mockGetAnnouncements.mockReturnValue([
      build({ link: { href: '/search', label: '検索を試す' } }),
    ])

    render(<NewsPage />)

    expect(screen.getByRole('link', { name: '検索を試す' })).toHaveAttribute('href', '/search')
  })

  it('外部リンクは別タブで開く', () => {
    mockGetAnnouncements.mockReturnValue([
      build({ link: { href: 'https://example.com', label: '詳しくはこちら' } }),
    ])

    render(<NewsPage />)

    const link = screen.getByRole('link', { name: '詳しくはこちら' })
    expect(link).toHaveAttribute('target', '_blank')
    expect(link).toHaveAttribute('rel', expect.stringContaining('noopener'))
  })

  it('0件なら空の案内を出す', () => {
    mockGetAnnouncements.mockReturnValue([])

    render(<NewsPage />)

    expect(screen.getByText('現在お知らせはありません')).toBeInTheDocument()
    expect(screen.queryByRole('listitem')).not.toBeInTheDocument()
  })

  it('複数件でも1つのカードにまとめて表示する', () => {
    mockGetAnnouncements.mockReturnValue([
      build({ id: 'a', date: '2026-09-06' }),
      build({ id: 'b', date: '2026-08-20', title: '定期メンテナンスのお知らせ' }),
      build({ id: 'c', date: '2026-08-01', title: '検索にジャンル絞り込みを追加しました' }),
    ])

    render(<NewsPage />)

    expect(screen.getAllByRole('listitem')).toHaveLength(3)
  })

  /**
   * モックアップ（News Page Redesign）は行全体をクリック可能にしホバー＋
   * シェブロンで反応させているが、詳細ページを持たないこのプロジェクトでは
   * 遷移先のない行を押せるように見せると「空振り」になる。link指定のない
   * 項目は行全体をリンクにしないことを固定する。
   */
  it('linkを指定していない項目は行全体をリンクにしない', () => {
    mockGetAnnouncements.mockReturnValue([build()])

    render(<NewsPage />)

    expect(screen.queryByRole('link')).not.toBeInTheDocument()
  })
})

/**
 * 1件ごとの詳細ページは持たない代わりに、/news#<id> で個別に指せるようにする。
 * トップのお知らせ行や外部からの共有リンクがここへ着地する。
 */
describe('お知らせのアンカー', () => {
  beforeEach(() => jest.clearAllMocks())

  it('各項目に id を振る', () => {
    mockGetAnnouncements.mockReturnValue([
      build(),
      build({ id: '2026-08-20-maintenance', date: '2026-08-20', title: 'メンテナンス' }),
    ])

    const { container } = render(<NewsPage />)

    // id セレクタは数字始まりだと不正になるため getElementById で取る
    const first = container.ownerDocument.getElementById('2026-09-06-beta')
    const second = container.ownerDocument.getElementById('2026-08-20-maintenance')

    expect(first).not.toBeNull()
    expect(second).not.toBeNull()
    expect(first).toHaveTextContent('ベータ版を公開しました')
    expect(second).toHaveTextContent('メンテナンス')
  })

  it('id はデータのIDをそのまま使う（リンク側と一致させる）', () => {
    mockGetAnnouncements.mockReturnValue([build({ id: 'custom-id' })])

    const { container } = render(<NewsPage />)

    expect(container.ownerDocument.getElementById('custom-id')).not.toBeNull()
    expect(announcementHref('custom-id')).toBe('/news#custom-id')
  })

  it('アンカーで指された項目が分かるようにする', () => {
    mockGetAnnouncements.mockReturnValue([build()])

    const { container } = render(<NewsPage />)
    const item = container.ownerDocument.getElementById('2026-09-06-beta')

    // :target で背景を変える。枠線だと高さが変わって他の項目が動く
    expect(item).toHaveClass('target:bg-primary/5')
    // ヘッダー直下に貼り付かないよう余白を残す
    expect(item).toHaveClass('scroll-mt-6')
  })

  it('IDが重複していない（重複するとアンカーが先頭の項目に固定される）', () => {
    mockGetAnnouncements.mockReturnValue([
      build({ id: 'a' }),
      build({ id: 'b', date: '2026-08-20' }),
      build({ id: 'c', date: '2026-08-01' }),
    ])

    const { container } = render(<NewsPage />)
    // tsconfig の target が es5 のため NodeList はスプレッドできない。Array.from を使う
    const ids = Array.from(container.querySelectorAll('li[id]')).map((el) => el.id)

    expect(ids).toEqual(['a', 'b', 'c'])
    expect(new Set(ids).size).toBe(ids.length)
  })
})
