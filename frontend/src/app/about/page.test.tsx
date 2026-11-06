import React from 'react'
import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import AboutPage from './page'
import { useAuth } from '@/contexts/AuthContext'

jest.mock('@/contexts/AuthContext', () => ({ useAuth: jest.fn() }))
jest.mock('next/navigation', () => ({
  useRouter: () => ({ push: jest.fn(), replace: jest.fn() }),
  usePathname: () => '/about',
  useSearchParams: () => new URLSearchParams(),
}))

jest.mock('@/lib/queries/currentSeasonAnime', () => ({
  currentSeasonAnimeQuery: { queryKey: ['current-season'], queryFn: async () => ({ animes: [] }) },
}))

const mockUseAuth = useAuth as jest.Mock

const renderAbout = () =>
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <AboutPage />
    </QueryClientProvider>
  )

const CTA = '今日から、記録をはじめる。'

// 確認中に出すと、ログイン済みの人にも先に出て消え、レイアウトがずれる
describe('サイト紹介の「記録をはじめる」', () => {
  it('認証状態の確認中は出さない', () => {
    mockUseAuth.mockReturnValue({ isAuthenticated: false, isLoading: true })

    renderAbout()

    expect(screen.queryByText(CTA)).not.toBeInTheDocument()
  })

  it('未ログインなら、確認が終わったあとに出す', () => {
    mockUseAuth.mockReturnValue({ isAuthenticated: false, isLoading: false })

    renderAbout()

    expect(screen.getByText(CTA)).toBeInTheDocument()
  })

  it('ログイン済みなら出さない', () => {
    mockUseAuth.mockReturnValue({ isAuthenticated: true, isLoading: false })

    renderAbout()

    expect(screen.queryByText(CTA)).not.toBeInTheDocument()
  })
})
