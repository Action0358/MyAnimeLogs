import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import MyPage from './page'
import { useAuth } from '@/contexts/AuthContext'
import { useRouter } from 'next/navigation'
import { useWatchStatus } from '@/hooks/useWatchStatus'
import { ApiClientError } from '@/lib/api/client'

// Mock dependencies
jest.mock('@/contexts/AuthContext')
jest.mock('next/navigation')
jest.mock('@/hooks/useWatchStatus')
jest.mock('@/components/auth/ProtectedRoute', () => ({
  ProtectedRoute: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

// ユーザー情報は AuthContext から受け取り、APIから取り直さないことを確認する
global.fetch = jest.fn()

const mockUseWatchStatus = useWatchStatus as jest.Mock

// useWatchStatus の戻り値を生成するヘルパー
const createWatchStatusResult = (overrides: Partial<ReturnType<typeof useWatchStatus>> = {}) => ({
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
      render(<MyPage />)

      await waitFor(() => {
        expect(screen.getByText('テストユーザー')).toBeInTheDocument()
      })
    })

    it('視聴ステータス別タブを表示し、すべて・中断・視聴履歴タブは表示しない', async () => {
      render(<MyPage />)

      await waitFor(() => {
        expect(screen.getByText('テストユーザー')).toBeInTheDocument()
      })

      const tab = (name: string) => screen.getByRole('button', { name: new RegExp(name) })
      expect(tab('視聴中')).toBeInTheDocument()
      expect(tab('視聴予定')).toBeInTheDocument()
      expect(tab('視聴済')).toBeInTheDocument()

      // 廃止したタブ（すべて / 中断 / 視聴履歴）は表示されない
      expect(screen.queryByText('すべて')).not.toBeInTheDocument()
      expect(screen.queryByText('中断')).not.toBeInTheDocument()
      expect(screen.queryByText('視聴履歴')).not.toBeInTheDocument()
    })

    it('視聴ステータスをカードとして表示する', async () => {
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

    it('視聴ステータスの読み込み中は読み込み表示を出す', () => {
      mockUseWatchStatus.mockReturnValue(createWatchStatusResult({ isLoading: true }))

      render(<MyPage />)

      expect(screen.getByText('読み込み中...')).toBeInTheDocument()
    })

    it('ユーザー情報をAPIから取り直さない（#1236）', () => {
      render(<MyPage />)

      expect(global.fetch).not.toHaveBeenCalled()
      expect(screen.getByText('テストユーザー')).toBeInTheDocument()
    })

    it('視聴ステータス取得エラー時にエラーメッセージを表示する', async () => {
      mockUseWatchStatus.mockReturnValue(
        createWatchStatusResult({
          error: new Error('視聴ステータスの取得に失敗しました'),
        })
      )

      render(<MyPage />)

      await waitFor(() => {
        expect(screen.getByText('読み込めませんでした')).toBeInTheDocument()
        // APIの生メッセージは画面に出さない
        expect(screen.queryByText('視聴ステータスの取得に失敗しました')).not.toBeInTheDocument()
      })
    })

    it('ネットワークエラー時にエラーメッセージを表示する', async () => {
      mockUseWatchStatus.mockReturnValue(
        createWatchStatusResult({ error: new ApiClientError('通信エラー', 0) })
      )

      render(<MyPage />)

      await waitFor(() => {
        expect(screen.getByText('接続できませんでした')).toBeInTheDocument()
      })
    })
  })
})
