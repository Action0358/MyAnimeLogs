'use client'

import React, { useState, useRef, useEffect } from 'react'
import { cn } from '@/lib/utils'
import { FavoriteStatus } from '@/lib/api/generated/myanimelogs'
import { STATUS_CONFIG } from './StatusSelector'

interface FavoriteButtonProps {
  animeId: number
  currentStatus: FavoriteStatus | null
  onStatusChange: (status: FavoriteStatus) => void
  onRemove: () => void
  size?: 'sm' | 'md' | 'lg'
  variant?: 'button' | 'icon'
  isLoading?: boolean
}

export const FavoriteButton: React.FC<FavoriteButtonProps> = ({
  animeId,
  currentStatus,
  onStatusChange,
  onRemove,
  size = 'md',
  variant = 'button',
  isLoading = false,
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
    sm: { button: 'text-xs px-2 py-1', icon: 'w-8 h-8' },
    md: { button: 'text-sm px-3 py-2', icon: 'w-10 h-10' },
    lg: { button: 'text-base px-4 py-2', icon: 'w-12 h-12' },
  }

  const isFavorite = currentStatus !== null
  const statuses = Object.entries(STATUS_CONFIG) as [FavoriteStatus, typeof STATUS_CONFIG[FavoriteStatus]][]

  if (variant === 'icon') {
    return (
      <div className="relative" ref={dropdownRef}>
        <button
          type="button"
          onClick={() => setIsOpen(!isOpen)}
          disabled={isLoading}
          className={cn(
            'flex items-center justify-center rounded-full transition-all',
            'focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2',
            'disabled:opacity-50 disabled:cursor-not-allowed',
            sizeStyles[size].icon,
            isFavorite
              ? 'bg-pink-500 text-white hover:bg-pink-600'
              : 'bg-gray-200 text-gray-600 hover:bg-gray-300'
          )}
          aria-label={isFavorite ? 'お気に入り設定を変更' : 'お気に入りに追加'}
        >
          {isLoading ? (
            <svg className="animate-spin w-5 h-5" fill="none" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
            </svg>
          ) : (
            <svg className="w-5 h-5" fill={isFavorite ? 'currentColor' : 'none'} stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
            </svg>
          )}
        </button>

        {isOpen && (
          <div className="absolute right-0 z-50 mt-2 w-40 bg-white border border-gray-200 rounded-md shadow-lg">
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
                  >
                    <span className={cn('w-2 h-2 rounded-full', config.color)} />
                    <span>{config.label}</span>
                  </button>
                </li>
              ))}
              {isFavorite && (
                <>
                  <li className="border-t border-gray-200 my-1" />
                  <li>
                    <button
                      type="button"
                      onClick={() => {
                        onRemove()
                        setIsOpen(false)
                      }}
                      className="w-full flex items-center gap-2 px-3 py-2 text-left text-sm text-red-600 hover:bg-red-50"
                    >
                      <span>✕</span>
                      <span>削除</span>
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

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        disabled={isLoading}
        className={cn(
          'flex items-center gap-2 rounded-md border transition-colors',
          'focus:outline-none focus:ring-2 focus:ring-blue-500',
          'disabled:opacity-50 disabled:cursor-not-allowed',
          sizeStyles[size].button,
          isFavorite
            ? 'border-pink-500 bg-pink-50 text-pink-700 hover:bg-pink-100'
            : 'border-gray-300 bg-white text-gray-700 hover:bg-gray-50'
        )}
      >
        {isLoading ? (
          <svg className="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
        ) : (
          <svg className="w-4 h-4" fill={isFavorite ? 'currentColor' : 'none'} stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
          </svg>
        )}
        {isFavorite && currentStatus ? (
          <>
            <span className={cn('w-2 h-2 rounded-full', STATUS_CONFIG[currentStatus].color)} />
            <span>{STATUS_CONFIG[currentStatus].label}</span>
          </>
        ) : (
          <span>お気に入り</span>
        )}
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {isOpen && (
        <div className="absolute left-0 z-50 mt-2 w-40 bg-white border border-gray-200 rounded-md shadow-lg">
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
                >
                  <span className={cn('w-2 h-2 rounded-full', config.color)} />
                  <span>{config.label}</span>
                </button>
              </li>
            ))}
            {isFavorite && (
              <>
                <li className="border-t border-gray-200 my-1" />
                <li>
                  <button
                    type="button"
                    onClick={() => {
                      onRemove()
                      setIsOpen(false)
                    }}
                    className="w-full flex items-center gap-2 px-3 py-2 text-left text-sm text-red-600 hover:bg-red-50"
                  >
                    <span>✕</span>
                    <span>削除</span>
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
