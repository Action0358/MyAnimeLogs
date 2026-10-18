'use client'

import React, { useState, useRef, useEffect } from 'react'
import { cn } from '@/lib/utils'
import { RecentSearchesDropdown, useRecentSearches } from './RecentSearches'

export interface SearchFormProps {
  onSearch: (query: string) => void
  placeholder?: string
  showSuggestions?: boolean
  recentSearches?: string[]
  onRemoveHistory?: (query: string) => void
  onClearHistory?: () => void
  initialValue?: string
  className?: string
}

export const SearchForm: React.FC<SearchFormProps> = ({
  onSearch,
  placeholder = 'アニメを検索...',
  showSuggestions = true,
  recentSearches = [],
  onRemoveHistory,
  onClearHistory,
  initialValue = '',
  className,
}) => {
  const [query, setQuery] = useState(initialValue)
  const inputRef = useRef<HTMLInputElement>(null)

  const recent = useRecentSearches(
    recentSearches,
    (suggestion) => {
      setQuery(suggestion)
      onSearch(suggestion)
    },
    showSuggestions
  )

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (query.trim()) {
      onSearch(query.trim())
      recent.close()
    }
  }

  const handleClear = () => {
    setQuery('')
    inputRef.current?.focus()
  }

  return (
    <div ref={recent.containerRef} className={cn('relative', className)}>
      <form onSubmit={handleSubmit}>
        <div className="relative">
          <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none">
            <svg
              className="h-4 w-4 text-muted-foreground"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
              />
            </svg>
          </div>
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            {...recent.inputProps}
            placeholder={placeholder}
            className={cn(
              'block w-full h-10 pl-10 pr-10 border border-input rounded-full',
              'text-sm text-foreground placeholder-muted-foreground',
              'focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent',
              'transition-shadow'
            )}
            aria-label="検索"
          />
          {query && (
            <button
              type="button"
              onClick={handleClear}
              className="absolute inset-y-0 right-10 flex items-center pr-1.5 text-muted-foreground hover:text-foreground"
              aria-label="検索をクリア"
            >
              <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          )}
          <button
            type="submit"
            className="absolute inset-y-0 right-0 flex items-center pr-3.5 text-primary hover:text-primary/80"
            aria-label="検索を実行"
          >
            <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M14 5l7 7m0 0l-7 7m7-7H3"
              />
            </svg>
          </button>
        </div>
      </form>

      {showSuggestions && (
        <RecentSearchesDropdown
          items={recent.items}
          isOpen={recent.isOpen}
          selectedIndex={recent.selectedIndex}
          onSelect={recent.select}
          onRemove={onRemoveHistory}
          onClear={onClearHistory}
        />
      )}
    </div>
  )
}

export default SearchForm
