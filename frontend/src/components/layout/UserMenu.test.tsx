import React from 'react'
import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import { UserMenu } from './UserMenu'
import { useAuth } from '@/contexts/AuthContext'

jest.mock('@/contexts/AuthContext')

const mockUseAuth = useAuth as jest.Mock

/**
 * ユーザーメニュー（ヘッダー右上のドロップダウン）の検証。
 *
 * モックアップ「User Menu Redesign」で、アバター＋名前のヘッダー・項目アイコン・
 * ログアウトの分離・開閉状態の3点が修正対象になったため、それぞれ崩れていないかを見る。
 */
describe('UserMenu', () => {
  beforeEach(() => {
    mockUseAuth.mockReturnValue({
      user: { name: 'Daiyu', avatar_url: null },
      logout: jest.fn(),
    })
  })

  it('ヘッダー部にアバターと名前を表示する', () => {
    render(<UserMenu isOpen onClose={jest.fn()} />)

    expect(screen.getByText('Daiyu')).toBeInTheDocument()
    // avatar_url が無いときはイニシャルで代替する
    expect(screen.getByText('D')).toBeInTheDocument()
  })

  it('マイページ・お気に入りの順でアイコン付きリンクを出す', () => {
    render(<UserMenu isOpen onClose={jest.fn()} />)

    const items = screen.getAllByRole('menuitem')
    // 末尾のログアウトを除く2項目
    expect(items[0]).toHaveTextContent('マイページ')
    expect(items[0].querySelector('svg')).toBeInTheDocument()
    expect(items[1]).toHaveTextContent('お気に入り')
    expect(items[1].querySelector('svg')).toBeInTheDocument()
  })

  it('ログアウトは他の項目と別の見た目（赤系）で分離する', () => {
    render(<UserMenu isOpen onClose={jest.fn()} />)

    const logout = screen.getByRole('menuitem', { name: /ログアウト/ })
    expect(logout.tagName).toBe('BUTTON')
    expect(logout).toHaveClass('text-destructive')
  })

  it('閉じている間はaria-hiddenにし、クリックを透過させる', () => {
    render(<UserMenu isOpen={false} onClose={jest.fn()} />)

    const menu = screen.getByRole('menu', { hidden: true })
    expect(menu).toHaveAttribute('aria-hidden', 'true')
    expect(menu).toHaveClass('pointer-events-none')
  })
})
