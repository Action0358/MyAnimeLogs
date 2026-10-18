import React from 'react'
import { render, screen, within } from '@testing-library/react'
import '@testing-library/jest-dom'
import { AnnouncementDetail } from './AnnouncementDetail'
import type { Announcement } from '@/constants/announcements'

const build = (overrides: Partial<Announcement> = {}): Announcement => ({
  id: 'a',
  date: '2026-08-28',
  category: 'feature',
  title: '視聴記録のエクスポート機能を追加しました',
  body: '導入文です。',
  ...overrides,
})

describe('AnnouncementDetail', () => {
  it('contentが無ければbodyを段落として出す', () => {
    render(<AnnouncementDetail announcement={build()} />)

    expect(screen.getByText('導入文です。')).toBeInTheDocument()
    expect(screen.getByText('2026年8月28日')).toHaveAttribute('datetime', '2026-08-28')
  })

  it('contentの段落・見出し・リスト・本文内リンクを描画する', () => {
    render(
      <AnnouncementDetail
        announcement={build({
          content: [
            { type: 'p', children: ['ダウンロードできます。'] },
            { type: 'heading', text: '使い方' },
            { type: 'ul', items: [['ひとつめ'], ['ふたつめ']] },
            { type: 'p', children: [{ text: 'マイページ', href: '/mypage' }, 'から選びます。'] },
          ],
        })}
      />,
    )

    expect(screen.getByRole('heading', { level: 2, name: '使い方' })).toBeInTheDocument()
    expect(within(screen.getByRole('list')).getAllByRole('listitem')).toHaveLength(2)
    expect(screen.getByRole('link', { name: 'マイページ' })).toHaveAttribute('href', '/mypage')
  })

  it('任意のlinkは本文の後に出し、外部リンクは別タブで開く', () => {
    render(
      <AnnouncementDetail
        announcement={build({ link: { href: 'https://example.com', label: '詳しくはこちら' } })}
      />,
    )

    const link = screen.getByRole('link', { name: '詳しくはこちら' })
    expect(link).toHaveAttribute('target', '_blank')
    expect(link).toHaveAttribute('rel', expect.stringContaining('noopener'))
  })
})
