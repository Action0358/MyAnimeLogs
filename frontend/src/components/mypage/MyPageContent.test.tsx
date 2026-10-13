import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import '@testing-library/jest-dom'
import { MyPageContent } from './MyPageContent'
import { useAuth } from '@/contexts/AuthContext'
import { useWatchStatus } from '@/hooks/useWatchStatus'
import { useRouter } from 'next/navigation'
import { get } from '@/lib/api/client'

jest.mock('next/navigation')
jest.mock('@/contexts/AuthContext', () => ({ useAuth: jest.fn() }))
jest.mock('@/hooks/useWatchStatus', () => ({ useWatchStatus: jest.fn() }))
jest.mock('@/lib/api/client', () => ({
  get: jest.fn(),
  ApiClientError: class extends Error {},
}))

const user = { id: 1, name: 'テストユーザー', email: 'test@example.com', created_at: '2026-01-01T00:00:00Z' }

const anime = (id: number, status: string, title: string) => ({
  id,
  title,
  user_watch_status: { status },
})

describe('MyPageContent のタブ', () => {
  const mockReplace = jest.fn()
  const mockPush = jest.fn()

  beforeEach(() => {
    jest.clearAllMocks()
    ;(useRouter as jest.Mock).mockReturnValue({ replace: mockReplace, push: mockPush })
    ;(useAuth as jest.Mock).mockReturnValue({
      user,
      token: 't',
      isLoading: false,
      isAuthenticated: true,
    })
    ;(get as jest.Mock).mockResolvedValue(user)
    ;(useWatchStatus as jest.Mock).mockReturnValue({
      watchStatuses: [anime(1, 'planned', '予定の作品'), anime(2, 'watching', '視聴中の作品')],
      isLoading: false,
      error: null,
    })
    window.history.replaceState(null, '', '/mypage')
  })

  const tab = (name: string) => screen.getByRole('button', { name: new RegExp(`^${name}`) })

  it('?tab が無ければ視聴予定を開く', async () => {
    render(<MyPageContent />)
    await waitFor(() => expect(tab('視聴予定')).toHaveAttribute('aria-current', 'page'))
  })

  // 作品詳細の「マイページに戻る」で、遷移前のタブへ戻ってこられるようにする
  it('?tab=watching なら視聴中を開く', async () => {
    window.history.replaceState(null, '', '/mypage?tab=watching')
    render(<MyPageContent />)

    await waitFor(() => expect(tab('視聴中')).toHaveAttribute('aria-current', 'page'))
    expect(screen.getByText('視聴中の作品')).toBeInTheDocument()
  })

  it('不正な ?tab は視聴予定に戻す', async () => {
    window.history.replaceState(null, '', '/mypage?tab=unknown')
    render(<MyPageContent />)
    await waitFor(() => expect(tab('視聴予定')).toHaveAttribute('aria-current', 'page'))
  })

  it('タブを切り替えるとURLの ?tab を履歴を積まずに更新する', async () => {
    render(<MyPageContent />)
    await waitFor(() => expect(tab('視聴中')).toBeInTheDocument())

    await userEvent.click(tab('視聴中'))

    expect(mockReplace).toHaveBeenCalledWith('/mypage?tab=watching', { scroll: false })
  })

  // 設定画面のパンくずから、開いていたタブへ戻れるようにする
  it('設定画面へ表示中のタブを渡す', async () => {
    window.history.replaceState(null, '', '/mypage?tab=watching')
    render(<MyPageContent />)

    await userEvent.click(await screen.findByRole('button', { name: '設定' }))

    expect(mockPush).toHaveBeenCalledWith('/mypage/edit?tab=watching')
  })

  it('作品カードの戻り先に表示中のタブを含める', async () => {
    window.history.replaceState(null, '', '/mypage?tab=watching')
    render(<MyPageContent />)

    const link = await screen.findByRole('link', { name: /視聴中の作品/ })
    expect(link).toHaveAttribute(
      'href',
      `/anime/2?from=${encodeURIComponent('/mypage?tab=watching')}`,
    )
  })
})
