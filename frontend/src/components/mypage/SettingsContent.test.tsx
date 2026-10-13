import React from 'react'
import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import { SettingsContent } from './SettingsContent'
import { useAuth } from '@/contexts/AuthContext'

jest.mock('@/contexts/AuthContext', () => ({ useAuth: jest.fn() }))
jest.mock('@/lib/api/account', () => ({ deleteAccount: jest.fn() }))

describe('SettingsContent のパンくず', () => {
  beforeEach(() => {
    ;(useAuth as jest.Mock).mockReturnValue({ user: { id: 1, name: 'テスト' }, logout: jest.fn() })
  })

  // マイページで開いていたタブへ戻す
  it('?tab で受け取ったタブ付きのマイページへ戻る', () => {
    window.history.replaceState(null, '', '/mypage/edit?tab=watching')
    render(<SettingsContent />)

    expect(screen.getByRole('link', { name: 'マイページ' })).toHaveAttribute(
      'href',
      '/mypage?tab=watching',
    )
  })

  it('?tab が無ければ視聴予定のマイページへ戻る', () => {
    window.history.replaceState(null, '', '/mypage/edit')
    render(<SettingsContent />)

    expect(screen.getByRole('link', { name: 'マイページ' })).toHaveAttribute(
      'href',
      '/mypage?tab=planned',
    )
  })
})
