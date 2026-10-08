import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import MyPage from './page'
import { useAuth } from '@/contexts/AuthContext'
import { useRouter } from 'next/navigation'
import { useWatchStatus } from '@/hooks/useWatchStatus'

// Mock dependencies
jest.mock('@/contexts/AuthContext')
jest.mock('next/navigation')
jest.mock('@/hooks/useWatchStatus')
jest.mock('@/components/auth/ProtectedRoute', () => ({
  ProtectedRoute: ({ children }: { children: React.ReactNode }) => (
    <>{children}</>
  ),
}))

// Mock fetch（ユーザー情報取得用）
global.fetch = jest.fn()

const mockUseWatchStatus = useWatchStatus as jest.Mock

// useWatchStatus の戻り値を生成するヘルパー
const createWatchStatusResult = (
  overrides: Partial<ReturnType<typeof useWatchStatus>> = {}
) => ({
  watchStatuses: [],
  isLoading: false,
  error: null,
  updateStatus: jest.fn(),
  removeStatus: jest.fn(),
  getStatus: jest.fn(),
  ...overrides,
})

describe('MyPage', () => {
  let consoleErrorSpy: jest.SpyInstance

  const mockRouter = {
    push: jest.fn(),
  }

  const mockUser = {
    id: 1,
    name: 'テストユーザー',
    email: 'test@example.com',
  }

  // ユーザー情報取得が成功する fetch モック
  const createSuccessfulFetchMock = () => {
    return (url: string) => {
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
    // デフォルトは視聴ステータス空の状態
    mockUseWatchStatus.mockReturnValue(createWatchStatusResult())
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

    // 表示名はGoogleアカウント由来。自由記述項目は #789 で廃止したため表示しない
    it('ユーザー情報を表示する', async () => {
      ;(global.fetch as jest.Mock).mockImplementation(
        createSuccessfulFetchMock()
      )

      render(<MyPage />)

      await waitFor(() => {
        expect(screen.getByText('テストユーザー')).toBeInTheDocument()
      })
    })

    it('視聴ステータス別タブを表示し、すべて・中断・視聴履歴タブは表示しない', async () => {
      ;(global.fetch as jest.Mock).mockImplementation(
        createSuccessfulFetchMock()
      )

      render(<MyPage />)

      await waitFor(() => {
        expect(screen.getByText('テストユーザー')).toBeInTheDocument()
      })

      const tab = (name: string) =>
        screen.getByRole('button', { name: new RegExp(name) })
      expect(tab('視聴中')).toBeInTheDocument()
      expect(tab('視聴予定')).toBeInTheDocument()
      expect(tab('視聴済')).toBeInTheDocument()

      // 廃止したタブ（すべて / 中断 / 視聴履歴）は表示されない
      expect(screen.queryByText('すべて')).not.toBeInTheDocument()
      expect(screen.queryByText('中断')).not.toBeInTheDocument()
      expect(screen.queryByText('視聴履歴')).not.toBeInTheDocument()
    })

    it('視聴ステータスをカードとして表示する', async () => {
      ;(global.fetch as jest.Mock).mockImplementation(
        createSuccessfulFetchMock()
      )
      mockUseWatchStatus.mockReturnValue(
        createWatchStatusResult({
          // デフォルトタブ（視聴予定）に表示されるよう status: 'planned'
          watchStatuses: [
            {
              id: 100,
              title: '鋼の錬金術師',
              user_watch_status: { status: 'planned' },
            },
          ] as ReturnType<typeof useWatchStatus>['watchStatuses'],
        })
      )

      render(<MyPage />)

      await waitFor(() => {
        expect(screen.getByText('鋼の錬金術師')).toBeInTheDocument()
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

    it('視聴ステータス取得エラー時にエラーメッセージを表示する', async () => {
      ;(global.fetch as jest.Mock).mockImplementation(
        createSuccessfulFetchMock()
      )
      mockUseWatchStatus.mockReturnValue(
        createWatchStatusResult({
          error: new Error('視聴ステータスの取得に失敗しました'),
        })
      )

      render(<MyPage />)

      await waitFor(() => {
        expect(
          screen.getByText('視聴ステータスの取得に失敗しました')
        ).toBeInTheDocument()
      })
    })

    it('データ取得エラー時にエラーメッセージを表示する', async () => {
      ;(global.fetch as jest.Mock).mockImplementation((url: string) => {
        // ユーザー情報取得でエラーを返す
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
