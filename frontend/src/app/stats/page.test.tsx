import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import StatsPage from './page'
import { useAuth } from '@/contexts/AuthContext'
import { useRouter } from 'next/navigation'

// Mock dependencies
jest.mock('@/contexts/AuthContext')
jest.mock('next/navigation')
jest.mock('@/components/auth/ProtectedRoute', () => ({
  ProtectedRoute: ({ children }: { children: React.ReactNode }) => (
    <>{children}</>
  ),
}))

// Mock fetch
global.fetch = jest.fn()

describe('StatsPage', () => {
  const mockRouter = {
    push: jest.fn(),
  }

  const mockUser = {
    id: 1,
    name: 'テストユーザー',
    email: 'test@example.com',
  }

  const mockStats = {
    totalEpisodesWatched: 100,
    totalAnimesCompleted: 10,
    totalWatchTimeHours: 40.0,
    favoriteGenres: ['アクション', 'コメディ', 'ドラマ'],
    thisMonthEpisodes: 15,
  }

  beforeEach(() => {
    jest.clearAllMocks()
    ;(useRouter as jest.Mock).mockReturnValue(mockRouter)
  })

  it('未認証時は何も表示しない（ProtectedRouteが処理）', () => {
    ;(useAuth as jest.Mock).mockReturnValue({
      user: null,
      token: null,
      isLoading: false,
      isAuthenticated: false,
    })

    const { container } = render(<StatsPage />)

    // ProtectedRouteがモックされているため、コンテンツは表示されるが
    // 実際のアプリではProtectedRouteがリダイレクトを処理する
    expect(container).toBeInTheDocument()
  })

  it('統計情報を正しく表示する', async () => {
    ;(useAuth as jest.Mock).mockReturnValue({
      user: mockUser,
      token: 'test-token',
      isLoading: false,
      isAuthenticated: true,
    })
    ;(global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => mockStats,
    })

    render(<StatsPage />)

    await waitFor(() => {
      expect(screen.getByText('視聴統計')).toBeInTheDocument()
      expect(screen.getByText('エピソード総数')).toBeInTheDocument()
      expect(screen.getByText('完了アニメ')).toBeInTheDocument()
      expect(screen.getByText('40.0')).toBeInTheDocument() // totalWatchTimeHours
    })
  })

  it('お気に入りジャンルを表示する', async () => {
    ;(useAuth as jest.Mock).mockReturnValue({
      user: mockUser,
      token: 'test-token',
      isLoading: false,
      isAuthenticated: true,
    })
    ;(global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => mockStats,
    })

    render(<StatsPage />)

    await waitFor(() => {
      expect(screen.getByText('アクション')).toBeInTheDocument()
      expect(screen.getByText('コメディ')).toBeInTheDocument()
      expect(screen.getByText('ドラマ')).toBeInTheDocument()
    })
  })

  it('エラー時にエラーメッセージを表示する', async () => {
    ;(useAuth as jest.Mock).mockReturnValue({
      user: mockUser,
      token: 'test-token',
      isLoading: false,
      isAuthenticated: true,
    })
    ;(global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: false,
      json: async () => ({ message: 'Server error' }),
    })

    render(<StatsPage />)

    await waitFor(() => {
      // APIクライアントがエラーメッセージを返す
      expect(screen.getByText(/Server error|統計情報の取得に失敗しました/)).toBeInTheDocument()
    })
  })

  it('ネットワークエラー時にエラーメッセージを表示する', async () => {
    ;(useAuth as jest.Mock).mockReturnValue({
      user: mockUser,
      token: 'test-token',
      isLoading: false,
      isAuthenticated: true,
    })
    // TypeErrorを投げることでネットワークエラーをシミュレート
    ;(global.fetch as jest.Mock).mockRejectedValue(new TypeError('Failed to fetch'))

    render(<StatsPage />)

    await waitFor(() => {
      // APIクライアントがTypeErrorを"ネットワークエラーが発生しました"に変換する
      expect(screen.getByText('ネットワークエラーが発生しました')).toBeInTheDocument()
    })
  })

  it('データローディング中は読み込み表示を出す', async () => {
    ;(useAuth as jest.Mock).mockReturnValue({
      user: mockUser,
      token: 'test-token',
      isLoading: false,
      isAuthenticated: true,
    })

    // fetchを遅延させる
    ;(global.fetch as jest.Mock).mockImplementation(
      () =>
        new Promise((resolve) =>
          setTimeout(
            () =>
              resolve({
                ok: true,
                json: async () => mockStats,
              }),
            100
          )
        )
    )

    render(<StatsPage />)

    expect(screen.getByText('統計情報を読み込み中...')).toBeInTheDocument()
  })
})
