import React from 'react'
import { act, render, screen } from '@testing-library/react'
import { MaintenanceBackLink } from './MaintenanceBackLink'
import { resetMaintenance, setMaintenance } from '@/lib/maintenance/maintenanceStore'

describe('MaintenanceBackLink', () => {
  afterEach(() => resetMaintenance())

  it('通常時は何も出さない', () => {
    render(<MaintenanceBackLink />)
    expect(screen.queryByRole('link', { name: '戻る' })).not.toBeInTheDocument()
  })

  it('メンテナンス中だけ「戻る」を出し、トップ（メンテナンス画面）へ向ける', () => {
    render(<MaintenanceBackLink />)
    act(() => setMaintenance({}))
    expect(screen.getByRole('link', { name: '戻る' })).toHaveAttribute('href', '/')
  })
})
