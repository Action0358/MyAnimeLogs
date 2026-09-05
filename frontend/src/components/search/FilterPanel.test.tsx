import React from 'react'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import '@testing-library/jest-dom'
import { FilterPanel } from './FilterPanel'

/**
 * ジャンルの複数選択を検証する（#811）。
 *
 * 以前は単一の <select> だったため、トップページから
 * genre=アクション,ドラマ で来ると <option> のどれにも一致せず
 * 「すべて」表示に戻り、選択が消えたように見えていた。
 */
describe('FilterPanel のジャンル絞り込み', () => {
  const baseProps = {
    onSeasonChange: jest.fn(),
    onGenresChange: jest.fn(),
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  const openGenreDropdown = async (label: string) => {
    await userEvent.click(screen.getByRole('button', { name: label }))
  }

  it('未選択なら「すべて」と表示し、初期状態では選択肢を畳んでいる', () => {
    render(<FilterPanel {...baseProps} />)

    expect(screen.getByRole('button', { name: 'すべて' })).toBeInTheDocument()
    expect(screen.queryByRole('checkbox', { name: 'アクション' })).not.toBeInTheDocument()
  })

  it('選択中はジャンル名をボタンに表示する', () => {
    render(<FilterPanel {...baseProps} selectedGenres={['アクション', 'ドラマ']} />)

    expect(screen.getByRole('button', { name: 'アクション、ドラマ' })).toBeInTheDocument()
  })

  it('ボタンを開くとジャンルを選べる', async () => {
    render(<FilterPanel {...baseProps} selectedGenres={['アクション']} />)
    await openGenreDropdown('アクション')

    expect(screen.getByRole('checkbox', { name: 'アクション' })).toBeChecked()

    await userEvent.click(screen.getByRole('checkbox', { name: 'SF' }))

    expect(baseProps.onGenresChange).toHaveBeenCalledWith(['アクション', 'SF'])
  })

  it('選択済みを押すとその1つだけ外れる', async () => {
    render(<FilterPanel {...baseProps} selectedGenres={['アクション', 'ドラマ']} />)
    await openGenreDropdown('アクション、ドラマ')

    await userEvent.click(screen.getByRole('checkbox', { name: 'アクション' }))

    expect(baseProps.onGenresChange).toHaveBeenCalledWith(['ドラマ'])
  })

  it('選択中があれば「すべて解除」を出す', async () => {
    render(<FilterPanel {...baseProps} selectedGenres={['アクション']} />)

    await userEvent.click(screen.getByRole('button', { name: 'すべて解除' }))

    expect(baseProps.onGenresChange).toHaveBeenCalledWith([])
  })

  it('未選択のときは「すべて解除」を出さない', () => {
    render(<FilterPanel {...baseProps} />)

    expect(screen.queryByRole('button', { name: 'すべて解除' })).not.toBeInTheDocument()
  })
})
