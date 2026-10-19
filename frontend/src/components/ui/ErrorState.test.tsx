import { render, screen, fireEvent } from '@testing-library/react'
import { ErrorState, getErrorStateType, isNotFoundError } from './ErrorState'

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

describe('ErrorState type=notFound', () => {
  it('作品が見つからない文言を出し、ボタンは検索、サブリンクはトップへ向ける', () => {
    render(<ErrorState type="notFound" />)
    expect(screen.getByRole('alert')).toBeInTheDocument()
    expect(screen.getByRole('heading', { level: 1, name: '作品が見つかりません' })).toBeInTheDocument()
    expect(screen.getByText('削除されたか、URLが間違っている可能性があります。')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: '作品を検索' })).toHaveAttribute('href', '/search')
    expect(screen.getByRole('link', { name: 'トップへ戻る' })).toHaveAttribute('href', '/')
    // 再読み込みでは直らないので出さない
    expect(screen.queryByRole('button', { name: '再読み込み' })).not.toBeInTheDocument()
  })
})

describe('isNotFoundError', () => {
  it('axios / apiClient どちらの404も判定する', () => {
    expect(isNotFoundError({ isAxiosError: true, response: { status: 404 } })).toBe(true)
    expect(isNotFoundError({ status: 404 })).toBe(true)
  })
  it('404以外・エラーでない値は false', () => {
    expect(isNotFoundError({ isAxiosError: true, response: { status: 500 } })).toBe(false)
    expect(isNotFoundError({ status: 0 })).toBe(false)
    expect(isNotFoundError(null)).toBe(false)
  })
  it('getErrorStateType は404を notFound にしない（作品詳細以外では再読み込みを出すため）', () => {
    expect(getErrorStateType({ isAxiosError: true, response: { status: 404 } })).toBe('unknown')
  })
})
