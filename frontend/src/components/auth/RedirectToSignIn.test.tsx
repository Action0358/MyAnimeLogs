import React from 'react'
import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import { RedirectToSignIn } from './RedirectToSignIn'
import { useRouter, useSearchParams } from 'next/navigation'

jest.mock('next/navigation')

const mockUseRouter = useRouter as jest.Mock
const mockUseSearchParams = useSearchParams as jest.Mock

describe('RedirectToSignIn', () => {
  const mockReplace = jest.fn()

  beforeEach(() => {
    jest.clearAllMocks()
    mockUseRouter.mockReturnValue({ replace: mockReplace })
  })

  it('/signin へ転送する', () => {
    mockUseSearchParams.mockReturnValue(new URLSearchParams(''))

    render(<RedirectToSignIn />)

    expect(mockReplace).toHaveBeenCalledWith('/signin')
  })

  // セッション切れの案内が消えないよう、クエリは引き継ぐ必要がある
  it('クエリパラメータを引き継ぐ', () => {
    mockUseSearchParams.mockReturnValue(new URLSearchParams('reason=timeout'))

    render(<RedirectToSignIn />)

    expect(mockReplace).toHaveBeenCalledWith('/signin?reason=timeout')
  })

  it('転送中であることを表示する', () => {
    mockUseSearchParams.mockReturnValue(new URLSearchParams(''))

    render(<RedirectToSignIn />)

    expect(screen.getByText('ログイン画面へ移動しています...')).toBeInTheDocument()
  })
})
