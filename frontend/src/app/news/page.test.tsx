import React from 'react'
import { act, fireEvent, render, screen, within } from '@testing-library/react'
import '@testing-library/jest-dom'
import NewsPage from './page'
import { announcementHref, getAnnouncements } from '@/constants/announcements'
import type { Announcement } from '@/constants/announcements'
import {
  NEWS_FILTER_MIN_COUNT,
  NEWS_PAGE_SIZE,
} from '@/components/announcements/AnnouncementBrowser'

jest.mock('@/constants/announcements', () => {
  const actual = jest.requireActual('@/constants/announcements')
  return { ...actual, getAnnouncements: jest.fn() }
})

const mockReplace = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({ replace: mockReplace }),
}))

const mockGetAnnouncements = getAnnouncements as jest.Mock

// jsdom は scrollIntoView を実装していない
beforeAll(() => {
  Element.prototype.scrollIntoView = jest.fn()
})

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

  it('日付・カテゴリ・タイトル・本文の導入を出す', () => {
    mockGetAnnouncements.mockReturnValue([build()])

    render(<NewsPage />)

    expect(screen.getByText('2026年9月6日')).toBeInTheDocument()
    expect(screen.getByText('リリース')).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'ベータ版を公開しました' })).toBeInTheDocument()
    const preview = screen.getByText(/本日よりベータ版を公開しました/)
    // 全文は詳細で読ませるため、一覧では2行に省略する
    expect(preview).toHaveClass('line-clamp-2')
  })

  it('日付を time 要素で持つ', () => {
    mockGetAnnouncements.mockReturnValue([build()])

    const { container } = render(<NewsPage />)

    expect(container.querySelector('time')).toHaveAttribute('dateTime', '2026-09-06')
  })

  it('任意のlinkは一覧に出さない（行全体がリンクのため。詳細ページに出す）', () => {
    mockGetAnnouncements.mockReturnValue([
      build({ link: { href: '/search', label: '検索を試す' } }),
    ])

    render(<NewsPage />)

    expect(screen.queryByRole('link', { name: '検索を試す' })).not.toBeInTheDocument()
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

  it('各行が詳細ページへのリンクになる', () => {
    mockGetAnnouncements.mockReturnValue([
      build(),
      build({
        id: '2026-08-20-maintenance',
        date: '2026-08-20',
        title: 'メンテナンス',
        body: '作業予定です。',
      }),
    ])

    render(<NewsPage />)

    expect(screen.getByRole('link', { name: /ベータ版を公開しました/ })).toHaveAttribute(
      'href',
      '/news/2026-09-06-beta'
    )
    expect(screen.getByRole('link', { name: /メンテナンス/ })).toHaveAttribute(
      'href',
      '/news/2026-08-20-maintenance'
    )
    expect(announcementHref('custom-id')).toBe('/news/custom-id')
  })
})

/** 新しい順に並んだ n 件。日付は1日ずつ遡り、年をまたがない */
const buildMany = (count: number, overrides: (i: number) => Partial<Announcement> = () => ({})) =>
  Array.from({ length: count }, (_, i) =>
    build({
      id: `n${String(count - i).padStart(2, '0')}`,
      date: `2026-09-${String(28 - i).padStart(2, '0')}`,
      title: `お知らせ${i + 1}`,
      ...overrides(i),
    })
  )

/**
 * 件数が増えたときの受け皿（モックアップ「News Page Redesign」修正版）。
 * 少ないうちは出さず、過剰な装置に見せない。
 */
