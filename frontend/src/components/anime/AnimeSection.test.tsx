import React from 'react'
import { render, screen, fireEvent } from '@testing-library/react'
import '@testing-library/jest-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import AnimeSection from './AnimeSection'

jest.mock('./AnimeCarousel', () => ({
  AnimeCarousel: () => <div data-testid="carousel" />,
  CarouselSkeleton: () => <div />,
}))

const renderSection = (queryFn: () => Promise<never>) => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={client}>
      <AnimeSection
        title="今期アニメ"
        queryKey={['section-err']}
        queryFn={queryFn}
        linkHref="/anime/current"
        emptyMessage="なし"
        displayMode="carousel"
      />
    </QueryClientProvider>
  )
}

describe('AnimeSection の取得エラー', () => {
  it('見出しを残し、compact の ErrorState を出して「もっと見る」は隠す', async () => {
    renderSection(async () => {
      throw { status: 0 }
    })
    expect(await screen.findByText('接続できませんでした')).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: '今期アニメ' })).toBeInTheDocument()
    expect(screen.queryByText('もっと見る')).not.toBeInTheDocument()
    expect(screen.queryByText('トップへ戻る')).not.toBeInTheDocument()
  })

  it('再読み込みでそのセクションのAPIを再取得する', async () => {
    const queryFn = jest.fn().mockRejectedValue(new Error('x'))
    renderSection(queryFn)
    fireEvent.click(await screen.findByRole('button', { name: '再読み込み' }))
    expect(queryFn).toHaveBeenCalledTimes(2)
  })
})
