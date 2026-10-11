import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { useAnimes } from './useAnimes'
import { getAnimes } from '@/lib/api/generated/myanimelogs'

jest.mock('@/lib/api/generated/myanimelogs', () => ({
  getAnimes: jest.fn(),
}))

const mockedGetAnimes = getAnimes as jest.MockedFunction<typeof getAnimes>

// ページ番号ごとに異なるIDの1件を返すダミーレスポンス。
// 「表示されている作品がどのページのものか」をIDで判別するために使う。
const buildResponse = (page: number) => ({
  data: {
    data: [{ id: page, title: `anime-page-${page}` }],
    meta: { total: 100, total_pages: 10, current_page: page },
  },
})

describe('useAnimes', () => {
  beforeEach(() => {
    mockedGetAnimes.mockImplementation(async (params) =>
      buildResponse((params?.page as number) ?? 1) as never
    )
  })

  afterEach(() => {
    jest.clearAllMocks()
  })

  const wrapper = ({ children }: { children: ReactNode }) => {
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  }

  it('initialPageが変化すると、表示中の一覧も新しいページに同期される', async () => {
    // 検索ページの実際の使い方を再現する: フィルター変更前はinitialPage=3（3ページ目を閲覧中）
    const { result, rerender } = renderHook(
      ({ initialPage }) => useAnimes({ initialPage, prefetchNextPage: false }),
      { wrapper, initialProps: { initialPage: 3 } }
    )

    await waitFor(() => expect(result.current.animes).toHaveLength(1))
    expect(result.current.animes[0].id).toBe(3)
    expect(result.current.currentPage).toBe(3)

    // フィルター変更（useSearchがURLのpageパラメータを削除する）により、
    // 呼び出し側から渡されるinitialPageが1に変わるケースを再現する
    rerender({ initialPage: 1 })

    await waitFor(() => expect(result.current.currentPage).toBe(1))
    await waitFor(() => expect(result.current.animes[0]?.id).toBe(1))
  })
})
