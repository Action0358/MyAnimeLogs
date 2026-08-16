'use client'

import React, { useState, useRef } from 'react'
import { cn } from '@/lib/utils'
import { useOnClickOutside } from '@/hooks/useOnClickOutside'

export interface SortOption {
  value: string
  label: string
}

export interface SortDropdownProps {
  options: SortOption[]
  value?: string
  onChange: (value: string) => void
  className?: string
}

export const sortOptions: SortOption[] = [
  { value: 'year', label: '放送年順' },
  { value: 'popularity', label: '人気順' },
  { value: 'title', label: 'タイトル順' },
]

export const SortDropdown: React.FC<SortDropdownProps> = ({
  options = sortOptions,
  value,
  onChange,
  className,
}) => {
  const [isOpen, setIsOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  useOnClickOutside(containerRef, () => setIsOpen(false))

  const selectedOption = options.find((opt) => opt.value === value)

  return (
    <div ref={containerRef} className={cn('relative', className)}>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className={cn(
          'inline-flex w-full items-center justify-between gap-2 h-[38px] px-3.5',
          'bg-background border border-input rounded-lg',
          'text-sm font-medium text-foreground hover:bg-muted',
          'focus:outline-none focus:ring-2 focus:ring-ring',
          // 狭い画面では下限を持たせない。140px を固定下限にすると iPhone SE 幅で
          // 同じ行の件数が押し出され、レイアウトが崩れる
          'min-w-0 lg:min-w-[140px]'
        )}
        aria-haspopup="listbox"
        aria-expanded={isOpen}
      >
        <span className="truncate">{selectedOption?.label || '並び替え'}</span>
        <svg
          className={cn('w-4 h-4 flex-shrink-0 transition-transform', isOpen && 'rotate-180')}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {isOpen && (
        <div
          className="absolute z-10 w-full mt-1 bg-popover border border-border rounded-lg shadow-lg"
          role="listbox"
        >
          <ul className="py-1">
            {options.map((option) => (
              <li
                key={option.value}
                role="option"
                aria-selected={option.value === value}
                className={cn(
                  'px-4 py-2 text-sm cursor-pointer',
                  option.value === value
                    ? 'bg-primary/10 text-primary font-semibold'
                    : 'text-popover-foreground hover:bg-muted'
                )}
                onClick={() => {
                  onChange(option.value)
                  setIsOpen(false)
                }}
              >
                {option.label}
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  )
}

export default SortDropdown
