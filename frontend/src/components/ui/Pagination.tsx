'use client'

import React from 'react'
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
  if (totalPages <= 1) return null

  const getVisiblePages = (): (number | 'ellipsis')[] => {
    const pages: (number | 'ellipsis')[] = []
    const half = Math.floor(maxVisiblePages / 2)

    let start = Math.max(1, currentPage - half)
    let end = Math.min(totalPages, currentPage + half)

    if (currentPage <= half) {
      end = Math.min(totalPages, maxVisiblePages)
    }
    if (currentPage > totalPages - half) {
      start = Math.max(1, totalPages - maxVisiblePages + 1)
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

  const visiblePages = getVisiblePages()

  const buttonBase =
    'inline-flex items-center justify-center min-w-[32px] sm:min-w-[40px] h-8 sm:h-10 px-1.5 sm:px-3 text-xs sm:text-sm font-medium rounded-md transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2'

  return (
    <nav
      className={cn('flex flex-wrap items-center justify-center gap-1', className)}
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

      {showPageNumbers &&
        visiblePages.map((page, index) =>
          page === 'ellipsis' ? (
            <span key={`ellipsis-${index}`} className="px-1 sm:px-2 text-gray-500 text-xs sm:text-sm">
              ...
            </span>
          ) : (
            <button
              key={page}
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
