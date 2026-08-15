import React from 'react'
import { render, screen, act } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import '@testing-library/jest-dom'
import RegisterPage from './page'
import { useAuth } from '@/contexts/AuthContext'

jest.mock('@/contexts/AuthContext', () => ({
  useAuth: jest.fn(),
  DEFAULT_LOGIN_ERROR: 'ログイン処理中にエラーが発生しました。しばらく時間をおいて再度お試しください',
}))
jest.mock('@/components/auth/AuthPosterPanel', () => ({
  AuthPosterPanel: ({
    headline,
    children,
  }: {
    headline: React.ReactNode
    children?: React.ReactNode
  }) => (
    <div>
      {headline}
      {children}
    </div>
  ),
}))

const mockUseAuth = useAuth as jest.Mock

describe('RegisterPage', () => {
  const mockLoginWithGoogle = jest.fn()

  beforeEach(() => {
    jest.clearAllMocks()
    mockUseAuth.mockReturnValue({ loginWithGoogle: mockLoginWithGoogle })
  })

  it('Google登録ボタンとログインリンクを表示する', () => {
    render(<RegisterPage />)

    expect(screen.getByRole('button', { name: 'Googleで登録' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'ログイン' })).toHaveAttribute('href', '/login')
  })

  // #738 で撤廃した項目。表示名は Google プロフィールから自動設定される
  it('ユーザー名の入力欄を表示しない', () => {
    render(<RegisterPage />)

    expect(screen.queryByLabelText(/ユーザー名/)).not.toBeInTheDocument()
    expect(screen.queryByPlaceholderText('マイアニメログ')).not.toBeInTheDocument()
  })

  it('メールアドレス・パスワードの入力欄を表示しない', () => {
    render(<RegisterPage />)

    expect(screen.queryByLabelText(/メールアドレス/)).not.toBeInTheDocument()
    expect(screen.queryByLabelText(/パスワード/)).not.toBeInTheDocument()
  })

  it('利用規約とプライバシーポリシーへのリンクを表示する', () => {
    render(<RegisterPage />)

    expect(screen.getByRole('link', { name: '利用規約' })).toHaveAttribute('href', '/terms')
    expect(screen.getByRole('link', { name: 'プライバシーポリシー' })).toHaveAttribute(
      'href',
      '/privacy'
    )
  })

  it('ボタン押下で loginWithGoogle を呼ぶ', async () => {
    mockLoginWithGoogle.mockResolvedValue(undefined)

    render(<RegisterPage />)
    await act(async () => {
      await userEvent.click(screen.getByRole('button', { name: 'Googleで登録' }))
    })

    expect(mockLoginWithGoogle).toHaveBeenCalledTimes(1)
  })

  // 別アカウントで登録したいユーザーが、Googleにサインイン済みの既存アカウントで
  // 無言でログインしてしまうのを防ぐ
  it('アカウント選択画面を強制するオプション付きで呼ぶ', async () => {
    mockLoginWithGoogle.mockResolvedValue(undefined)

    render(<RegisterPage />)
    await act(async () => {
      await userEvent.click(screen.getByRole('button', { name: 'Googleで登録' }))
    })

    expect(mockLoginWithGoogle).toHaveBeenCalledWith({ forceAccountSelection: true })
  })

  it('失敗時はエラーを表示し、再操作できる状態に戻す', async () => {
    mockLoginWithGoogle.mockRejectedValue(new Error('Googleログインがキャンセルされました'))

    render(<RegisterPage />)
    await act(async () => {
      await userEvent.click(screen.getByRole('button', { name: 'Googleで登録' }))
    })

    expect(await screen.findByText('Googleログインがキャンセルされました')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Googleで登録' })).toBeEnabled()
  })
})
