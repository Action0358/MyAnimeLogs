import React from 'react'
import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import { MaintenanceScreen, formatEndsAt } from './MaintenanceScreen'

describe('formatEndsAt', () => {
  it('ISO 8601 を JST の「◯年◯月◯日 HH:mmごろ」に整形する', () => {
    expect(formatEndsAt('2026-10-25T06:00:00Z')).toBe('2026年10月25日 15:00ごろ')
    expect(formatEndsAt('2026-10-25T15:30:00+09:00')).toBe('2026年10月25日 15:30ごろ')
  })

  it('日付をまたぐ場合もJSTで数える', () => {
    expect(formatEndsAt('2026-10-24T15:00:00Z')).toBe('2026年10月25日 00:00ごろ')
  })

  it('未指定・不正な値は null', () => {
    expect(formatEndsAt(undefined)).toBeNull()
    expect(formatEndsAt('')).toBeNull()
    expect(formatEndsAt('not-a-date')).toBeNull()
  })
})

describe('MaintenanceScreen', () => {
  it('role="status" でメンテナンス中の案内を表示する', () => {
    render(<MaintenanceScreen />)
    expect(screen.getByRole('status')).toBeInTheDocument()
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'メンテナンス中です' })).toBeInTheDocument()
    expect(screen.queryByRole('button')).not.toBeInTheDocument()
  })

  it('endsAt があるときだけ終了予定を出す', () => {
    const { rerender } = render(<MaintenanceScreen endsAt="2026-10-25T06:00:00Z" />)
    expect(screen.getByText(/終了予定：2026年10月25日 15:00ごろ/)).toBeInTheDocument()
    rerender(<MaintenanceScreen />)
    expect(screen.queryByText(/終了予定/)).not.toBeInTheDocument()
    rerender(<MaintenanceScreen endsAt="broken" />)
    expect(screen.queryByText(/終了予定/)).not.toBeInTheDocument()
  })

  it('お知らせへのリンクは既定で /news、指定で差し替えられる', () => {
    const { rerender } = render(<MaintenanceScreen />)
    expect(screen.getByRole('link', { name: 'お知らせを見る' })).toHaveAttribute('href', '/news')
    rerender(<MaintenanceScreen newsHref="/status.html" />)
    expect(screen.getByRole('link', { name: 'お知らせを見る' })).toHaveAttribute('href', '/status.html')
  })

  it('ナビ・ログイン・お問い合わせは出さない', () => {
    render(<MaintenanceScreen />)
    expect(screen.queryByText('ログイン')).not.toBeInTheDocument()
    expect(screen.queryByText('お問い合わせ')).not.toBeInTheDocument()
    expect(screen.queryByText('今期アニメ')).not.toBeInTheDocument()
  })
})
