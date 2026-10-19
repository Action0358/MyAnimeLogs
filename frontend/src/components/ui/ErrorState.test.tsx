import { render, screen, fireEvent } from '@testing-library/react'
import { ErrorState, getErrorStateType } from './ErrorState'

describe('ErrorState', () => {
  it('network: 接続エラーの文言とトップへのリンクを表示する', () => {
    render(<ErrorState type="network" onRetry={() => {}} />)
    expect(screen.getByRole('alert')).toBeInTheDocument()
    expect(screen.getByText('接続できませんでした')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'トップへ戻る' })).toHaveAttribute('href', '/')
  })

  it('unknown: 読み込みエラーの文言とお問い合わせリンクを表示する', () => {
    render(<ErrorState type="unknown" onRetry={() => {}} />)
    expect(screen.getByText('読み込めませんでした')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: '解決しない場合はお問い合わせ' })).toHaveAttribute('href', '/contact')
  })

  it('再読み込みボタンで onRetry が呼ばれる', () => {
    const onRetry = jest.fn()
    render(<ErrorState type="unknown" onRetry={onRetry} />)
    fireEvent.click(screen.getByRole('button', { name: '再読み込み' }))
    expect(onRetry).toHaveBeenCalledTimes(1)
  })
})

describe('getErrorStateType', () => {
  it('レスポンスなしのエラーは network', () => {
    expect(getErrorStateType({ isAxiosError: true })).toBe('network')
    expect(getErrorStateType(new TypeError('Failed to fetch'))).toBe('network')
    expect(getErrorStateType({ status: 0 })).toBe('network')
  })
  it('レスポンスありのエラーは unknown', () => {
    expect(getErrorStateType({ isAxiosError: true, response: { status: 500 } })).toBe('unknown')
    expect(getErrorStateType(new Error('x'))).toBe('unknown')
  })
})

describe('ErrorState variant=compact', () => {
  it('サブリンクを出さず、再読み込みを呼べる', () => {
    const onRetry = jest.fn()
    render(<ErrorState variant="compact" type="network" onRetry={onRetry} />)
    expect(screen.getByRole('alert')).toBeInTheDocument()
    expect(screen.queryByRole('link')).not.toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: '再読み込み' }))
    expect(onRetry).toHaveBeenCalledTimes(1)
  })
})
