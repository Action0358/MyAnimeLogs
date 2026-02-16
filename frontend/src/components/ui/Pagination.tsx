'use client'

import React, { useMemo } from 'react'
import { cn } from '@/lib/utils'

export interface PaginationProps {
  currentPage: number
  totalPages: number
  onPageChange: (page: number) => void
  showPageNumbers?: boolean
  maxVisiblePages?: number
  className?: string
}

export const Pagination: React.FC<PaginationProps> = ({
  currentPage,
  totalPages,
  onPageChange,
  showPageNumbers = true,
  maxVisiblePages = 5,
  className,
}) => {
  // 早期リターン（不要な計算を防ぐ）
  if (totalPages <= 1) return null

  // ページ番号の計算をメモ化（showPageNumbersがfalseの場合は計算しない）
  const { mobilePages, desktopPages } = useMemo(() => {
    if (!showPageNumbers) {
      return { mobilePages: [], desktopPages: [] }
    }

    const getVisiblePages = (max: number): (number | 'ellipsis')[] => {
      const pages: (number | 'ellipsis')[] = []
      const half = Math.floor(max / 2)

      let start = Math.max(1, currentPage - half)
      let end = Math.min(totalPages, currentPage + half)

      if (currentPage <= half) {
        end = Math.min(totalPages, max)
      }
      if (currentPage > totalPages - half) {
        start = Math.max(1, totalPages - max + 1)
      }

      if (start > 1) {
        pages.push(1)
        if (start > 2) pages.push('ellipsis')
      }

      for (let i = start; i <= end; i++) {
        pages.push(i)
      }

      if (end < totalPages) {
        if (end < totalPages - 1) pages.push('ellipsis')
        pages.push(totalPages)
      }

      return pages
    }

    return {
      mobilePages: getVisiblePages(3),
      desktopPages: getVisiblePages(maxVisiblePages),
    }
  }, [currentPage, totalPages, maxVisiblePages, showPageNumbers])

  const buttonBase =
    'inline-flex items-center justify-center min-w-[32px] sm:min-w-[40px] h-8 sm:h-10 px-1.5 sm:px-3 text-xs sm:text-sm font-medium rounded-md transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2'

  // ページ番号のレンダリングを共通化
  const renderPageNumbers = (pages: (number | 'ellipsis')[], keyPrefix: string) => (
    pages.map((page, index) =>
      page === 'ellipsis' ? (
        <span
          key={`ellipsis-${keyPrefix}-${index}`}
          className="px-1 sm:px-2 text-gray-500 text-xs sm:text-sm"
          aria-hidden="true"
        >
          ...
        </span>
      ) : (
        <button
          key={`${keyPrefix}-${page}`}
          onClick={() => onPageChange(page)}
          className={cn(
            buttonBase,
            page === currentPage
              ? 'bg-blue-600 text-white border border-blue-600'
              : 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'
          )}
          aria-current={page === currentPage ? 'page' : undefined}
        >
          {page}
        </button>
      )
    )
  )

  return (
    <nav
      className={cn('flex items-center justify-center gap-1', className)}
      aria-label="ページネーション"
    >
      {/* 前のページへ */}
      <button
        onClick={() => onPageChange(currentPage - 1)}
        disabled={currentPage === 1}
        className={cn(
          buttonBase,
          'border border-gray-300',
          currentPage === 1
            ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
            : 'bg-white text-gray-700 hover:bg-gray-50'
        )}
        aria-label="前のページ"
      >
        <svg className="w-4 h-4 sm:w-5 sm:h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
        </svg>
      </button>

      {/* ページ番号 */}
      {showPageNumbers && (
        <>
          <div className="flex items-center gap-1 sm:hidden">
            {renderPageNumbers(mobilePages, 'mobile')}
          </div>
          <div className="hidden sm:flex items-center gap-1">
            {renderPageNumbers(desktopPages, 'desktop')}
          </div>
        </>
      )}

      {/* 次のページへ */}
      <button
        onClick={() => onPageChange(currentPage + 1)}
        disabled={currentPage === totalPages}
        className={cn(
          buttonBase,
          'border border-gray-300',
          currentPage === totalPages
            ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
            : 'bg-white text-gray-700 hover:bg-gray-50'
        )}
        aria-label="次のページ"
      >
        <svg className="w-4 h-4 sm:w-5 sm:h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
        </svg>
      </button>
    </nav>
  )
}

export default Pagination
