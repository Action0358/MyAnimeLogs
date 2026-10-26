import React from 'react'
import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import { Pagination } from './Pagination'

describe('Pagination', () => {
  it('総ページ数が 1 以下なら何も描画しない', () => {
    const { container } = render(
      <Pagination currentPage={1} totalPages={1} onPageChange={jest.fn()} />
    )
    expect(container).toBeEmptyDOMElement()
  })

  // 早期リターンをフックより後へ移したため、総ページ数が 1 と 2 以上の間で
  // 行き来する再描画を確認しておく
  it('総ページ数が 1 から増えても再描画できる', () => {
    const { rerender } = render(
      <Pagination currentPage={1} totalPages={1} onPageChange={jest.fn()} />
    )

    rerender(<Pagination currentPage={1} totalPages={5} onPageChange={jest.fn()} />)

    expect(screen.getByRole('navigation', { name: 'ページネーション' })).toBeInTheDocument()

    rerender(<Pagination currentPage={1} totalPages={1} onPageChange={jest.fn()} />)

    expect(screen.queryByRole('navigation')).not.toBeInTheDocument()
  })
})
