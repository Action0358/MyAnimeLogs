import React from 'react'
import { act, render, screen } from '@testing-library/react'
import Header from './Header'
import Footer from './Footer'
import { resetMaintenance, setMaintenance } from '@/lib/maintenance/maintenanceStore'

jest.mock('next/navigation', () => ({
  usePathname: () => '/news',
  useRouter: () => ({ push: jest.fn(), replace: jest.fn() }),
  useSearchParams: () => new URLSearchParams(),
}))
jest.mock('@/contexts/AuthContext', () => ({
  useAuth: () => ({ user: null, isLoading: false }),
}))
jest.mock('next/image', () => ({
  __esModule: true,
  default: ({ alt }: { alt: string }) => <span role="img" aria-label={alt} />,
}))

describe('メンテナンス中の Header / Footer', () => {
  afterEach(() => resetMaintenance())

  it('通常時はナビ・ログイン・フッターリンクを出す', () => {
    render(
      <>
        <Header />
        <Footer />
      </>
    )
    expect(screen.getAllByText('ログイン').length).toBeGreaterThan(0)
    expect(screen.getAllByText('お問い合わせ').length).toBeGreaterThan(0)
  })

  it('通常時の HTML には、エッジの目印で CSS が出し分けるメンテナンス版も並べる', () => {
    const { container } = render(
      <>
        <Header />
        <Footer />
      </>
    )
    const tagsOf = (when: string) =>
      Array.from(container.querySelectorAll(`[data-show-when="${when}"]`)).map((el) => el.tagName)

    expect(tagsOf('normal')).toEqual(['HEADER', 'FOOTER'])
    expect(tagsOf('maintenance')).toEqual(['HEADER', 'FOOTER'])
    // 通常版のナビ・ログインは、normal 側にだけある
    const normalHeader = container.querySelector('header[data-show-when="normal"]')
    expect(normalHeader).toHaveTextContent('ログイン')
    expect(container.querySelector('header[data-show-when="maintenance"]')).not.toHaveTextContent(
      'ログイン'
    )
  })

  it('メンテナンス中はナビ・ログイン・フッターリンクを隠す', () => {
    render(
      <>
        <Header />
        <Footer />
      </>
    )
    act(() => setMaintenance({}))

    expect(screen.queryByText('ログイン')).not.toBeInTheDocument()
    expect(screen.queryByText('お問い合わせ')).not.toBeInTheDocument()
    expect(screen.queryByText('利用規約')).not.toBeInTheDocument()
    expect(screen.getByText(/All rights reserved/)).toBeInTheDocument()
    // 検出後はメンテナンス版だけを、CSS の出し分けなしで描画する
    expect(document.querySelector('[data-show-when]')).toBeNull()
  })
})
