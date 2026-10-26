import React from 'react'
import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import PrivacyPage from './page'

/**
 * アクセス解析（Cloudflare Web Analytics）の記載が実態と一致していることを
 * 検証する（#868）。導入したのに記載を忘れる、または記載したのに実際には
 * 未導入、のどちらのずれも防ぐ。
 */
describe('プライバシーポリシー', () => {
  it('外部サービスの利用にアクセス解析を明記する', () => {
    render(<PrivacyPage />)

    expect(
      screen.getByText(/Cloudflareのアクセス解析機能（Cloudflare Web Analytics）/)
    ).toBeInTheDocument()
  })

  it('Cookieを使用しない解析であることを明記する', () => {
    render(<PrivacyPage />)

    expect(
      screen.getByText(/アクセス解析（Cloudflare Web Analytics）はCookieを使用せず/)
    ).toBeInTheDocument()
  })

  it('ログイン維持用Cookieの記載は維持する', () => {
    render(<PrivacyPage />)

    expect(
      screen.getByText(/ログイン状態を維持するためにブラウザのローカルストレージおよびCookieを使用/)
    ).toBeInTheDocument()
  })
})
