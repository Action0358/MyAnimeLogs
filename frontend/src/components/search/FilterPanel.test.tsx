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
    onStatusChange: jest.fn(),
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('inline（デスクトップ）', () => {
    const openGenreMenu = async () => {
      await userEvent.click(screen.getByRole('button', { name: /^ジャンル：/ }))
    }

    it('未選択なら「すべて」と表示する', () => {
      render(<FilterPanel {...baseProps} variant="inline" />)

      expect(screen.getByRole('button', { name: 'ジャンル：すべて' })).toBeInTheDocument()
    })

    it('複数選択されていても「すべて」に戻らない', () => {
      render(
        <FilterPanel {...baseProps} selectedGenres={['アクション', 'ドラマ']} variant="inline" />
      )

      // 表示は幅を一定に保つため件数、読み上げには選択中のジャンルを出す
      const trigger = screen.getByRole('button', { name: 'ジャンル：アクション、ドラマ' })
      expect(trigger).toHaveTextContent('ジャンル2件')
      expect(screen.queryByRole('button', { name: 'ジャンル：すべて' })).not.toBeInTheDocument()
    })

    it('1つだけならジャンル名をそのまま表示する', () => {
      render(<FilterPanel {...baseProps} selectedGenres={['アクション']} variant="inline" />)

      // 他のチップと同じく、選択中は項目名を省いて値だけ見せる
      expect(screen.getByRole('button', { name: 'ジャンル：アクション' })).toHaveTextContent(
        'アクション'
      )
    })

    it('選択中のジャンルにチェックが入っている', async () => {
      render(<FilterPanel {...baseProps} selectedGenres={['アクション']} variant="inline" />)
      await openGenreMenu()

      expect(screen.getByRole('checkbox', { name: 'アクション' })).toBeChecked()
      expect(screen.getByRole('checkbox', { name: 'ドラマ' })).not.toBeChecked()
    })

    it('選択を追加すると既存の選択を保ったまま通知する', async () => {
      render(<FilterPanel {...baseProps} selectedGenres={['アクション']} variant="inline" />)
      await openGenreMenu()

      await userEvent.click(screen.getByRole('checkbox', { name: 'ドラマ' }))

      expect(baseProps.onGenresChange).toHaveBeenCalledWith(['アクション', 'ドラマ'])
    })

    it('選択済みを押すとその1つだけ外れる', async () => {
      render(
        <FilterPanel {...baseProps} selectedGenres={['アクション', 'ドラマ']} variant="inline" />
      )
      await openGenreMenu()

      await userEvent.click(screen.getByRole('checkbox', { name: 'アクション' }))

      expect(baseProps.onGenresChange).toHaveBeenCalledWith(['ドラマ'])
    })

    it('すべて解除すると空配列を通知する', async () => {
      render(
        <FilterPanel {...baseProps} selectedGenres={['アクション', 'ドラマ']} variant="inline" />
      )
      await openGenreMenu()

      await userEvent.click(screen.getByRole('button', { name: '選択をすべて解除' }))

      expect(baseProps.onGenresChange).toHaveBeenCalledWith([])
    })

    it('未選択のときは解除ボタンを出さない', async () => {
      render(<FilterPanel {...baseProps} variant="inline" />)
      await openGenreMenu()

      expect(screen.queryByRole('button', { name: '選択をすべて解除' })).not.toBeInTheDocument()
    })
  })

  describe('sheet（モバイル）', () => {
    it('開閉なしでジャンルを直接選べる', async () => {
      render(<FilterPanel {...baseProps} selectedGenres={['アクション']} variant="sheet" />)

      expect(screen.getByRole('checkbox', { name: 'アクション' })).toBeChecked()

      await userEvent.click(screen.getByRole('checkbox', { name: 'SF' }))

      expect(baseProps.onGenresChange).toHaveBeenCalledWith(['アクション', 'SF'])
    })

    it('選択中があれば「すべて解除」を出す', async () => {
      render(<FilterPanel {...baseProps} selectedGenres={['アクション']} variant="sheet" />)

      await userEvent.click(screen.getByRole('button', { name: 'すべて解除' }))

      expect(baseProps.onGenresChange).toHaveBeenCalledWith([])
    })
  })
})
