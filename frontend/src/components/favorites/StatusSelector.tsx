'use client'

import React, { useState, useRef, useCallback } from 'react'
import { cn } from '@/lib/utils'
import { WatchStatusStatus } from '@/lib/api/generated/myanimelogs'
import { useOnClickOutside } from '@/hooks/useOnClickOutside'

// 視聴ステータスは 視聴予定 / 視聴中 / 視聴済 の3つ
export const STATUS_CONFIG = {
  planned: { label: '視聴予定', color: 'bg-yellow-500', textColor: 'text-yellow-600', icon: '⏰' },
  watching: { label: '視聴中', color: 'bg-blue-500', textColor: 'text-blue-600', icon: '▶' },
  completed: { label: '視聴済', color: 'bg-green-500', textColor: 'text-green-600', icon: '✓' },
} as const

interface StatusSelectorProps {
  currentStatus: WatchStatusStatus | null
  onStatusChange: (status: WatchStatusStatus) => void
  onRemove?: () => void
  disabled?: boolean
  size?: 'sm' | 'md' | 'lg'
  /** 外側のラッパーに付与する。幅・高さを呼び出し側で揃えたいときに渡す */
  className?: string
}

export const StatusSelector: React.FC<StatusSelectorProps> = ({
  currentStatus,
  onStatusChange,
  onRemove,
  disabled = false,
  size = 'md',
  className,
}) => {
  const [isOpen, setIsOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)

  const closeDropdown = useCallback(() => setIsOpen(false), [])
  useOnClickOutside(dropdownRef, closeDropdown)

  const sizeStyles = {
    sm: 'text-xs px-2 py-1',
    md: 'text-sm px-3 py-2',
    lg: 'text-base px-4 py-2',
  }

  const statuses = Object.entries(STATUS_CONFIG) as [WatchStatusStatus, typeof STATUS_CONFIG[WatchStatusStatus]][]

  return (
    <div className={cn('relative', className)} ref={dropdownRef}>
      <button
        type="button"
        onClick={() => !disabled && setIsOpen(!isOpen)}
        disabled={disabled}
        className={cn(
          // ラッパーに幅・高さが指定された場合はそれに追従する
          'flex w-full h-full items-center justify-center gap-2 rounded-md border border-gray-300 bg-white transition-colors',
          'hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500',
          'disabled:opacity-50 disabled:cursor-not-allowed',
          sizeStyles[size]
        )}
        aria-haspopup="listbox"
        aria-expanded={isOpen}
      >
        {currentStatus ? (
          <>
            <span className={cn('w-2 h-2 rounded-full', STATUS_CONFIG[currentStatus].color)} />
            <span>{STATUS_CONFIG[currentStatus].label}</span>
          </>
        ) : (
          <span className="text-gray-500">ステータスを選択</span>
        )}
        <svg className="w-4 h-4 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {isOpen && (
        <div className="absolute z-50 mt-1 w-full min-w-[140px] bg-white border border-gray-200 rounded-md shadow-lg">
          <ul role="listbox" className="py-1">
            {statuses.map(([status, config]) => (
              <li key={status}>
                <button
                  type="button"
                  onClick={() => {
                    onStatusChange(status)
                    setIsOpen(false)
                  }}
                  className={cn(
                    'w-full flex items-center gap-2 px-3 py-2 text-left text-sm',
                    'hover:bg-gray-100 transition-colors',
                    currentStatus === status && 'bg-gray-50'
                  )}
                  role="option"
                  aria-selected={currentStatus === status}
                >
                  <span className={cn('w-2 h-2 rounded-full', config.color)} />
                  <span>{config.label}</span>
                  {currentStatus === status && (
                    <svg className="w-4 h-4 ml-auto text-blue-600" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                    </svg>
                  )}
                </button>
              </li>
            ))}
            {onRemove && currentStatus && (
              <>
                <li className="border-t border-gray-200 my-1" />
                <li>
                  <button
                    type="button"
                    onClick={() => {
                      onRemove()
                      setIsOpen(false)
                    }}
                    className="w-full flex items-center gap-2 px-3 py-2 text-left text-sm text-red-600 hover:bg-red-50 transition-colors"
                  >
                    <span>✕</span>
                    <span>解除</span>
                  </button>
                </li>
              </>
            )}
          </ul>
        </div>
      )}
    </div>
  )
}
