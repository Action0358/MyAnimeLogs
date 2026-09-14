import React from 'react'
import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import ContactPage from './page'
import { SUPPORT_EMAIL, SUPPORT_FORM_URL } from '@/lib/constants'

/**
 * 問い合わせ窓口の出し分けを検証する（#869）。
 *
 * 一般の問い合わせはフォーム、個人情報に関する請求はメール、という
 * 役割分担が崩れると、法的な請求の受付経路が失われる。
 */
describe('お問い合わせページ', () => {
  it('見出しは他ページと同じサイズで出す', () => {
    render(<ContactPage />)

    const heading = screen.getByRole('heading', { name: 'お問い合わせ', level: 1 })
    expect(heading).toHaveClass('text-xl', 'md:text-[26px]')
  })

  it('フォームへの導線を主導線として出す', () => {
    render(<ContactPage />)

    const link = screen.getByRole('link', { name: /お問い合わせフォームを開く/ })
    expect(link).toHaveAttribute('href', SUPPORT_FORM_URL)
  })

  it('フォームは別タブで開き、rel を付ける', () => {
    render(<ContactPage />)

    const link = screen.getByRole('link', { name: /お問い合わせフォームを開く/ })
    expect(link).toHaveAttribute('target', '_blank')
    // 外部サイトへ window.opener を渡さない
    expect(link).toHaveAttribute('rel', 'noopener noreferrer')
  })

  it('ログイン不要であることを明記する', () => {
    render(<ContactPage />)

    expect(screen.getByText(/ログインは不要です/)).toBeInTheDocument()
  })

  it('個人情報に関する請求はメール窓口を残す', () => {
    render(<ContactPage />)

    expect(screen.getByRole('heading', { name: '個人情報に関するご請求' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: SUPPORT_EMAIL })).toHaveAttribute(
      'href',
      `mailto:${SUPPORT_EMAIL}`
    )
  })

  it('プライバシーポリシーへ案内する', () => {
    render(<ContactPage />)

    expect(screen.getByRole('link', { name: 'プライバシーポリシー' })).toHaveAttribute(
      'href',
      '/privacy'
    )
  })
})
