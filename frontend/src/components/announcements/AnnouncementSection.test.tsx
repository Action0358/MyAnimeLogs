import React from 'react'
import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import { AnnouncementSection } from './AnnouncementSection'
import { getAnnouncements, HOME_ANNOUNCEMENT_COUNT } from '@/constants/announcements'
import type { Announcement } from '@/constants/announcements'

// お知らせの中身はリポジトリの定数なので、件数に依存しないようここで差し替える
jest.mock('@/constants/announcements', () => {
  const actual = jest.requireActual('@/constants/announcements')
  return { ...actual, getAnnouncements: jest.fn() }
})

const mockGetAnnouncements = getAnnouncements as jest.Mock

const build = (id: string, date: string, title: string): Announcement => ({
  id,
  date,
  category: 'release',
  title,
  body: '本文',
})

describe('トップページのお知らせセクション', () => {
  beforeEach(() => jest.clearAllMocks())

  it('お知らせが0件ならセクションごと描画しない', () => {
    mockGetAnnouncements.mockReturnValue([])

    const { container } = render(<AnnouncementSection />)

    expect(container).toBeEmptyDOMElement()
  })

  it('見出しと「もっと見る」を出す', () => {
    mockGetAnnouncements.mockReturnValue([build('a', '2026-09-06', 'ベータ版を公開しました')])

    render(<AnnouncementSection />)

    expect(screen.getByRole('heading', { name: 'お知らせ' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'もっと見る' })).toHaveAttribute('href', '/news')
  })

  it('上限件数までしか出さない', () => {
    mockGetAnnouncements.mockReturnValue(
      Array.from({ length: HOME_ANNOUNCEMENT_COUNT + 2 }, (_, i) =>
        build(`a${i}`, '2026-09-06', `お知らせ${i}`),
      ),
    )

    render(<AnnouncementSection />)

    expect(screen.getAllByRole('listitem')).toHaveLength(HOME_ANNOUNCEMENT_COUNT)
  })

  it('日付・カテゴリ・タイトルを出し、/news へ送る', () => {
    mockGetAnnouncements.mockReturnValue([build('a', '2026-09-06', 'ベータ版を公開しました')])

    render(<AnnouncementSection />)

    expect(screen.getByText('2026年9月6日')).toBeInTheDocument()
    expect(screen.getByText('リリース')).toBeInTheDocument()
    expect(screen.getByText('ベータ版を公開しました')).toBeInTheDocument()
    // 詳細ページは持たないため、項目のリンク先も一覧ページ
    expect(screen.getByRole('link', { name: /ベータ版を公開しました/ })).toHaveAttribute(
      'href',
      '/news',
    )
  })

  it('本文はトップには出さない', () => {
    mockGetAnnouncements.mockReturnValue([build('a', '2026-09-06', 'タイトル')])

    render(<AnnouncementSection />)

    expect(screen.queryByText('本文')).not.toBeInTheDocument()
  })
})
