import React from 'react'
import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import AccountDeletedPage from './page'

describe('AccountDeletedPage', () => {
  it('削除完了を明示する', () => {
    render(<AccountDeletedPage />)

    expect(screen.getByRole('heading', { name: 'アカウントを削除しました' })).toBeInTheDocument()
  })

  // 再ログインで新規アカウントが作られることを事前に伝えないと
  // 「削除したのに復活した」と誤解される
  it('再ログイン時は新しいアカウントになり記録が戻らないことを説明する', () => {
    render(<AccountDeletedPage />)

    expect(screen.getByText(/新しいアカウントとして登録されます/)).toBeInTheDocument()
    expect(screen.getByText(/復元されません/)).toBeInTheDocument()
  })

  it('Googleアカウント側の連携解除について案内する', () => {
    render(<AccountDeletedPage />)

    expect(screen.getByText(/Googleのアカウント設定画面から解除/)).toBeInTheDocument()
  })

  it('トップページへの導線と問い合わせ先を表示する', () => {
    render(<AccountDeletedPage />)

    expect(screen.getByRole('link', { name: 'トップページへ' })).toHaveAttribute('href', '/')
    // 削除後はログインできないため、ログイン不要で送れるフォームへ案内する。
    // メールアドレスを直接出すとメーラーが必要になり、退会直後の導線として重い
    expect(screen.getByRole('link', { name: 'お問い合わせフォーム' })).toHaveAttribute(
      'href',
      '/contact'
    )
  })
})
