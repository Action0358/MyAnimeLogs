import React from 'react'
import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import NewsPage from './page'
import { getAnnouncements } from '@/constants/announcements'
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

    expect(screen.getByText('現在お知らせはありません。')).toBeInTheDocument()
    expect(screen.queryByRole('listitem')).not.toBeInTheDocument()
  })
})
