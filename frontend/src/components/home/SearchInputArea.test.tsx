import React from 'react'
import { render, screen, fireEvent } from '@testing-library/react'
import '@testing-library/jest-dom'
import SearchInputArea from './SearchInputArea'

const mockPush = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({ push: mockPush }),
}))

const HISTORY_KEY = 'anime_search_history'
const storedHistory = () => JSON.parse(localStorage.getItem(HISTORY_KEY) ?? '[]')

describe('トップページの検索欄の「最近の検索」', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    localStorage.clear()
  })

  it('フォーカスすると作品検索画面と共通の履歴を表示する', async () => {
    localStorage.setItem(HISTORY_KEY, JSON.stringify(['葬送のフリーレン', '薬屋のひとりごと']))
    render(<SearchInputArea activeTab="keyword" />)

    fireEvent.focus(screen.getByPlaceholderText('作品を探す'))

    expect(await screen.findByRole('listbox', { name: '最近の検索' })).toBeInTheDocument()
    expect(screen.getAllByRole('option').map((o) => o.textContent)).toEqual([
      '葬送のフリーレン',
      '薬屋のひとりごと',
    ])
  })

  it('履歴がなければ候補リストを出さない', () => {
    render(<SearchInputArea activeTab="keyword" />)
    fireEvent.focus(screen.getByPlaceholderText('作品を探す'))

    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
  })

  it('履歴を選ぶとそのキーワードで作品検索へ進む', async () => {
    localStorage.setItem(HISTORY_KEY, JSON.stringify(['葬送のフリーレン']))
    render(<SearchInputArea activeTab="keyword" />)
    fireEvent.focus(screen.getByPlaceholderText('作品を探す'))

    fireEvent.click(await screen.findByRole('option', { name: '葬送のフリーレン' }))

    expect(mockPush).toHaveBeenCalledWith(
      `/search?keyword=${encodeURIComponent('葬送のフリーレン')}`
    )
  })

  it('↓とEnterで履歴を選べる', async () => {
    localStorage.setItem(HISTORY_KEY, JSON.stringify(['葬送のフリーレン', '薬屋のひとりごと']))
    render(<SearchInputArea activeTab="keyword" />)
    const input = screen.getByPlaceholderText('作品を探す')
    fireEvent.focus(input)
    await screen.findByRole('listbox')

    fireEvent.keyDown(input, { key: 'ArrowDown' })
    fireEvent.keyDown(input, { key: 'ArrowDown' })
    fireEvent.keyDown(input, { key: 'Enter' })

    expect(mockPush).toHaveBeenCalledTimes(1)
    expect(mockPush).toHaveBeenCalledWith(
      `/search?keyword=${encodeURIComponent('薬屋のひとりごと')}`
    )
  })

  it('入力して検索すると履歴の先頭に追加する', () => {
    localStorage.setItem(HISTORY_KEY, JSON.stringify(['薬屋のひとりごと']))
    render(<SearchInputArea activeTab="keyword" />)
    const input = screen.getByPlaceholderText('作品を探す')

    fireEvent.change(input, { target: { value: '  鬼滅の刃 ' } })
    fireEvent.keyDown(input, { key: 'Enter' })

    expect(mockPush).toHaveBeenCalledWith(`/search?keyword=${encodeURIComponent('鬼滅の刃')}`)
    expect(storedHistory()).toEqual(['鬼滅の刃', '薬屋のひとりごと'])
  })

  it('「すべて削除」で履歴を消す', async () => {
    localStorage.setItem(HISTORY_KEY, JSON.stringify(['葬送のフリーレン']))
    render(<SearchInputArea activeTab="keyword" />)
    fireEvent.focus(screen.getByPlaceholderText('作品を探す'))

    fireEvent.click(await screen.findByRole('button', { name: 'すべて削除' }))

    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
    expect(localStorage.getItem(HISTORY_KEY)).toBeNull()
  })

  it('ドロップダウンには新しい順に最大5件だけ出す', async () => {
    const history = ['1', '2', '3', '4', '5', '6', '7']
    localStorage.setItem(HISTORY_KEY, JSON.stringify(history))
    render(<SearchInputArea activeTab="keyword" />)
    fireEvent.focus(screen.getByPlaceholderText('作品を探す'))

    await screen.findByRole('listbox')
    expect(screen.getAllByRole('option').map((o) => o.textContent)).toEqual([
      '1',
      '2',
      '3',
      '4',
      '5',
    ])
  })

  it('×で1件だけ削除し、その語では検索しない', async () => {
    localStorage.setItem(HISTORY_KEY, JSON.stringify(['葬送のフリーレン', '薬屋のひとりごと']))
    render(<SearchInputArea activeTab="keyword" />)
    fireEvent.focus(screen.getByPlaceholderText('作品を探す'))

    fireEvent.click(
      await screen.findByRole('button', { name: '「葬送のフリーレン」を履歴から削除' })
    )

    expect(mockPush).not.toHaveBeenCalled()
    expect(screen.getAllByRole('option').map((o) => o.textContent)).toEqual(['薬屋のひとりごと'])
    expect(storedHistory()).toEqual(['薬屋のひとりごと'])
  })
})
