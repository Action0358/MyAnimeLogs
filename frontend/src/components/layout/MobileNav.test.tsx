import React from 'react'
import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import { MobileNav } from './MobileNav'
import { useAuth } from '@/contexts/AuthContext'

jest.mock('@/contexts/AuthContext', () => ({ useAuth: jest.fn() }))
jest.mock('next/navigation', () => ({
  useRouter: () => ({ push: jest.fn(), replace: jest.fn() }),
  usePathname: () => '/',
  useSearchParams: () => new URLSearchParams(),
}))

const mockUseAuth = useAuth as jest.Mock

const renderNav = () =>
  render(<MobileNav isOpen onClose={jest.fn()} triggerRef={{ current: null }} />)

const USER = { id: 1, name: 'テスト', email: 'test@example.com' }

// 確認中に出し分けると、ログイン済みの人に「ログイン」が、未ログインの人に「ログアウト」が出る
describe('MobileNav のログイン / ログアウトの出し分け', () => {
  it('認証状態の確認中は、どちらも出さない', () => {
    mockUseAuth.mockReturnValue({ user: null, isLoading: true, logout: jest.fn() })

    renderNav()

    expect(screen.queryByRole('link', { name: 'ログイン' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'ログアウト' })).not.toBeInTheDocument()
  })

  it('未ログインなら「ログイン」を出す', () => {
    mockUseAuth.mockReturnValue({ user: null, isLoading: false, logout: jest.fn() })

    renderNav()

    expect(screen.getByRole('link', { name: 'ログイン' })).toBeInTheDocument()
  })

  it('ログイン済みなら「ログアウト」を出す', () => {
    mockUseAuth.mockReturnValue({ user: USER, isLoading: false, logout: jest.fn() })

    renderNav()

    expect(screen.getByRole('button', { name: 'ログアウト' })).toBeInTheDocument()
  })
})
