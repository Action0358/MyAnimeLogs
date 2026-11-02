import React from 'react'
import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import { ProtectedRoute } from './ProtectedRoute'
import { useAuth } from '@/contexts/AuthContext'
import { useRouter } from 'next/navigation'

jest.mock('next/navigation')
jest.mock('@/contexts/AuthContext')

const mockUseAuth = useAuth as jest.Mock
const mockUseRouter = useRouter as jest.Mock

const LOGGED_IN = { isAuthenticated: true, isLoading: false, isManualLogout: false }
const LOGGED_OUT = { isAuthenticated: false, isLoading: false, isManualLogout: false }

describe('ProtectedRoute', () => {
  const mockReplace = jest.fn()

  // 同じ要素を rerender すると React が描画を省くため、毎回作り直す
  const ui = () => (
    <ProtectedRoute>
      <div>保護された内容</div>
    </ProtectedRoute>
  )

  beforeEach(() => {
    jest.clearAllMocks()
    mockUseRouter.mockReturnValue({ replace: mockReplace })
  })

  it('ログイン中は内容を表示し、転送しない', () => {
    mockUseAuth.mockReturnValue(LOGGED_IN)

    render(ui())

    expect(screen.getByText('保護された内容')).toBeInTheDocument()
    expect(mockReplace).not.toHaveBeenCalled()
  })

  it('未ログインで開くとログイン画面へ送る', () => {
    mockUseAuth.mockReturnValue(LOGGED_OUT)

    render(ui())

    expect(mockReplace).toHaveBeenCalledTimes(1)
    expect(mockReplace.mock.calls[0][0]).toMatch(/^\/signin/)
  })

  it('開いたまま自分でログアウトすると、ログイン画面を経由せずトップへ移す', () => {
    mockUseAuth.mockReturnValue(LOGGED_IN)
    const { rerender } = render(ui())

    mockUseAuth.mockReturnValue({ ...LOGGED_OUT, isManualLogout: true })
    rerender(ui())

    expect(mockReplace).toHaveBeenCalledTimes(1)
    expect(mockReplace).toHaveBeenCalledWith('/')
  })

  it('タイムアウトなど自分の操作でないログアウトではトップへ移さない', () => {
    mockUseAuth.mockReturnValue(LOGGED_IN)
    const { rerender } = render(ui())

    mockUseAuth.mockReturnValue(LOGGED_OUT)
    rerender(ui())

    expect(mockReplace).not.toHaveBeenCalledWith('/')
  })

  it('ログアウト後に「戻る」で来た場合はログイン画面へ送る', () => {
    mockUseAuth.mockReturnValue({ ...LOGGED_OUT, isManualLogout: true })

    render(ui())

    expect(mockReplace).toHaveBeenCalledTimes(1)
    expect(mockReplace.mock.calls[0][0]).toMatch(/^\/signin/)
  })
})
