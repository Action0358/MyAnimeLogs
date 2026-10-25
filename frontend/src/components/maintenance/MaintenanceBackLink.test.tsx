import React from 'react'
import { act, render, screen } from '@testing-library/react'
import { MaintenanceBackLink } from './MaintenanceBackLink'
import { resetMaintenance, setMaintenance } from '@/lib/maintenance/maintenanceStore'

describe('MaintenanceBackLink', () => {
  afterEach(() => resetMaintenance())

  it('通常時は、エッジの目印があるときだけ表示される印を付けて出す（CSS で隠れる）', () => {
    render(<MaintenanceBackLink />)
    expect(screen.getByRole('link', { name: '戻る' })).toHaveAttribute(
      'data-show-when',
      'maintenance'
    )
  })

  it('メンテナンス中は印なしで「戻る」を出し、トップ（メンテナンス画面）へ向ける', () => {
    render(<MaintenanceBackLink />)
    act(() => setMaintenance({}))
    const link = screen.getByRole('link', { name: '戻る' })
    expect(link).toHaveAttribute('href', '/')
    expect(link).not.toHaveAttribute('data-show-when')
  })
})
