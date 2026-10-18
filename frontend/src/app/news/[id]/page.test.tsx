import React from 'react'
import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import NewsDetailPage, { generateMetadata, generateStaticParams } from './page'
import { ANNOUNCEMENTS } from '@/constants/announcements'

jest.mock('next/navigation', () => ({
  notFound: jest.fn(() => {
    throw new Error('NEXT_NOT_FOUND')
  }),
}))

const known = ANNOUNCEMENTS[0]

describe('お知らせ詳細ページ', () => {
  it('全件のIDを静的生成の対象にする', () => {
    expect(generateStaticParams()).toEqual(ANNOUNCEMENTS.map((a) => ({ id: a.id })))
  })

  it('日付・カテゴリ・タイトル・本文を出し、一覧への戻りリンクを1つだけ置く', () => {
    render(<NewsDetailPage params={{ id: known.id }} />)

    expect(screen.getByRole('heading', { level: 1, name: known.title })).toBeInTheDocument()
    expect(screen.getByText(known.body)).toBeInTheDocument()
    const back = screen.getByRole('link', { name: /お知らせ一覧/ })
    expect(back).toHaveAttribute('href', '/news')
    expect(screen.getAllByRole('link')).toHaveLength(1)
  })

  it('存在しないIDは404', () => {
    expect(() => render(<NewsDetailPage params={{ id: 'unknown' }} />)).toThrow('NEXT_NOT_FOUND')
  })

  it('titleは「記事タイトル | お知らせ | MyAnimeLogs」', () => {
    expect(generateMetadata({ params: { id: known.id } }).title).toBe(
      `${known.title} | お知らせ | MyAnimeLogs`,
    )
  })
})
