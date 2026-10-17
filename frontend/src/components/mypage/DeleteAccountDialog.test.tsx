import React from 'react'
import { render, screen, act } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import '@testing-library/jest-dom'
import { DeleteAccountDialog } from './DeleteAccountDialog'

describe('DeleteAccountDialog', () => {
  const baseProps = {
    isOpen: true,
    isDeleting: false,
    onConfirm: jest.fn(),
    onCancel: jest.fn(),
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('閉じているときは何も描画しない', () => {
    render(<DeleteAccountDialog {...baseProps} isOpen={false} />)

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  // 何が消えるかを具体的に示すことが、この確認画面の主目的
  it('削除される対象と復元不可である旨を表示する', () => {
    render(<DeleteAccountDialog {...baseProps} />)

    expect(screen.getByRole('dialog')).toBeInTheDocument()
    expect(screen.getByText('復元はできません')).toBeInTheDocument()
    expect(screen.getByText(/お気に入りに登録した作品/)).toBeInTheDocument()
    expect(screen.getByText(/視聴予定・視聴中・視聴完了の記録/)).toBeInTheDocument()
  })

  // アプリ側の削除だけではGoogle側の連携が残るため、案内が必要
  it('Googleアカウント側の連携解除について案内する', () => {
    render(<DeleteAccountDialog {...baseProps} />)

    expect(screen.getByText(/Googleのアカウント設定から別途解除/)).toBeInTheDocument()
  })

  // 不可逆な操作をワンクリックで実行させないための確認ステップ
  it('確認チェックボックスを入れるまで削除ボタンを押せない', async () => {
    render(<DeleteAccountDialog {...baseProps} />)

    const deleteButton = screen.getByRole('button', { name: 'アカウントを削除' })
    expect(deleteButton).toBeDisabled()

    await act(async () => {
      await userEvent.click(screen.getByRole('checkbox'))
    })

    expect(deleteButton).toBeEnabled()
  })

  it('削除ボタンで onConfirm、キャンセルで onCancel を呼ぶ', async () => {
    render(<DeleteAccountDialog {...baseProps} />)

    await act(async () => {
      await userEvent.click(screen.getByRole('checkbox'))
      await userEvent.click(screen.getByRole('button', { name: 'アカウントを削除' }))
    })
    expect(baseProps.onConfirm).toHaveBeenCalledTimes(1)

    await act(async () => {
      await userEvent.click(screen.getByRole('button', { name: 'キャンセル' }))
    })
    expect(baseProps.onCancel).toHaveBeenCalledTimes(1)
  })

  // Enter の連打で誤って削除されないよう、初期フォーカスはキャンセル側に置く
  it('開いた直後のフォーカスはキャンセルボタンにある', () => {
    render(<DeleteAccountDialog {...baseProps} />)

    expect(screen.getByRole('button', { name: 'キャンセル' })).toHaveFocus()
  })

  it('削除中は両ボタンを無効化し、進行中であることを示す', () => {
    render(<DeleteAccountDialog {...baseProps} isDeleting />)

    expect(screen.getByRole('button', { name: '削除しています...' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'キャンセル' })).toBeDisabled()
  })

  it('エラーを表示する', () => {
    render(<DeleteAccountDialog {...baseProps} error="アカウントの削除に失敗しました" />)

    expect(screen.getByText('アカウントの削除に失敗しました')).toBeInTheDocument()
  })
})
