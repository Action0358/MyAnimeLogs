'use client'

import React from 'react'
import { cn } from '@/lib/utils'
import { Episode } from '@/lib/api/generated/myanimelogs'

interface EpisodeCheckboxProps {
  episode: Episode
  isWatched: boolean
  onWatchedChange: (episodeId: number, watched: boolean) => void
  disabled?: boolean
  showTitle?: boolean
}

export const EpisodeCheckbox: React.FC<EpisodeCheckboxProps> = ({
  episode,
  isWatched,
  onWatchedChange,
  disabled = false,
  showTitle = true,
}) => {
  const handleChange = () => {
    if (!disabled && episode.id) {
      onWatchedChange(episode.id, !isWatched)
    }
  }

  return (
    <label
      className={cn(
        'flex items-center gap-3 p-3 rounded-lg border transition-all cursor-pointer',
        'hover:bg-gray-50',
        isWatched
          ? 'border-green-200 bg-green-50'
          : 'border-gray-200 bg-white',
        disabled && 'opacity-50 cursor-not-allowed'
      )}
    >
      <div className="relative flex items-center justify-center">
        <input
          type="checkbox"
          checked={isWatched}
          onChange={handleChange}
          disabled={disabled}
          className="sr-only"
          aria-label={`エピソード ${episode.episode_number} を視聴済みにする`}
        />
        <div
          className={cn(
            'w-6 h-6 rounded-full border-2 flex items-center justify-center transition-all',
            isWatched
              ? 'border-green-500 bg-green-500'
              : 'border-gray-300 bg-white'
          )}
        >
          {isWatched && (
            <svg className="w-4 h-4 text-white" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
            </svg>
          )}
        </div>
      </div>
      
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className={cn(
            'inline-flex items-center justify-center w-8 h-8 rounded-full text-sm font-medium',
            isWatched
              ? 'bg-green-100 text-green-700'
              : 'bg-gray-100 text-gray-600'
          )}>
            {episode.episode_number}
          </span>
          {showTitle && episode.title && (
            <span className={cn(
              'text-sm truncate',
              isWatched ? 'text-gray-600' : 'text-gray-900'
            )}>
              {episode.title}
            </span>
          )}
        </div>
      </div>
    </label>
  )
}
