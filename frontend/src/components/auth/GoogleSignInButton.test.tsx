import React from 'react'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import '@testing-library/jest-dom'
import { GoogleSignInButton } from './GoogleSignInButton'

describe('GoogleSignInButton', () => {
  it('ラベルを表示し、クリックで onClick を呼ぶ', async () => {
    const onClick = jest.fn()

    render(<GoogleSignInButton label="Googleでログイン" onClick={onClick} />)

    const button = screen.getByRole('button', { name: 'Googleでログイン' })
    await userEvent.click(button)

    expect(onClick).toHaveBeenCalledTimes(1)
  })

  it('ローディング中は接続中の表示に切り替わり、操作を受け付けない', async () => {
    const onClick = jest.fn()

    render(<GoogleSignInButton label="Googleでログイン" onClick={onClick} isLoading />)

    const button = screen.getByRole('button')
    expect(button).toBeDisabled()
    expect(screen.getByText('Googleに接続しています...')).toBeInTheDocument()
    expect(screen.queryByText('Googleでログイン')).not.toBeInTheDocument()

    // 二重にOAuthフローを開始しないこと
    await userEvent.click(button)
    expect(onClick).not.toHaveBeenCalled()
  })

  it('フォーム内に置かれても送信を発火させない type=button である', () => {
    render(<GoogleSignInButton label="Googleで登録" onClick={jest.fn()} />)

    expect(screen.getByRole('button')).toHaveAttribute('type', 'button')
  })
})
