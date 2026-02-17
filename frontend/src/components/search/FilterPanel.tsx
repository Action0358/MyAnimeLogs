'use client'

import React from 'react'
import { cn } from '@/lib/utils'
import { SEASONS, GENRES } from '@/constants/search'

export interface FilterPanelProps {
  selectedSeason?: string
  selectedGenre?: string
  selectedStatus?: 'airing' | 'finished' | 'not_yet_aired'
  onSeasonChange: (season: string | undefined) => void
  onGenreChange: (genre: string | undefined) => void
  onStatusChange: (status: 'airing' | 'finished' | 'not_yet_aired' | undefined) => void
  onClear: () => void
  className?: string
}

export const FilterPanel: React.FC<FilterPanelProps> = ({
  selectedSeason,
  selectedGenre,
  selectedStatus,
  onSeasonChange,
  onGenreChange,
  onStatusChange,
  onClear,
  className,
}) => {
  const hasFilters = selectedSeason || selectedGenre || selectedStatus

  return (
    <div className={cn('bg-white border border-gray-200 rounded-lg p-4', className)}>
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-sm font-semibold text-gray-900">フィルター</h3>
        {hasFilters && (
          <button
            type="button"
            onClick={onClear}
            className="text-sm text-blue-600 hover:text-blue-700"
          >
            クリア
          </button>
        )}
      </div>

      <div className="space-y-4">
        {/* 放送時期 */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">放送時期</label>
          <select
            value={selectedSeason || ''}
            onChange={(e) => onSeasonChange(e.target.value || undefined)}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">すべて</option>
            {SEASONS.map((season) => (
              <option key={season} value={season}>
                {season}
              </option>
            ))}
          </select>
        </div>

        {/* ジャンル */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">ジャンル</label>
          <select
            value={selectedGenre || ''}
            onChange={(e) => onGenreChange(e.target.value || undefined)}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">すべて</option>
            {GENRES.map((genre) => (
              <option key={genre} value={genre}>
                {genre}
              </option>
            ))}
          </select>
        </div>

        {/* 放送状態 */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">放送状態</label>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => onStatusChange(undefined)}
              className={cn(
                'flex-1 px-3 py-2 text-sm rounded-lg border transition-colors',
                !selectedStatus
                  ? 'bg-blue-600 text-white border-blue-600'
                  : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'
              )}
            >
              すべて
            </button>
            <button
              type="button"
              onClick={() => onStatusChange('airing')}
              className={cn(
                'flex-1 px-3 py-2 text-sm rounded-lg border transition-colors',
                selectedStatus === 'airing'
                  ? 'bg-blue-600 text-white border-blue-600'
                  : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'
              )}
            >
              放送中
            </button>
            <button
              type="button"
              onClick={() => onStatusChange('finished')}
              className={cn(
                'flex-1 px-3 py-2 text-sm rounded-lg border transition-colors',
                selectedStatus === 'finished'
                  ? 'bg-blue-600 text-white border-blue-600'
                  : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'
              )}
            >
              完結
            </button>
            <button
              type="button"
              onClick={() => onStatusChange('not_yet_aired')}
              className={cn(
                'flex-1 px-3 py-2 text-sm rounded-lg border transition-colors',
                selectedStatus === 'not_yet_aired'
                  ? 'bg-blue-600 text-white border-blue-600'
                  : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'
              )}
            >
              放送予定
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

export default FilterPanel
