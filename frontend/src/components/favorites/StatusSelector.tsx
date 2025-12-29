'use client'

import React, { useState, useRef, useEffect } from 'react'
import { cn } from '@/lib/utils'
import { FavoriteStatus } from '@/lib/api/generated/myanimelogs'

export const STATUS_CONFIG = {
  watching: { label: '視聴中', color: 'bg-blue-500', textColor: 'text-blue-600', icon: '▶' },
  completed: { label: '完了', color: 'bg-green-500', textColor: 'text-green-600', icon: '✓' },
  planned: { label: '視聴予定', color: 'bg-yellow-500', textColor: 'text-yellow-600', icon: '⏰' },
  dropped: { label: '中断', color: 'bg-red-500', textColor: 'text-red-600', icon: '✕' },
} as const

interface StatusSelectorProps {
  currentStatus: FavoriteStatus | null
  onStatusChange: (status: FavoriteStatus) => void
  onRemove?: () => void
  disabled?: boolean
  size?: 'sm' | 'md' | 'lg'
}

export const StatusSelector: React.FC<StatusSelectorProps> = ({
  currentStatus,
  onStatusChange,
  onRemove,
  disabled = false,
  size = 'md',
}) => {
  const [isOpen, setIsOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const sizeStyles = {
    sm: 'text-xs px-2 py-1',
    md: 'text-sm px-3 py-2',
    lg: 'text-base px-4 py-2',
  }

  const statuses = Object.entries(STATUS_CONFIG) as [FavoriteStatus, typeof STATUS_CONFIG[FavoriteStatus]][]

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        type="button"
        onClick={() => !disabled && setIsOpen(!isOpen)}
        disabled={disabled}
        className={cn(
          'flex items-center gap-2 rounded-md border border-gray-300 bg-white transition-colors',
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
                    <span>お気に入りから削除</span>
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
