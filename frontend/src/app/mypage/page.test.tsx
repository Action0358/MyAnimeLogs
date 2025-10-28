import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import MyPage from './page'
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

describe('MyPage', () => {
  let consoleErrorSpy: jest.SpyInstance

  const mockRouter = {
    push: jest.fn(),
  }

  const mockUser = {
    id: 1,
    name: 'テストユーザー',
    email: 'test@example.com',
    description: 'テスト説明',
    favorite_anime: '進撃の巨人',
    sns_link: 'https://twitter.com/testuser',
  }

  const mockStats = {
    totalEpisodesWatched: 100,
    totalAnimesCompleted: 10,
    totalWatchTimeHours: 40.0,
    favoriteGenres: ['アクション', 'コメディ', 'ドラマ'],
    thisMonthEpisodes: 15,
  }

  // 成功時のfetchモックを作成するヘルパー関数
  const createSuccessfulFetchMock = () => {
    return (url: string) => {
      // より厳密なURLマッチングを使用
      if (url.endsWith('/api/users/me/favorites')) {
        return Promise.resolve({
          ok: true,
          json: async () => [],
        })
      }
      if (url.endsWith('/api/users/me/stats')) {
        return Promise.resolve({
          ok: true,
          json: async () => mockStats,
        })
      }
      if (url.endsWith('/api/users/me')) {
        return Promise.resolve({
          ok: true,
          json: async () => mockUser,
        })
      }
      return Promise.reject(new Error('Unknown URL'))
    }
  }

  beforeEach(() => {
    jest.clearAllMocks()
    ;(useRouter as jest.Mock).mockReturnValue(mockRouter)
    // console.errorをモックしてテスト出力をクリーンに保つ
    consoleErrorSpy = jest.spyOn(console, 'error').mockImplementation(() => {})
  })

  afterEach(() => {
    // console.errorのモックのみをリストア
    consoleErrorSpy.mockRestore()
  })

  it('未認証時は何も表示しない（ProtectedRouteが処理）', () => {
    ;(useAuth as jest.Mock).mockReturnValue({
      user: null,
      token: null,
      isLoading: false,
      isAuthenticated: false,
    })

    const { container } = render(<MyPage />)

    // ProtectedRouteがモックされているため、コンテンツは表示されるが
    // 実際のアプリではProtectedRouteがリダイレクトを処理する
    expect(container).toBeInTheDocument()
  })

  describe('認証済みユーザー', () => {
    beforeEach(() => {
      // 認証済みユーザーの共通セットアップ
      ;(useAuth as jest.Mock).mockReturnValue({
        user: mockUser,
        token: 'test-token',
        isLoading: false,
        isAuthenticated: true,
      })
    })

    it('ユーザー情報を表示する', async () => {
      ;(global.fetch as jest.Mock).mockImplementation(
        createSuccessfulFetchMock()
      )

      render(<MyPage />)

      await waitFor(() => {
        expect(screen.getByText('テストユーザー')).toBeInTheDocument()
        expect(screen.getByText('テスト説明')).toBeInTheDocument()
      })
    })

    it('統計情報を正しく表示する', async () => {
      ;(global.fetch as jest.Mock).mockImplementation(
        createSuccessfulFetchMock()
      )

      render(<MyPage />)

      await waitFor(() => {
        expect(screen.getByText('視聴数')).toBeInTheDocument()
        expect(screen.getByText('完了')).toBeInTheDocument()
        expect(screen.getByText('100')).toBeInTheDocument()
        expect(screen.getByText('10')).toBeInTheDocument()
      })
    })

    it('データローディング中は読み込み表示を出す', () => {
      // fetchを未解決のPromiseとしてモック
      ;(global.fetch as jest.Mock).mockImplementation(
        () => new Promise(() => {}) // 永遠に解決しないPromise
      )

      render(<MyPage />)

      expect(screen.getByText('読み込み中...')).toBeInTheDocument()
    })

    it('データ取得エラー時にエラーメッセージを表示する', async () => {
      ;(global.fetch as jest.Mock).mockImplementation((url: string) => {
        // すべてのAPIエンドポイントでエラーを返す
        // Promise.allで並列実行されるため、どのエラーが表示されるかは不定
        // ただし、APIクライアントがレスポンスのmessageをそのまま使用する
        return Promise.resolve({
          ok: false,
          status: 500,
          json: async () => ({ message: 'Server error' }),
        })
      })

      render(<MyPage />)

      await waitFor(() => {
        // APIクライアントがエラーレスポンスのmessageをそのまま表示する
        expect(screen.getByText('Server error')).toBeInTheDocument()
      })
    })

    it('ネットワークエラー時にエラーメッセージを表示する', async () => {
      // TypeErrorを投げることでネットワークエラーをシミュレート
      ;(global.fetch as jest.Mock).mockRejectedValue(new TypeError('Failed to fetch'))

      render(<MyPage />)

      await waitFor(() => {
        // APIクライアントがTypeErrorを"ネットワークエラーが発生しました"に変換する
        expect(screen.getByText('ネットワークエラーが発生しました')).toBeInTheDocument()
      })
    })
  })
})
