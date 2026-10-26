import React from 'react'
import { render, waitFor } from '@testing-library/react'
import AuthCallbackPage from './page'
import { useRouter, useSearchParams } from 'next/navigation'
import { supabase } from '@/lib/supabase'

jest.mock('next/navigation')
jest.mock('@/lib/supabase', () => ({
  supabase: {
    auth: {
      exchangeCodeForSession: jest.fn(),
      getSession: jest.fn(),
    },
  },
}))
jest.mock('@/components/layout', () => ({
  Container: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

const mockUseRouter = useRouter as jest.Mock
const mockUseSearchParams = useSearchParams as jest.Mock
const mockExchange = supabase.auth.exchangeCodeForSession as jest.Mock
const mockGetSession = supabase.auth.getSession as jest.Mock

describe('AuthCallbackPage', () => {
  const mockReplace = jest.fn()

  beforeEach(() => {
    jest.clearAllMocks()
    sessionStorage.clear()
    mockUseRouter.mockReturnValue({ replace: mockReplace })
    mockUseSearchParams.mockReturnValue(new URLSearchParams('code=abc'))
    mockExchange.mockResolvedValue({ data: { session: { access_token: 't' } }, error: null })
  })

  it('戻り先が無ければトップへ遷移する', async () => {
    render(<AuthCallbackPage />)
    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith('/'))
  })

  // 作品詳細からログインした場合、その作品へ戻す
  it('ログイン前に退避した戻り先へ遷移し、戻り先を消す', async () => {
    sessionStorage.setItem('auth_return_path', '/anime/123')

    render(<AuthCallbackPage />)

    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith('/anime/123'))
    expect(sessionStorage.getItem('auth_return_path')).toBeNull()
  })

  it('implicit フローでも戻り先へ遷移する', async () => {
    mockUseSearchParams.mockReturnValue(new URLSearchParams(''))
    mockGetSession.mockResolvedValue({ data: { session: { access_token: 't' } } })
    sessionStorage.setItem('auth_return_path', '/anime/123')

    render(<AuthCallbackPage />)

    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith('/anime/123'))
  })

  // 失敗時はログイン画面で再試行するため、戻り先は使わない
  it('セッション確立に失敗したらログイン画面へ戻す', async () => {
    mockExchange.mockResolvedValue({ data: { session: null }, error: new Error('x') })
    sessionStorage.setItem('auth_return_path', '/anime/123')

    render(<AuthCallbackPage />)

    await waitFor(() =>
      expect(mockReplace).toHaveBeenCalledWith('/signin?reason=auth-callback-failed')
    )
  })
})
