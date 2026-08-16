import React from 'react'
import { render, screen, waitFor, act } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import '@testing-library/jest-dom'
import SignInPage from './page'
import { useAuth } from '@/contexts/AuthContext'
import { useRouter, usePathname, useSearchParams } from 'next/navigation'

jest.mock('@/contexts/AuthContext', () => ({
  useAuth: jest.fn(),
  DEFAULT_LOGIN_ERROR: 'ログイン処理中にエラーが発生しました。しばらく時間をおいて再度お試しください',
}))
jest.mock('next/navigation')
jest.mock('@/components/auth/AuthPosterPanel', () => ({
  AuthPosterPanel: ({ headline }: { headline: React.ReactNode }) => <div>{headline}</div>,
}))

const mockUseAuth = useAuth as jest.Mock
const mockUseRouter = useRouter as jest.Mock
const mockUsePathname = usePathname as jest.Mock
const mockUseSearchParams = useSearchParams as jest.Mock

describe('SignInPage', () => {
  const mockLoginWithGoogle = jest.fn()
  const mockReplace = jest.fn()

  const setSearchParams = (query: string) => {
    mockUseSearchParams.mockReturnValue(new URLSearchParams(query))
  }

  beforeEach(() => {
    jest.clearAllMocks()
    mockUseAuth.mockReturnValue({ loginWithGoogle: mockLoginWithGoogle })
    mockUseRouter.mockReturnValue({ replace: mockReplace, push: jest.fn() })
    mockUsePathname.mockReturnValue('/signin')
    setSearchParams('')
  })

  // ログインと新規登録が同じ操作であることを見出しで示す
  it('ログインと新規作成の両方を担うことを見出しで示す', async () => {
    render(<SignInPage />)

    expect(
      await screen.findByRole('heading', { name: 'ログインまたは新規作成' })
    ).toBeInTheDocument()
    expect(screen.getByText('初めての方はそのまま登録が完了します。')).toBeInTheDocument()
  })

  it('主操作はGoogleボタン1つに絞る', async () => {
    render(<SignInPage />)

    expect(await screen.findByRole('button', { name: 'Googleで続行' })).toBeInTheDocument()
    // 統合したため、別画面へ誘導するリンクは持たない
    expect(screen.queryByRole('link', { name: '新規登録' })).not.toBeInTheDocument()
    expect(screen.queryByRole('link', { name: 'ログイン' })).not.toBeInTheDocument()
  })

  it('パスワード入力欄を表示しない', async () => {
    render(<SignInPage />)
    await screen.findByRole('button', { name: 'Googleで続行' })

    expect(screen.queryByLabelText(/メールアドレス/)).not.toBeInTheDocument()
    expect(screen.queryByLabelText(/パスワード/)).not.toBeInTheDocument()
  })

  // 登録もこの画面で発生するため、同意文言が必要
  it('利用規約とプライバシーポリシーへのリンクを表示する', async () => {
    render(<SignInPage />)
    await screen.findByRole('button', { name: 'Googleで続行' })

    expect(screen.getByRole('link', { name: '利用規約' })).toHaveAttribute('href', '/terms')
    expect(screen.getByRole('link', { name: 'プライバシーポリシー' })).toHaveAttribute(
      'href',
      '/privacy'
    )
  })

  // アカウント選択の強制は loginWithGoogle 側で常に行うため、画面からは引数を渡さない
  it('主操作から引数なしでログインを開始する', async () => {
    mockLoginWithGoogle.mockResolvedValue(undefined)

    render(<SignInPage />)
    await act(async () => {
      await userEvent.click(await screen.findByRole('button', { name: 'Googleで続行' }))
    })

    expect(mockLoginWithGoogle).toHaveBeenCalledWith()
  })

  // 主ボタンが常にアカウント選択を出すため、別導線は不要になった
  it('「別のGoogleアカウントを使う」を表示しない', async () => {
    render(<SignInPage />)
    await screen.findByRole('button', { name: 'Googleで続行' })

    expect(
      screen.queryByRole('button', { name: '別のGoogleアカウントを使う' })
    ).not.toBeInTheDocument()
  })

  it('失敗時はエラーを表示し、再操作できる状態に戻す', async () => {
    mockLoginWithGoogle.mockRejectedValue(new Error('Googleログインがキャンセルされました'))

    render(<SignInPage />)
    await act(async () => {
      await userEvent.click(await screen.findByRole('button', { name: 'Googleで続行' }))
    })

    expect(await screen.findByText('Googleログインがキャンセルされました')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Googleで続行' })).toBeEnabled()
  })

  it('reason=timeout のときセッションタイムアウトを案内する', async () => {
    setSearchParams('reason=timeout')

    render(<SignInPage />)

    expect(
      await screen.findByText('タイムアウトしました。再度ログインしてください')
    ).toBeInTheDocument()
  })

  it('ボタン押下時にURLから reason パラメータを取り除く', async () => {
    setSearchParams('reason=timeout')
    mockLoginWithGoogle.mockResolvedValue(undefined)

    render(<SignInPage />)
    await act(async () => {
      await userEvent.click(await screen.findByRole('button', { name: 'Googleで続行' }))
    })

    await waitFor(() => {
      expect(mockReplace).toHaveBeenCalledWith('/signin')
    })
  })
})
