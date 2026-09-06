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

const buildMany = (count: number) =>
  Array.from({ length: count }, (_, i) => build(`a${i}`, '2026-09-06', `お知らせ${i}`))

describe('トップページのお知らせセクション', () => {
  beforeEach(() => jest.clearAllMocks())

  it('見出しを出す', () => {
    mockGetAnnouncements.mockReturnValue([build('a', '2026-09-06', 'ベータ版を公開しました')])

    render(<AnnouncementSection />)

    expect(screen.getByRole('heading', { name: 'お知らせ' })).toBeInTheDocument()
  })

  it('日付・カテゴリ・タイトルを出し、/news へ送る', () => {
    mockGetAnnouncements.mockReturnValue([build('a', '2026-09-06', 'ベータ版を公開しました')])

    render(<AnnouncementSection />)

    expect(screen.getByText('2026年9月6日')).toBeInTheDocument()
    expect(screen.getByText('リリース')).toBeInTheDocument()
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

  it('上限件数までしか出さない', () => {
    mockGetAnnouncements.mockReturnValue(buildMany(HOME_ANNOUNCEMENT_COUNT + 2))

    render(<AnnouncementSection />)

    expect(screen.getAllByRole('listitem')).toHaveLength(HOME_ANNOUNCEMENT_COUNT)
  })

  /**
   * 表示件数で収まりきる件数しかないのに「もっと見る」を置くと、
   * 遷移先に続きがある期待を与えてしまう。
   */
  describe('「もっと見る」の出し分け', () => {
    it('表示しきれない分があるときだけ出す', () => {
      mockGetAnnouncements.mockReturnValue(buildMany(HOME_ANNOUNCEMENT_COUNT + 1))

      render(<AnnouncementSection />)

      expect(screen.getByRole('link', { name: /もっと見る/ })).toHaveAttribute('href', '/news')
    })

    it('ちょうど表示件数に収まるときは出さない', () => {
      mockGetAnnouncements.mockReturnValue(buildMany(HOME_ANNOUNCEMENT_COUNT))

      render(<AnnouncementSection />)

      expect(screen.queryByRole('link', { name: /もっと見る/ })).not.toBeInTheDocument()
    })

    it('1件だけのときは出さない', () => {
      mockGetAnnouncements.mockReturnValue([build('a', '2026-09-06', 'タイトル')])

      render(<AnnouncementSection />)

      expect(screen.queryByRole('link', { name: /もっと見る/ })).not.toBeInTheDocument()
    })
  })

  describe('0件のとき', () => {
    it('セクションは残したまま、お知らせがない旨を出す', () => {
      mockGetAnnouncements.mockReturnValue([])

      render(<AnnouncementSection />)

      expect(screen.getByRole('heading', { name: 'お知らせ' })).toBeInTheDocument()
      expect(screen.getByText('現在お知らせはありません')).toBeInTheDocument()
      expect(screen.queryByRole('listitem')).not.toBeInTheDocument()
    })

    it('「もっと見る」は出さない', () => {
      mockGetAnnouncements.mockReturnValue([])

      render(<AnnouncementSection />)

      expect(screen.queryByRole('link', { name: /もっと見る/ })).not.toBeInTheDocument()
    })
  })
})
