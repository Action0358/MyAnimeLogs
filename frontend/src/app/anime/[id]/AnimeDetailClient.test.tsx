import React from 'react'
import { act, fireEvent, render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import AnimeDetailClient from './AnimeDetailClient'
import { getAnimesId } from '@/lib/api/generated/myanimelogs'

let mockParams: { id: string } = { id: '123' }
jest.mock('next/navigation', () => ({
  useParams: () => mockParams,
  usePathname: () => '/anime/123',
}))
jest.mock('@/lib/api/generated/myanimelogs', () => ({
  getAnimesId: jest.fn(),
}))
jest.mock('@/contexts/AuthContext', () => ({
  useAuth: () => ({ isAuthenticated: false }),
}))
jest.mock('@/hooks/useFavorites', () => ({
  useFavorites: () => ({
    isFavorite: () => false,
    addFavorite: jest.fn(),
    removeFavorite: jest.fn(),
  }),
}))
jest.mock('@/hooks/useWatchStatus', () => ({
  useWatchStatus: () => ({
    getStatus: () => null,
    updateStatus: jest.fn(),
    removeStatus: jest.fn(),
  }),
}))

const mockGetAnimesId = getAnimesId as jest.Mock

// @types/jest が advanceTimersByTimeAsync（Jest 29.5+）をまだ持たないため型だけ補う
const advanceTimers = (ms: number) =>
  (
    jest as unknown as { advanceTimersByTimeAsync: (ms: number) => Promise<void> }
  ).advanceTimersByTimeAsync(ms)

const renderDetail = () => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={client}>
      <AnimeDetailClient />
    </QueryClientProvider>
  )
}

describe('作品詳細の取得エラー', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockParams = { id: '123' }
  })

  it('APIが404を返したら「作品が見つかりません」を出す', async () => {
    mockGetAnimesId.mockRejectedValue({ isAxiosError: true, response: { status: 404 } })

    renderDetail()

    expect(await screen.findByRole('heading', { name: '作品が見つかりません' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: '作品を検索' })).toHaveAttribute('href', '/search')
  })

  it('IDが不正なときも「作品が見つかりません」を出す', async () => {
    mockParams = { id: 'abc' }

    renderDetail()

    expect(await screen.findByRole('heading', { name: '作品が見つかりません' })).toBeInTheDocument()
    expect(mockGetAnimesId).not.toHaveBeenCalled()
  })

  it('通信エラーは存在しない作品と区別し、再読み込みで再取得する', async () => {
    // 404以外は3回まで再試行する（指数バックオフ）ため、タイマーを進めて待つ
    jest.useFakeTimers()
    try {
      mockGetAnimesId.mockRejectedValue({ isAxiosError: true })

      renderDetail()
      await act(async () => {
        await advanceTimers(10000)
      })

      expect(screen.getByText('接続できませんでした')).toBeInTheDocument()
      expect(screen.queryByText('作品が見つかりません')).not.toBeInTheDocument()
      const calls = mockGetAnimesId.mock.calls.length

      fireEvent.click(screen.getByRole('button', { name: '再読み込み' }))
      await act(async () => {
        await advanceTimers(10000)
      })
      expect(mockGetAnimesId.mock.calls.length).toBeGreaterThan(calls)
    } finally {
      jest.useRealTimers()
    }
  })

  it('404は再試行せず、すぐに表示する', async () => {
    mockGetAnimesId.mockRejectedValue({ isAxiosError: true, response: { status: 404 } })

    renderDetail()
    await screen.findByRole('heading', { name: '作品が見つかりません' })

    expect(mockGetAnimesId).toHaveBeenCalledTimes(1)
  })
})

// 静的エクスポートでは詳細ページのシェル（params.id = '_'）が /anime/<id> のリライト先と
// 404.html を兼ねるため、実際のURLから作品IDか「ページ自体が無い」かを判断する
describe('シェル（404.html兼用）として描画されたとき', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockParams = { id: '_' }
  })

  afterEach(() => {
    window.history.pushState({}, '', '/')
  })

  it('作品以外のURLでは「ページが見つかりません」を出し、作品を取得しない', async () => {
    window.history.pushState({}, '', '/this-page-does-not-exist')

    renderDetail()

    expect(
      await screen.findByRole('heading', { name: 'ページが見つかりません' })
    ).toBeInTheDocument()
    expect(mockGetAnimesId).not.toHaveBeenCalled()
  })

  it('/anime/ 配下で数値でないIDは、ローディングで止まらず「作品が見つかりません」を出す', async () => {
    window.history.pushState({}, '', '/anime/abc/')

    renderDetail()

    expect(await screen.findByRole('heading', { name: '作品が見つかりません' })).toBeInTheDocument()
    expect(mockGetAnimesId).not.toHaveBeenCalled()
  })

  it('/anime/<id> ではURLのIDで作品を取得する', async () => {
    window.history.pushState({}, '', '/anime/456/')
    mockGetAnimesId.mockRejectedValue({ isAxiosError: true, response: { status: 404 } })

    renderDetail()

    await screen.findByRole('heading', { name: '作品が見つかりません' })
    expect(mockGetAnimesId).toHaveBeenCalledWith(456)
  })
})
