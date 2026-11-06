import React from 'react'
import { render, screen } from '@testing-library/react'
import { AnimeCard, type AnimeCardData } from './AnimeCard'

jest.mock('next/navigation', () => ({
  usePathname: () => '/search',
}))

const createAnime = (overrides: Partial<AnimeCardData> = {}): AnimeCardData => ({
  id: 1,
  title: '作品1',
  ...overrides,
})

describe('AnimeCard の画像', () => {
  it('カバー画像があればそれを表示する', () => {
    render(
      <AnimeCard
        anime={createAnime({
          cover_image_large: 'https://s4.anilist.co/large.jpg',
          image_url: 'https://sekiro-anime.jp/og-image4.jpg',
        })}
      />
    )

    expect(screen.getByRole('img', { name: '作品1' })).toHaveAttribute(
      'src',
      'https://s4.anilist.co/large.jpg'
    )
  })

  it('CSP で許可されないドメインの image_url は読み込まず、代わりのポスターを表示する', () => {
    const { container } = render(
      <AnimeCard anime={createAnime({ image_url: 'https://sekiro-anime.jp/og-image4.jpg' })} />
    )

    expect(container.querySelector('img')).toBeNull()
    expect(screen.getByText('画像準備中')).toBeInTheDocument()
  })

  it('許可されたドメインの image_url は表示する', () => {
    render(<AnimeCard anime={createAnime({ image_url: 'https://cdn.myanimelist.net/a.jpg' })} />)

    expect(screen.getByRole('img', { name: '作品1' })).toHaveAttribute(
      'src',
      'https://cdn.myanimelist.net/a.jpg'
    )
  })
})
