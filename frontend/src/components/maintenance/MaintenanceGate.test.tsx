import React from 'react'
import { act, render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import { MaintenanceGate, isMaintenanceExemptPath } from './MaintenanceGate'
import { resetMaintenance, setMaintenance } from '@/lib/maintenance/maintenanceStore'

let mockPathname = '/'
jest.mock('next/navigation', () => ({
  usePathname: () => mockPathname,
}))

beforeEach(() => {
  mockPathname = '/'
  resetMaintenance()
  // /maintenance.json は取得できない（未配置）環境として扱う
  global.fetch = jest.fn(async () => ({ ok: false })) as unknown as typeof fetch
})

const renderGate = () =>
  render(
    <MaintenanceGate>
      <div>通常の画面</div>
    </MaintenanceGate>
  )

describe('MaintenanceGate', () => {
  it('メンテナンスでなければ子要素をそのまま描画する', () => {
    renderGate()
    expect(screen.getByText('通常の画面')).toBeInTheDocument()
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })

  it('メンテナンス中は子要素の代わりにメンテナンス画面を出す', () => {
    renderGate()
    act(() => setMaintenance({ endsAt: '2026-10-25T06:00:00Z' }))
    expect(screen.queryByText('通常の画面')).not.toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'メンテナンス中です' })).toBeInTheDocument()
    expect(screen.getByText(/2026年10月25日 15:00ごろ/)).toBeInTheDocument()
  })

  it('/news はメンテナンス中も通常どおり開ける', () => {
    mockPathname = '/news'
    renderGate()
    act(() => setMaintenance({}))
    expect(screen.getByText('通常の画面')).toBeInTheDocument()
  })

  it('/maintenance.json が enabled: true ならメンテナンス画面を出す', async () => {
    global.fetch = jest.fn(async () => ({
      ok: true,
      json: async () => ({ enabled: true, endsAt: null }),
    })) as unknown as typeof fetch
    renderGate()
    expect(await screen.findByRole('heading', { name: 'メンテナンス中です' })).toBeInTheDocument()
  })
})

describe('isMaintenanceExemptPath', () => {
  it('お知らせ配下だけが対象', () => {
    expect(isMaintenanceExemptPath('/news')).toBe(true)
    expect(isMaintenanceExemptPath('/news/3')).toBe(true)
    expect(isMaintenanceExemptPath('/newsletter')).toBe(false)
    expect(isMaintenanceExemptPath('/')).toBe(false)
    expect(isMaintenanceExemptPath(null)).toBe(false)
  })
})
