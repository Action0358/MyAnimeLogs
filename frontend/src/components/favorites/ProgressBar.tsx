'use client'

import React from 'react'
import { cn } from '@/lib/utils'

interface ProgressBarProps {
  current: number
  total: number
  showDetails?: boolean
  showPercentage?: boolean
  size?: 'sm' | 'md' | 'lg'
  animated?: boolean
  className?: string
}

export const ProgressBar: React.FC<ProgressBarProps> = ({
  current,
  total,
  showDetails = true,
  showPercentage = true,
  size = 'md',
  animated = true,
  className,
}) => {
  const percentage = total > 0 ? Math.round((current / total) * 100) : 0
  const isComplete = current >= total && total > 0

  const sizeStyles = {
    sm: 'h-1.5',
    md: 'h-2.5',
    lg: 'h-4',
  }

  const getProgressColor = () => {
    if (isComplete) return 'bg-green-500'
    if (percentage >= 75) return 'bg-blue-500'
    if (percentage >= 50) return 'bg-blue-400'
    if (percentage >= 25) return 'bg-blue-300'
    return 'bg-blue-200'
  }

  return (
    <div className={cn('w-full', className)}>
      {(showDetails || showPercentage) && (
        <div className="flex items-center justify-between mb-1 text-sm">
          {showDetails && (
            <span className="text-gray-600">
              {current} / {total} 話
            </span>
          )}
          {showPercentage && (
            <span className={cn('font-medium', isComplete ? 'text-green-600' : 'text-gray-700')}>
              {percentage}%
            </span>
          )}
        </div>
      )}

      <div className={cn('w-full bg-gray-200 rounded-full overflow-hidden', sizeStyles[size])}>
        <div
          className={cn(
            'h-full rounded-full transition-all duration-500',
            getProgressColor(),
            animated && 'transition-[width]'
          )}
          style={{ width: `${percentage}%` }}
          role="progressbar"
          aria-valuenow={current}
          aria-valuemin={0}
          aria-valuemax={total}
          aria-label={`視聴進捗: ${current}/${total}話 (${percentage}%)`}
        />
      </div>

      {isComplete && (
        <div className="flex items-center gap-1 mt-1 text-sm text-green-600">
          <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
            <path
              fillRule="evenodd"
              d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
              clipRule="evenodd"
            />
          </svg>
          <span>視聴完了</span>
        </div>
      )}
    </div>
  )
}
