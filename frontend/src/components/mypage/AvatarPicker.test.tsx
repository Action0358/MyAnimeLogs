import React from 'react'
import { render, screen, act } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import '@testing-library/jest-dom'
import { AvatarPicker } from './AvatarPicker'
import { PRESET_AVATAR_IDS } from '@/lib/avatars'

describe('AvatarPicker', () => {
  const onSelect = jest.fn()

  beforeEach(() => jest.clearAllMocks())

  it('プリセットをすべて選択肢として表示する', () => {
    render(<AvatarPicker selectedIcon={null} onSelect={onSelect} />)

    expect(screen.getAllByRole('radio')).toHaveLength(PRESET_AVATAR_IDS.length)
  })

  it('選択中のプリセットを aria-checked で示す', () => {
    render(<AvatarPicker selectedIcon="preset-03" onSelect={onSelect} />)

    const selected = screen.getByRole('radio', { name: 'アイコン 03' })
    expect(selected).toHaveAttribute('aria-checked', 'true')
    expect(screen.getByRole('radio', { name: 'アイコン 01' })).toHaveAttribute(
      'aria-checked',
      'false'
    )
  })

  it('クリックで選択したIDを通知する', async () => {
    render(<AvatarPicker selectedIcon={null} onSelect={onSelect} />)

    await act(async () => {
      await userEvent.click(screen.getByRole('radio', { name: 'アイコン 05' }))
    })

    expect(onSelect).toHaveBeenCalledWith('preset-05')
  })

  it('disabled のときは操作を受け付けない', async () => {
    render(<AvatarPicker selectedIcon={null} onSelect={onSelect} disabled />)

    const button = screen.getByRole('radio', { name: 'アイコン 01' })
    expect(button).toBeDisabled()

    await act(async () => {
      await userEvent.click(button)
    })
    expect(onSelect).not.toHaveBeenCalled()
  })

  // 静的アセットとして配信するため、APIサーバーのURLを前置してはならない
  it('画像はフロントエンドの静的パスを参照する', () => {
    const { container } = render(<AvatarPicker selectedIcon={null} onSelect={onSelect} />)

    const first = container.querySelector('img')
    expect(first).toHaveAttribute('src', '/avatars/preset-01.svg')
  })
})
