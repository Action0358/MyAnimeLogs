'use client'

import React, { useMemo } from 'react'
import { cn } from '@/lib/utils'

export interface PaginationProps {
  currentPage: number
  totalPages: number
  onPageChange: (page: number) => void
  showPageNumbers?: boolean
  /** モバイル表示時の最大ページ数（デフォルト: 3） */
  mobileMaxPages?: number
  /** PC表示時の最大ページ数（デフォルト: 10） */
  desktopMaxPages?: number
  className?: string
}

export const Pagination: React.FC<PaginationProps> = ({
  currentPage,
  totalPages,
  onPageChange,
  showPageNumbers = true,
  mobileMaxPages = 3,
  desktopMaxPages = 10,
  className,
}) => {
  // 早期リターン（不要な計算を防ぐ）
  if (totalPages <= 1) return null

  // ページ番号の計算をメモ化
  const { mobilePages, desktopPages } = useMemo(() => {
    if (!showPageNumbers) return { mobilePages: [], desktopPages: [] }

    const getVisiblePages = (maxPages: number): number[] => {
      // 表示するページ数を実際のページ数に制限
      const visibleCount = Math.min(maxPages, totalPages)
      
      // 現在ページを中央に配置するための開始位置を計算
      const half = Math.floor(visibleCount / 2)
      const idealStart = currentPage - half
      
      // 開始位置を1〜(totalPages - visibleCount + 1)の範囲に収める
      const start = Math.max(1, Math.min(idealStart, totalPages - visibleCount + 1))
      
      return Array.from({ length: visibleCount }, (_, i) => start + i)
    }

    return {
      mobilePages: getVisiblePages(mobileMaxPages),
      desktopPages: getVisiblePages(desktopMaxPages),
    }
  }, [currentPage, totalPages, showPageNumbers, mobileMaxPages, desktopMaxPages])

  const buttonBase =
    'inline-flex items-center justify-center min-w-[36px] sm:min-w-[44px] h-9 sm:h-11 px-2 sm:px-3 text-sm font-medium rounded-md transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2'

  // ページ番号ボタンのレンダリングを共通化
  const renderPageNumbers = (pages: number[], keyPrefix: string) =>
    pages.map((page) => (
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
    ))

  return (
    <nav
      className={cn('flex items-center justify-between w-full', className)}
      aria-label="ページネーション"
    >
      {/* 左側: 最初へ + 前へ */}
      <div className="flex items-center gap-1">
        <button
          onClick={() => onPageChange(1)}
          disabled={currentPage === 1}
          className={cn(
            buttonBase,
            'border border-gray-300',
            currentPage === 1
              ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
              : 'bg-white text-gray-700 hover:bg-gray-50'
          )}
          aria-label="最初のページ"
        >
          <svg className="w-4 h-4 sm:w-5 sm:h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 19l-7-7 7-7M18 19l-7-7 7-7" />
          </svg>
        </button>
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
      </div>

      {/* ページ番号 */}
      {showPageNumbers && (
        <>
          {/* モバイル用 */}
          <div className="flex items-center gap-2 sm:hidden">
            {renderPageNumbers(mobilePages, 'mobile')}
          </div>
          {/* PC用 */}
          <div className="hidden sm:flex items-center gap-2">
            {renderPageNumbers(desktopPages, 'desktop')}
          </div>
        </>
      )}

      {/* 右側: 次へ + 最後へ */}
      <div className="flex items-center gap-1">
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
        <button
          onClick={() => onPageChange(totalPages)}
          disabled={currentPage === totalPages}
          className={cn(
            buttonBase,
            'border border-gray-300',
            currentPage === totalPages
              ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
              : 'bg-white text-gray-700 hover:bg-gray-50'
          )}
          aria-label="最後のページ"
        >
          <svg className="w-4 h-4 sm:w-5 sm:h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 5l7 7-7 7M6 5l7 7-7 7" />
          </svg>
        </button>
      </div>
    </nav>
  )
}

export default Pagination