describe('絞り込みとページ送り', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    window.location.hash = ''
  })

  it('5件以下ではカテゴリタブ・年の区切り・ページ送りを出さない', () => {
    mockGetAnnouncements.mockReturnValue(buildMany(NEWS_FILTER_MIN_COUNT - 1))

    render(<NewsPage />)

    expect(screen.queryByRole('group', { name: 'カテゴリで絞り込む' })).not.toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: '2026年' })).not.toBeInTheDocument()
    expect(screen.queryByRole('navigation', { name: 'ページネーション' })).not.toBeInTheDocument()
  })

  it('6件以上でタブと年の区切りを出す。タブはデータにあるカテゴリだけ', () => {
    mockGetAnnouncements.mockReturnValue(
      buildMany(NEWS_FILTER_MIN_COUNT, (i) => ({ category: i % 2 ? 'maintenance' : 'feature' }))
    )

    render(<NewsPage />)

    const tabs = within(screen.getByRole('group', { name: 'カテゴリで絞り込む' }))
    expect(tabs.getAllByRole('button').map((b) => b.textContent)).toEqual([
      'すべて',
      '機能',
      'メンテナンス',
    ])
    expect(tabs.getByRole('button', { name: 'すべて' })).toHaveAttribute('aria-pressed', 'true')
    expect(screen.getByRole('heading', { name: '2026年' })).toBeInTheDocument()
  })

  it('カテゴリで絞り込むと、項目と件数が絞られる', () => {
    mockGetAnnouncements.mockReturnValue(
      buildMany(8, (i) => ({ category: i < 3 ? 'maintenance' : 'feature' }))
    )

    render(<NewsPage />)
    fireEvent.click(screen.getByRole('button', { name: 'メンテナンス' }))

    expect(screen.getAllByRole('listitem')).toHaveLength(3)
    const heading = screen.getByRole('heading', { name: 'お知らせ', level: 1 })
    expect(heading.parentElement).toHaveTextContent('3件')
  })

  it('年をまたぐと年ごとに区切る', () => {
    mockGetAnnouncements.mockReturnValue([
      ...buildMany(5),
      build({ id: 'old', date: '2025-12-24', title: '去年のお知らせ' }),
    ])

    render(<NewsPage />)

    expect(screen.getByRole('heading', { name: '2026年' })).toBeInTheDocument()
    const lastYear = screen.getByRole('region', { name: '2025年' })
    expect(within(lastYear).getByText('去年のお知らせ')).toBeInTheDocument()
  })

  it(`${NEWS_PAGE_SIZE}件を超えるとページ送りを出し、ページを切り替えられる`, () => {
    mockGetAnnouncements.mockReturnValue(buildMany(NEWS_PAGE_SIZE + 2))

    render(<NewsPage />)

    expect(screen.getAllByRole('listitem')).toHaveLength(NEWS_PAGE_SIZE)
    expect(screen.getByRole('button', { name: '前のページ' })).toBeDisabled()

    fireEvent.click(screen.getByRole('button', { name: '2' }))

    expect(screen.getAllByRole('listitem')).toHaveLength(2)
    expect(screen.getByRole('button', { name: '2' })).toHaveAttribute('aria-current', 'page')
    expect(screen.getByRole('button', { name: '次のページ' })).toBeDisabled()
  })

  it('絞り込みを変えると1ページ目に戻る', () => {
    mockGetAnnouncements.mockReturnValue(buildMany(NEWS_PAGE_SIZE + 2))

    render(<NewsPage />)
    fireEvent.click(screen.getByRole('button', { name: '2' }))
    fireEvent.click(screen.getByRole('button', { name: 'すべて' }))

    expect(screen.getAllByRole('listitem')).toHaveLength(NEWS_PAGE_SIZE)
  })

  /**
   * 詳細ページを追加する前は /news#<id> が個別の着地点だった。
   * すでに共有されたリンクは詳細ページへ転送する。
   */
  it('旧アンカー /news#<id> は該当の詳細ページへ転送する', () => {
    const items = buildMany(3)
    mockGetAnnouncements.mockReturnValue(items)
    window.location.hash = `#${items[1].id}`

    render(<NewsPage />)

    expect(mockReplace).toHaveBeenCalledWith(`/news/${items[1].id}`)
  })

  it('表示後にハッシュが変わったときも転送する', () => {
    const items = buildMany(3)
    mockGetAnnouncements.mockReturnValue(items)

    render(<NewsPage />)
    expect(mockReplace).not.toHaveBeenCalled()

    act(() => {
      window.location.hash = `#${items[0].id}`
      window.dispatchEvent(new HashChangeEvent('hashchange'))
    })

    expect(mockReplace).toHaveBeenCalledWith(`/news/${items[0].id}`)
  })

  it('存在しないIDのハッシュでは転送しない', () => {
    mockGetAnnouncements.mockReturnValue(buildMany(3))
    window.location.hash = '#unknown'

    render(<NewsPage />)

    expect(mockReplace).not.toHaveBeenCalled()
  })
})
