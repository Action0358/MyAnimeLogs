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

const mockAnimes = (overrides = {}) =>
  mockUseAnimes.mockReturnValue({
    animes: [],
    isLoading: false,
    totalPages: 1,
    total: 0,
    setPage: jest.fn(),
    isFetching: false,
    ...overrides,
  })

describe('検索画面の絞り込みチップ行', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockAnimes()
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

/**
 * 見出し・件数・フィルターの配置を、今期アニメ／人気アニメ（AnimeListPage）に
 * 揃えたことを固定する（#843）。
 */
describe('検索画面のヘッダー', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockAnimes({ total: 8171 })
    setFilters({})
  })

  it('見出しは一覧ページと同じサイズで、SPでも隠さない', () => {
    render(<SearchPage />)

    const heading = screen.getByRole('heading', { name: 'アニメ検索', level: 1 })
    // AnimeListPage と同じクラス。sr-only で消さない
    expect(heading).toHaveClass('text-xl', 'md:text-[26px]')
    expect(heading).not.toHaveClass('sr-only')
  })

  it('件数を見出しの隣に「N 件」で表示する', () => {
    render(<SearchPage />)

    const heading = screen.getByRole('heading', { name: 'アニメ検索', level: 1 })
    const headerRow = heading.parentElement as HTMLElement

    expect(headerRow).toHaveTextContent('8,171')
    // 一覧ページと同じ文言。「件のアニメ」ではない
    expect(headerRow).toHaveTextContent('件')
    expect(headerRow).not.toHaveTextContent('件のアニメ')
  })

  it('読み込み中は件数の代わりに読み込み表示を出す', () => {
    mockAnimes({ isLoading: true, total: 0 })
    render(<SearchPage />)

    expect(screen.getByText('読み込み中...')).toBeInTheDocument()
  })

  it('フィルターはメディアタブと同じ行に置く', () => {
    render(<SearchPage />)

    const tabs = screen.getByRole('tablist', { name: 'メディアタイプフィルター' })
    const filterButton = screen.getByRole('button', { name: /フィルター/ })

    // 同じ親（タブ行）に属していること
    expect(tabs.parentElement).toBe(filterButton.parentElement)
  })

  it('絞り込み中はフィルターにバッジ件数を出す', () => {
    setFilters({ season: '2024年春', genre: ['アクション', 'ドラマ'] })
    render(<SearchPage />)

    const filterButton = screen.getByRole('button', { name: /フィルター/ })
    // 放送時期1 + ジャンル2
    expect(filterButton).toHaveTextContent('3')
  })
})
