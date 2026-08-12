import React from 'react'
import { render, screen, waitFor, act } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import '@testing-library/jest-dom'
import LoginPage from './page'
import { useAuth } from '@/contexts/AuthContext'
import { useRouter, usePathname, useSearchParams } from 'next/navigation'

jest.mock('@/contexts/AuthContext', () => ({
  useAuth: jest.fn(),
  DEFAULT_LOGIN_ERROR: 'ログイン処理中にエラーが発生しました。しばらく時間をおいて再度お試しください',
}))
jest.mock('next/navigation')
// ポスターパネルは TanStack Query を使うため、ここでは描画のみ確認できれば十分
jest.mock('@/components/auth/AuthPosterPanel', () => ({
  AuthPosterPanel: ({ headline }: { headline: React.ReactNode }) => <div>{headline}</div>,
}))

const mockUseAuth = useAuth as jest.Mock
const mockUseRouter = useRouter as jest.Mock
const mockUsePathname = usePathname as jest.Mock
const mockUseSearchParams = useSearchParams as jest.Mock

describe('LoginPage', () => {
  const mockLoginWithGoogle = jest.fn()
  const mockReplace = jest.fn()

  const setSearchParams = (query: string) => {
    mockUseSearchParams.mockReturnValue(new URLSearchParams(query))
  }

  beforeEach(() => {
    jest.clearAllMocks()
    mockUseAuth.mockReturnValue({ loginWithGoogle: mockLoginWithGoogle })
    mockUseRouter.mockReturnValue({ replace: mockReplace, push: jest.fn() })
    mockUsePathname.mockReturnValue('/login')
    setSearchParams('')
  })

  it('Googleログインボタンと新規登録リンクを表示する', async () => {
    render(<LoginPage />)

    expect(await screen.findByRole('button', { name: 'Googleでログイン' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: '新規登録' })).toHaveAttribute('href', '/register')
  })

  // パスワード認証は廃止済み。UIに復活していないことを保証する
  it('メールアドレス・パスワードの入力欄を表示しない', async () => {
    render(<LoginPage />)
    await screen.findByRole('button', { name: 'Googleでログイン' })

    expect(screen.queryByLabelText(/メールアドレス/)).not.toBeInTheDocument()
    expect(screen.queryByLabelText(/パスワード/)).not.toBeInTheDocument()
    expect(screen.queryByText(/お忘れの方/)).not.toBeInTheDocument()
  })

  it('ボタン押下で loginWithGoogle を呼ぶ', async () => {
    mockLoginWithGoogle.mockResolvedValue(undefined)

    render(<LoginPage />)
    await act(async () => {
      await userEvent.click(await screen.findByRole('button', { name: 'Googleでログイン' }))
    })

    expect(mockLoginWithGoogle).toHaveBeenCalledTimes(1)
  })

  it('成功時はGoogleへ遷移するため、ローディング表示を維持する', async () => {
    mockLoginWithGoogle.mockResolvedValue(undefined)

    render(<LoginPage />)
    await act(async () => {
      await userEvent.click(await screen.findByRole('button', { name: 'Googleでログイン' }))
    })

    await waitFor(() => {
      expect(screen.getByText('Googleに接続しています...')).toBeInTheDocument()
    })
  })

  it('失敗時はエラーを表示し、再操作できる状態に戻す', async () => {
    mockLoginWithGoogle.mockRejectedValue(new Error('Googleログインがキャンセルされました'))

    render(<LoginPage />)
    await act(async () => {
      await userEvent.click(await screen.findByRole('button', { name: 'Googleでログイン' }))
    })

    expect(await screen.findByText('Googleログインがキャンセルされました')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Googleでログイン' })).toBeEnabled()
  })

  it('reason=timeout のときセッションタイムアウトを案内する', async () => {
    setSearchParams('reason=timeout')

    render(<LoginPage />)

    expect(
      await screen.findByText('セッションがタイムアウトしました。再度ログインしてください。')
    ).toBeInTheDocument()
  })

  it('reason=auth-callback-failed のときGoogleログイン失敗を案内する', async () => {
    setSearchParams('reason=auth-callback-failed')

    render(<LoginPage />)

    expect(
      await screen.findByText(
        'Googleログインを完了できませんでした。お手数ですが再度お試しください。'
      )
    ).toBeInTheDocument()
  })

  it('ボタン押下時にURLから reason パラメータを取り除く', async () => {
    setSearchParams('reason=timeout')
    mockLoginWithGoogle.mockResolvedValue(undefined)

    render(<LoginPage />)
    await act(async () => {
      await userEvent.click(await screen.findByRole('button', { name: 'Googleでログイン' }))
    })

    expect(mockReplace).toHaveBeenCalledWith('/login')
  })
})
