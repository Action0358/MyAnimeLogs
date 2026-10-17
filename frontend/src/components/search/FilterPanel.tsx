'use client'

import React, { useState } from 'react'
import { Check, ChevronDown } from 'lucide-react'
import { cn } from '@/lib/utils'
import { SEASONS, GENRES } from '@/constants/search'

export interface SortOption {
  value: string
  label: string
}

export const sortOptions: SortOption[] = [
  { value: 'year', label: '放送年順' },
  { value: 'popularity', label: '人気順' },
]

export interface FilterPanelProps {
  selectedSeason?: string
  /** 複数選択。未選択は空配列または undefined */
  selectedGenres?: string[]
  sort: string
  /** falseのとき放送時期の選択を出さない（今期・人気アニメ一覧向け。放送時期が固定/無関係なため） */
  showSeason?: boolean
  onSeasonChange?: (season: string | undefined) => void
  onGenresChange: (genres: string[]) => void
  onSortChange: (sort: string) => void
  className?: string
}

const toggleGenre = (genres: string[], genre: string): string[] =>
  genres.includes(genre) ? genres.filter((g) => g !== genre) : [...genres, genre]

interface GenreCheckboxListProps {
  selectedGenres: string[]
  onToggle: (genre: string) => void
}

const GenreCheckboxList: React.FC<GenreCheckboxListProps> = ({ selectedGenres, onToggle }) => (
  <div role="group" aria-label="ジャンル">
    {GENRES.map((genre) => {
      const isSelected = selectedGenres.includes(genre)
      return (
        <label
          key={genre}
          className="flex cursor-pointer items-center gap-2.5 px-3 py-2 text-sm text-foreground hover:bg-muted"
        >
          <input
            type="checkbox"
            checked={isSelected}
            onChange={() => onToggle(genre)}
            className="sr-only"
          />
          {/* ネイティブのcheckboxは色や大きさの調整が効かないため、見た目は自前で描く */}
          <span
            aria-hidden="true"
            className={cn(
              'flex h-4 w-4 flex-shrink-0 items-center justify-center rounded border',
              isSelected ? 'border-primary bg-primary text-primary-foreground' : 'border-input'
            )}
          >
            {isSelected && <Check className="h-3 w-3" strokeWidth={3} />}
          </span>
          {genre}
        </label>
      )
    })}
  </div>
)

interface GenreDropdownProps {
  selectedGenres: string[]
  onGenresChange: (genres: string[]) => void
}

/**
 * 放送時期の <select> と見た目を揃えるため、折りたたみ済みのボタンから
 * 開閉する形にする。
 *
 * ネイティブ <select multiple> にしないのは、複数選択に Ctrl/Cmd+クリックを
 * 要求しブラウザ間で見た目も揃わないため。絶対配置にしないのは、シート内で
 * スクロールしたときにボタン位置からズレて浮いて見えるのを避けるため
 * （通常のフローで下に伸ばす）。
 */
const GenreDropdown: React.FC<GenreDropdownProps> = ({ selectedGenres, onGenresChange }) => {
  const [isOpen, setIsOpen] = useState(false)

  return (
    <div>
      <button
        type="button"
        aria-haspopup="true"
        aria-expanded={isOpen}
        onClick={() => setIsOpen((prev) => !prev)}
        className="flex w-full items-center justify-between px-3 py-2 border border-input bg-background rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-ring"
      >
        <span className={selectedGenres.length > 0 ? 'text-foreground' : 'text-muted-foreground'}>
          {selectedGenres.length === 0 ? 'すべて' : selectedGenres.join('、')}
        </span>
        <ChevronDown className={cn('h-4 w-4 text-muted-foreground transition-transform', isOpen && 'rotate-180')} />
      </button>

      {isOpen && (
        <div className="mt-1.5 rounded-lg border border-input">
          <GenreCheckboxList
            selectedGenres={selectedGenres}
            onToggle={(genre) => onGenresChange(toggleGenre(selectedGenres, genre))}
          />
        </div>
      )}
    </div>
  )
}

export const FilterPanel: React.FC<FilterPanelProps> = ({
  selectedSeason,
  selectedGenres,
  sort,
  showSeason = true,
  onSeasonChange,
  onGenresChange,
  onSortChange,
  className,
}) => {
  const genres = selectedGenres ?? []

  return (
    <div className={cn('space-y-4', className)}>
      <div>
        <label htmlFor="filter-sort" className="block text-sm font-medium text-foreground mb-2">
          並び替え
        </label>
        <select
          id="filter-sort"
          value={sort}
          onChange={(e) => onSortChange(e.target.value)}
          className="w-full px-3 py-2 border border-input bg-background rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-ring"
        >
          {sortOptions.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
      </div>

      {showSeason && (
        <div>
          <label className="block text-sm font-medium text-foreground mb-2">放送時期</label>
          <select
            value={selectedSeason || ''}
            onChange={(e) => onSeasonChange?.(e.target.value || undefined)}
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
      )}

      <div>
        <div className="mb-2 flex items-center justify-between">
          <span className="text-sm font-medium text-foreground">ジャンル</span>
          {genres.length > 0 && (
            <button
              type="button"
              onClick={() => onGenresChange([])}
              className="text-xs text-muted-foreground hover:text-foreground"
            >
              すべて解除
            </button>
          )}
        </div>
        <GenreDropdown selectedGenres={genres} onGenresChange={onGenresChange} />
      </div>
    </div>
  )
}

export default FilterPanel
