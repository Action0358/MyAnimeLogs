'use client'

import React, { useRef, useState } from 'react'
import { Check, ChevronDown } from 'lucide-react'
import { cn } from '@/lib/utils'
import { SEASONS, GENRES } from '@/constants/search'
import { useOnClickOutside } from '@/hooks/useOnClickOutside'

export interface FilterPanelProps {
  selectedSeason?: string
  /** 複数選択。未選択は空配列または undefined */
  selectedGenres?: string[]
  onSeasonChange: (season: string | undefined) => void
  onGenresChange: (genres: string[]) => void
  /** 'inline': デスクトップツールバー用の横並びチップ / 'sheet': モバイルシート用の縦並び+ラベル付き */
  variant?: 'inline' | 'sheet'
  className?: string
}

interface ChipSelectProps {
  ariaLabel: string
  value: string
  options: { value: string; label: string }[]
  onChange: (value: string) => void
}

/**
 * 選択中のジャンルを1行のラベルにする。
 *
 * 他のチップ（放送時期・放送状態）と同じく、未選択のときだけ「◯◯：すべて」と
 * 項目名を出し、選択中は値だけを見せる。項目名を付けたままだとチップが伸びて
 * ツールバーが折り返す。複数選択時は件数表記にして幅を一定に保つ。
 */
const genreSummaryLabel = (genres: string[]): string => {
  if (genres.length === 0) return 'ジャンル：すべて'
  if (genres.length === 1) return genres[0]
  return `ジャンル${genres.length}件`
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

interface GenreMultiSelectProps {
  selectedGenres: string[]
  onGenresChange: (genres: string[]) => void
}

/**
 * デスクトップのツールバー用。他のチップと高さを揃えたボタンで
 * チェックボックス一覧を開く。
 *
 * <select multiple> を使わないのは、ブラウザ標準の複数選択が
 * Ctrl/Cmd+クリックを要求し、チップ列の中で高さも揃わないため。
 */
const GenreMultiSelect: React.FC<GenreMultiSelectProps> = ({ selectedGenres, onGenresChange }) => {
  const [isOpen, setIsOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)
  useOnClickOutside(containerRef, () => setIsOpen(false))

  return (
    <div ref={containerRef} className="relative">
      <button
        type="button"
        aria-haspopup="true"
        aria-expanded={isOpen}
        // 表示ラベルは幅の都合で「アクション」「ジャンル3件」と短くするため、
        // 読み上げには何の絞り込みで何を選んでいるかを明示する
        aria-label={
          selectedGenres.length === 0
            ? 'ジャンル：すべて'
            : `ジャンル：${selectedGenres.join('、')}`
        }
        onClick={() => setIsOpen((prev) => !prev)}
        className={cn(
          'inline-flex h-[38px] items-center gap-2 rounded-lg border border-input bg-background',
          'pl-3.5 pr-2.5 text-sm font-medium text-foreground',
          'focus:outline-none focus:ring-2 focus:ring-ring',
          selectedGenres.length > 0 && 'border-primary text-primary'
        )}
      >
        {genreSummaryLabel(selectedGenres)}
        <ChevronDown className="h-3.5 w-3.5 text-muted-foreground" />
      </button>

      {isOpen && (
        <div
          className="absolute left-0 top-[42px] z-30 w-52 rounded-lg border border-border bg-background py-1.5 shadow-lg"
          // Escで閉じられないと、開いたまま操作不能に見えてしまう
          onKeyDown={(e) => {
            if (e.key === 'Escape') setIsOpen(false)
          }}
        >
          <GenreCheckboxList
            selectedGenres={selectedGenres}
            onToggle={(genre) => onGenresChange(toggleGenre(selectedGenres, genre))}
          />
          {selectedGenres.length > 0 && (
            <button
              type="button"
              onClick={() => onGenresChange([])}
              className="mt-1 w-full border-t border-border px-3 pb-1 pt-2 text-left text-xs text-muted-foreground hover:text-foreground"
            >
              選択をすべて解除
            </button>
          )}
        </div>
      )}
    </div>
  )
}

interface GenreDropdownProps {
  selectedGenres: string[]
  onGenresChange: (genres: string[]) => void
}

/**
 * モバイルシート用。放送時期の <select> と見た目を揃えるため、
 * 折りたたみ済みのボタンから開閉する形にする。
 *
 * ネイティブ <select multiple> ではなく開閉パネルにするのは、
 * デスクトップの GenreMultiSelect と同じ理由（複数選択の操作性）。
 * 絶対配置にしないのは、シート内でスクロールしたときにボタン位置から
 * ズレて浮いて見えるのを避けるため（通常のフローで下に伸ばす）。
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
  selectedGenres,
  onSeasonChange,
  onGenresChange,
  variant = 'inline',
  className,
}) => {
  const genres = selectedGenres ?? []

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
      <GenreMultiSelect selectedGenres={genres} onGenresChange={onGenresChange} />
    </div>
  )
}

export default FilterPanel
