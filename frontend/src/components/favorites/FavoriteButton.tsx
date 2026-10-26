'use client'

import React from 'react'
import { cn } from '@/lib/utils'

const LoadingSpinner: React.FC<{ className?: string }> = ({ className = 'w-5 h-5' }) => (
  <svg className={cn('animate-spin', className)} fill="none" viewBox="0 0 24 24">
    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
    <path
      className="opacity-75"
      fill="currentColor"
      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
    />
  </svg>
)

const HeartIcon: React.FC<{ filled: boolean; className?: string }> = ({
  filled,
  className = 'w-5 h-5',
}) => (
  <svg
    className={className}
    fill={filled ? 'currentColor' : 'none'}
    stroke="currentColor"
    viewBox="0 0 24 24"
  >
    <path
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={2}
      d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"
    />
  </svg>
)

interface FavoriteButtonProps {
  isFavorite: boolean
  onToggle: () => void
  size?: 'sm' | 'md' | 'lg'
  variant?: 'button' | 'icon'
  isLoading?: boolean
  /** 幅・高さを呼び出し側で揃えたいときに渡す */
  className?: string
}

// お気に入り（ハート）のON/OFFを切り替えるボタン。視聴ステータスとは独立。
export const FavoriteButton: React.FC<FavoriteButtonProps> = ({
  isFavorite,
  onToggle,
  size = 'md',
  variant = 'button',
  isLoading = false,
  className,
}) => {
  const sizeStyles = {
    sm: { button: 'text-xs px-2 py-1', icon: 'w-8 h-8' },
    md: { button: 'text-sm px-3 py-2', icon: 'w-10 h-10' },
    lg: { button: 'text-base px-4 py-2', icon: 'w-12 h-12' },
  }

  if (variant === 'icon') {
    return (
      <button
        type="button"
        onClick={onToggle}
        disabled={isLoading}
        className={cn(
          'flex items-center justify-center rounded-md border transition-colors',
          'focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2',
          'disabled:opacity-50 disabled:cursor-not-allowed',
          sizeStyles[size].icon,
          isFavorite
            ? 'border-pink-200 bg-pink-50 text-pink-600 hover:bg-pink-100'
            : 'border-gray-300 bg-white text-gray-600 hover:bg-gray-50'
        )}
        aria-label={isFavorite ? 'お気に入りから削除' : 'お気に入りに追加'}
        aria-pressed={isFavorite}
      >
        {isLoading ? <LoadingSpinner /> : <HeartIcon filled={isFavorite} />}
      </button>
    )
  }

  return (
    <button
      type="button"
      onClick={onToggle}
      disabled={isLoading}
      className={cn(
        'flex items-center gap-2 rounded-md border transition-colors',
        'focus:outline-none focus:ring-2 focus:ring-blue-500',
        'disabled:opacity-50 disabled:cursor-not-allowed',
        sizeStyles[size].button,
        isFavorite
          ? 'border-pink-500 bg-pink-50 text-pink-700 hover:bg-pink-100'
          : 'border-gray-300 bg-white text-gray-700 hover:bg-gray-50',
        className
      )}
      aria-pressed={isFavorite}
    >
      {isLoading ? (
        <LoadingSpinner className="w-4 h-4" />
      ) : (
        <HeartIcon filled={isFavorite} className="w-4 h-4" />
      )}
      <span>{isFavorite ? 'お気に入り済み' : 'お気に入り'}</span>
    </button>
  )
}
