import React from 'react'
import { render, screen, fireEvent } from '@testing-library/react'
import '@testing-library/jest-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AnimeListPage, parseListFilters, buildListQuery } from './AnimeListPage'

const mockPush = jest.fn()
let mockQuery = ''
jest.mock('next/navigation', () => ({
  useRouter: () => ({ push: mockPush, replace: jest.fn() }),
  usePathname: () => '/anime/current',
  useSearchParams: () => new URLSearchParams(mockQuery),
}))

const mockGrid = jest.fn()
jest.mock('@/components/anime', () => ({
  AnimeGrid: (props: { animes: { title: string }[]; from?: string }) => {
    mockGrid(props)
    return (
      <ul>
        {props.animes.map((a) => (
          <li key={a.title}>{a.title}</li>
        ))}
      </ul>
    )
  },
}))

const animes = [
  { id: 1, title: 'TVのアクション', media: 'tv', genres: ['アクション'], season: '2026-autumn', status: 'airing' },
  { id: 2, title: 'TVのドラマ', media: 'tv', genres: ['ドラマ'], season: '2026-autumn', status: 'airing' },
  { id: 3, title: '映画のアクション', media: 'movie', genres: ['アクション'], season: '2026-autumn', status: 'airing' },
]

const renderPage = () => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={client}>
      <AnimeListPage
        title="今期アニメ"
        queryKey={['test', mockQuery]}
        queryFn={async () => ({ animes, total: animes.length }) as never}
        emptyMessage="なし"
        showMediaTabs
        showGenreSort
      />
    </QueryClientProvider>
  )
}

describe('今期・人気アニメの絞り込み条件のURL保持', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockQuery = ''
  })

  it('URLの条件で絞り込んだ状態を復元する（詳細から戻ったとき）', async () => {
    mockQuery = 'media=tv&genre=アクション'
    renderPage()

    expect(await screen.findByText('TVのアクション')).toBeInTheDocument()
    expect(screen.queryByText('TVのドラマ')).not.toBeInTheDocument()
    expect(screen.queryByText('映画のアクション')).not.toBeInTheDocument()
  })

  it('詳細ページへの遷移元に条件付きのURLを渡す', async () => {
    mockQuery = 'media=tv&genre=アクション'
    renderPage()
    await screen.findByText('TVのアクション')

    const lastProps = mockGrid.mock.calls.at(-1)[0]
    expect(lastProps.from).toBe(`/anime/current?media=tv&genre=${encodeURIComponent('アクション')}`)
  })

  it('条件がなければ遷移元はパスだけにする', async () => {
    renderPage()
    await screen.findByText('TVのドラマ')

    expect(mockGrid.mock.calls.at(-1)[0].from).toBe('/anime/current')
  })

  it('タブを切り替えると条件をURLに書き込み、ページを1に戻す', async () => {
    // 2件しか該当しないため2ページ目は空になるが、タブの操作だけを確かめる
    mockQuery = 'genre=アクション&page=2'
    renderPage()

    fireEvent.click(await screen.findByRole('tab', { name: '映画' }))
    expect(mockPush).toHaveBeenCalledWith(
      `/anime/current?media=movie&genre=${encodeURIComponent('アクション')}`,
      { scroll: false }
    )
  })
})

describe('parseListFilters / buildListQuery', () => {
  it('不正な値は既定値に落とす', () => {
    expect(parseListFilters(new URLSearchParams('media=dvd&sort=xxx&page=-1'))).toEqual({
      media: 'all',
      genres: [],
      sort: 'popularity',
      page: 1,
    })
  })

  it('既定値はクエリから省く', () => {
    expect(buildListQuery({ media: 'all', genres: [], sort: 'popularity', page: 1 })).toBe('')
  })

  it('往復で条件が変わらない', () => {
    const filters = { media: 'web' as const, genres: ['SF', 'ドラマ'], sort: 'year', page: 3 }
    expect(parseListFilters(new URLSearchParams(buildListQuery(filters)))).toEqual(filters)
  })
})

describe('一覧の取得エラー', () => {
  it('見出しを残してErrorStateを表示し、ページ固有の文言は出さない', async () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(
      <QueryClientProvider client={client}>
        <AnimeListPage
          title="今期アニメ"
          queryKey={['err-test']}
          queryFn={async () => {
            throw { status: 0 }
          }}
          emptyMessage="なし"
        />
      </QueryClientProvider>
    )
    expect(await screen.findByText('接続できませんでした')).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: '今期アニメ' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '再読み込み' })).toBeInTheDocument()
  })
})
