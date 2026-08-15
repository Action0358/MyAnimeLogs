import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthPosterPanel } from './AuthPosterPanel'
import { animeApi, type DBAnime } from '@/lib/api/annict'

jest.mock('@/lib/api/annict', () => ({
  animeApi: {
    getPopularAnimes: jest.fn(),
  },
}))

const mockGetPopularAnimes = animeApi.getPopularAnimes as jest.Mock

const createAnime = (id: number, overrides: Partial<DBAnime> = {}): DBAnime => ({
  id,
  title: `作品${id}`,
  watchers_count: 0,
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
  cover_image_large: `https://example.com/large-${id}.jpg`,
  ...overrides,
})

const renderPanel = (ui: React.ReactNode) => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(<QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>)
}

describe('AuthPosterPanel', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('見出しと子要素を表示する', async () => {
    mockGetPopularAnimes.mockResolvedValue({ animes: [], total: 0 })

    renderPanel(
      <AuthPosterPanel headline="観た作品を、そのまま記録に。">
        <p>補足テキスト</p>
      </AuthPosterPanel>
    )

    expect(screen.getByText('観た作品を、そのまま記録に。')).toBeInTheDocument()
    expect(screen.getByText('補足テキスト')).toBeInTheDocument()
    expect(screen.getByText('MyAnimeLogs')).toBeInTheDocument()
  })

  it('取得した人気アニメのカバー画像を表示する', async () => {
    mockGetPopularAnimes.mockResolvedValue({
      animes: [createAnime(1), createAnime(2)],
      total: 2,
    })

    const { container } = renderPanel(<AuthPosterPanel headline="見出し" />)

    await waitFor(() => {
      expect(container.querySelectorAll('img')).toHaveLength(2)
    })
    const images = container.querySelectorAll('img')
    expect(images[0]).toHaveAttribute('src', 'https://example.com/large-1.jpg')
    // 装飾目的の画像のため、読み上げ対象にしない
    expect(images[0]).toHaveAttribute('alt', '')
  })

  it('cover_image_large が無い場合は medium、それも無ければ image_url を使う', async () => {
    mockGetPopularAnimes.mockResolvedValue({
      animes: [
        createAnime(1, {
          cover_image_large: undefined,
          cover_image_medium: 'https://example.com/medium.jpg',
        }),
        createAnime(2, {
          cover_image_large: undefined,
          cover_image_medium: undefined,
          image_url: 'https://example.com/fallback.jpg',
        }),
      ],
      total: 2,
    })

    const { container } = renderPanel(<AuthPosterPanel headline="見出し" />)

    await waitFor(() => {
      expect(container.querySelectorAll('img')).toHaveLength(2)
    })
    const srcs = Array.from(container.querySelectorAll('img')).map((img) => img.getAttribute('src'))
    expect(srcs).toEqual(['https://example.com/medium.jpg', 'https://example.com/fallback.jpg'])
  })

  it('画像URLを一つも持たないアニメは描画しない', async () => {
    mockGetPopularAnimes.mockResolvedValue({
      animes: [
        createAnime(1, { cover_image_large: undefined }),
        createAnime(2),
      ],
      total: 2,
    })

    const { container } = renderPanel(<AuthPosterPanel headline="見出し" />)

    await waitFor(() => {
      expect(container.querySelectorAll('img')).toHaveLength(1)
    })
  })

  // 認証画面の装飾でしかないため、画像が取れなくてもログイン導線は成立させる
  it('API が失敗しても見出しを表示し続ける', async () => {
    mockGetPopularAnimes.mockRejectedValue(new Error('人気アニメの取得に失敗しました'))

    const { container } = renderPanel(<AuthPosterPanel headline="見出し" />)

    await waitFor(() => {
      expect(mockGetPopularAnimes).toHaveBeenCalled()
    })
    expect(screen.getByText('見出し')).toBeInTheDocument()
    expect(container.querySelectorAll('img')).toHaveLength(0)
  })
})
