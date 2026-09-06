import React from 'react'
import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import SearchPage from './page'
import { useSearch, type SearchFilters } from '@/hooks/useSearch'
import { useAnimes } from '@/hooks/useAnimes'

jest.mock('next/navigation', () => ({
  useSearchParams: () => new URLSearchParams(),
  useRouter: () => ({ push: jest.fn(), replace: jest.fn() }),
}))
jest.mock('@/hooks/useSearch')
jest.mock('@/hooks/useAnimes')
jest.mock('@/components/anime', () => ({
  AnimeGrid: () => <div data-testid="anime-grid" />,
}))

const mockUseSearch = useSearch as jest.Mock
const mockUseAnimes = useAnimes as jest.Mock

const setFilters = (filters: SearchFilters) => {
  mockUseSearch.mockReturnValue({
    filters,
    updateFilters: jest.fn(),
    searchHistory: [],
    addToHistory: jest.fn(),
    clearHistory: jest.fn(),
  })
}

/** チップ行は中身がないと mb-6 ぶんの余白だけを作る。存在有無で判定する */
const chipRow = () => document.querySelector('.flex.flex-wrap.gap-2')

describe('検索画面の絞り込みチップ行', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockUseAnimes.mockReturnValue({
      animes: [],
      isLoading: false,
      totalPages: 1,
      total: 0,
      setPage: jest.fn(),
      isFetching: false,
    })
  })

  /**
   * メディアタブには対応するチップがないため、media だけが指定されている状態で
   * チップ行を描画すると、高さ0のまま余白だけが増えてタブ列が24pxずれる（#838）。
   */
  it('メディア種別だけが指定されている場合はチップ行を描画しない', () => {
    setFilters({ media: 'tv' } as SearchFilters)
    render(<SearchPage />)

    expect(chipRow()).toBeNull()
  })

  it('絞り込みが何もない場合もチップ行を描画しない', () => {
    setFilters({})
    render(<SearchPage />)

    expect(chipRow()).toBeNull()
  })

  it('キーワードがあればチップ行を描画する', () => {
    setFilters({ keyword: '鬼滅' })
    render(<SearchPage />)

    expect(chipRow()).not.toBeNull()
    expect(screen.getByText(/鬼滅/)).toBeInTheDocument()
  })

  it('ジャンルがあればチップ行を描画する', () => {
    setFilters({ genre: ['アクション'] })
    render(<SearchPage />)

    expect(chipRow()).not.toBeNull()
  })

  it('放送時期があればチップ行を描画する', () => {
    setFilters({ season: '2024年春' })
    render(<SearchPage />)

    expect(chipRow()).not.toBeNull()
  })

  it('既定以外の並び替えがあればチップ行を描画する', () => {
    setFilters({ sort: 'popularity' })
    render(<SearchPage />)

    expect(chipRow()).not.toBeNull()
  })

  it('メディア種別と既定の並び替えの組み合わせでもチップ行を描画しない', () => {
    setFilters({ media: 'movie', sort: 'year' } as SearchFilters)
    render(<SearchPage />)

    expect(chipRow()).toBeNull()
  })
})
