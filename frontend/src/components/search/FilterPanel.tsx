'use client'

import React from 'react'
import { ChevronDown } from 'lucide-react'
import { cn } from '@/lib/utils'
import { SEASONS, GENRES } from '@/constants/search'

export interface FilterPanelProps {
  selectedSeason?: string
  selectedGenre?: string
  selectedStatus?: 'airing' | 'finished' | 'not_yet_aired'
  onSeasonChange: (season: string | undefined) => void
  onGenreChange: (genre: string | undefined) => void
  onStatusChange: (status: 'airing' | 'finished' | 'not_yet_aired' | undefined) => void
  /** 'inline': デスクトップツールバー用の横並びチップ / 'sheet': モバイルシート用の縦並び+ラベル付き */
  variant?: 'inline' | 'sheet'
  className?: string
}

const STATUS_VALUES = ['airing', 'finished', 'not_yet_aired'] as const

const STATUS_LABELS: Record<(typeof STATUS_VALUES)[number], string> = {
  airing: '放送中',
  finished: '完結',
  not_yet_aired: '放送予定',
}

interface ChipSelectProps {
  ariaLabel: string
  value: string
  options: { value: string; label: string }[]
  onChange: (value: string) => void
}

const ChipSelect: React.FC<ChipSelectProps> = ({ ariaLabel, value, options, onChange }) => (
  <div className="relative inline-flex items-center">
    <select
      aria-label={ariaLabel}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className={cn(
        'h-[38px] pl-3.5 pr-8 rounded-lg border border-input bg-background',
        'text-sm font-medium text-foreground appearance-none cursor-pointer',
        'focus:outline-none focus:ring-2 focus:ring-ring'
      )}
    >
      {options.map((opt) => (
        <option key={opt.value} value={opt.value}>
          {opt.label}
        </option>
      ))}
    </select>
    <ChevronDown className="w-3.5 h-3.5 text-muted-foreground absolute right-2.5 pointer-events-none" />
  </div>
)

export const FilterPanel: React.FC<FilterPanelProps> = ({
  selectedSeason,
  selectedGenre,
  selectedStatus,
  onSeasonChange,
  onGenreChange,
  onStatusChange,
  variant = 'inline',
  className,
}) => {
  if (variant === 'sheet') {
    return (
      <div className={cn('space-y-4', className)}>
        <h3 className="text-sm font-semibold text-foreground">フィルター</h3>

        <div>
          <label className="block text-sm font-medium text-foreground mb-2">放送時期</label>
          <select
            value={selectedSeason || ''}
            onChange={(e) => onSeasonChange(e.target.value || undefined)}
            className="w-full px-3 py-2 border border-input bg-background rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-ring"
          >
            <option value="">すべて</option>
            {SEASONS.map((season) => (
              <option key={season} value={season}>
                {season}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-foreground mb-2">ジャンル</label>
          <select
            value={selectedGenre || ''}
            onChange={(e) => onGenreChange(e.target.value || undefined)}
            className="w-full px-3 py-2 border border-input bg-background rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-ring"
          >
            <option value="">すべて</option>
            {GENRES.map((genre) => (
              <option key={genre} value={genre}>
                {genre}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-foreground mb-2">放送状態</label>
          <select
            value={selectedStatus || ''}
            onChange={(e) =>
              onStatusChange(
                (e.target.value || undefined) as 'airing' | 'finished' | 'not_yet_aired' | undefined
              )
            }
            className="w-full px-3 py-2 border border-input bg-background rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-ring"
          >
            <option value="">すべて</option>
            {STATUS_VALUES.map((status) => (
              <option key={status} value={status}>
                {STATUS_LABELS[status]}
              </option>
            ))}
          </select>
        </div>
      </div>
    )
  }

  return (
    <div className={cn('flex items-center gap-2.5 flex-wrap', className)}>
      <ChipSelect
        ariaLabel="放送時期"
        value={selectedSeason || ''}
        options={[
          { value: '', label: '放送時期：すべて' },
          ...SEASONS.map((season) => ({ value: season, label: season })),
        ]}
        onChange={(v) => onSeasonChange(v || undefined)}
      />
      <ChipSelect
        ariaLabel="ジャンル"
        value={selectedGenre || ''}
        options={[
          { value: '', label: 'ジャンル：すべて' },
          ...GENRES.map((genre) => ({ value: genre, label: genre })),
        ]}
        onChange={(v) => onGenreChange(v || undefined)}
      />
      <ChipSelect
        ariaLabel="放送状態"
        value={selectedStatus || ''}
        options={[
          { value: '', label: '放送状態：すべて' },
          ...STATUS_VALUES.map((status) => ({ value: status, label: STATUS_LABELS[status] })),
        ]}
        onChange={(v) =>
          onStatusChange((v || undefined) as 'airing' | 'finished' | 'not_yet_aired' | undefined)
        }
      />
    </div>
  )
}

export default FilterPanel
