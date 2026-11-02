import React from 'react'
import { render } from '@testing-library/react'
import { useQueryClient, type QueryClient } from '@tanstack/react-query'
import QueryProvider from './QueryProvider'
import { useAuth } from '@/contexts/AuthContext'

jest.mock('@/contexts/AuthContext')

const mockUseAuth = useAuth as jest.Mock

describe('QueryProvider のログアウト時のキャッシュ', () => {
  let client: QueryClient

  const Capture: React.FC = () => {
    client = useQueryClient()
    return null
  }

  // 同じ要素を rerender すると React が描画を省くため、毎回作り直す
  const ui = () => (
    <QueryProvider>
      <Capture />
    </QueryProvider>
  )

  it('ログアウトするとユーザーのデータだけを消し、公開データは残す', () => {
    mockUseAuth.mockReturnValue({ isAuthenticated: true })
    const { rerender } = render(ui())
    client.setQueryData(['/users/me/favorites'], ['お気に入り'])
    client.setQueryData(['/users/me/watch-status'], ['視聴ステータス'])
    client.setQueryData(['animes', 'current'], ['公開データ'])

    rerender(ui())
    expect(client.getQueryData(['/users/me/favorites'])).toEqual(['お気に入り'])

    mockUseAuth.mockReturnValue({ isAuthenticated: false })
    rerender(ui())

    expect(client.getQueryData(['/users/me/favorites'])).toBeUndefined()
    expect(client.getQueryData(['/users/me/watch-status'])).toBeUndefined()
    expect(client.getQueryData(['animes', 'current'])).toEqual(['公開データ'])
  })
})
