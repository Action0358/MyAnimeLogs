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
  { value: 'title', label: 'タイトル順' },
  { value: 'year', label: '放送年順' },
  { value: 'popularity', label: '人気順' },
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
          'inline-flex items-center justify-between gap-2 px-4 py-2',
          'bg-white border border-gray-300 rounded-lg',
          'text-sm text-gray-700 hover:bg-gray-50',
          'focus:outline-none focus:ring-2 focus:ring-blue-500',
          'min-w-[140px]'
        )}
        aria-haspopup="listbox"
        aria-expanded={isOpen}
      >
        <span>{selectedOption?.label || '並び替え'}</span>
        <svg
          className={cn('w-4 h-4 transition-transform', isOpen && 'rotate-180')}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {isOpen && (
        <div
          className="absolute z-10 w-full mt-1 bg-white border border-gray-200 rounded-lg shadow-lg"
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
                    ? 'bg-blue-50 text-blue-700'
                    : 'text-gray-700 hover:bg-gray-50'
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
