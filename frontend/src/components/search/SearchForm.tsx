'use client'

import React, { useState, useRef, useEffect } from 'react'
import { cn } from '@/lib/utils'
import { useOnClickOutside } from '@/hooks/useOnClickOutside'

export interface SearchFormProps {
  onSearch: (query: string) => void
  placeholder?: string
  showSuggestions?: boolean
  recentSearches?: string[]
  onClearHistory?: () => void
  initialValue?: string
  className?: string
}

export const SearchForm: React.FC<SearchFormProps> = ({
  onSearch,
  placeholder = 'アニメを検索...',
  showSuggestions = true,
  recentSearches = [],
  onClearHistory,
  initialValue = '',
  className,
}) => {
  const [query, setQuery] = useState(initialValue)
  const [isFocused, setIsFocused] = useState(false)
  const [selectedIndex, setSelectedIndex] = useState(-1)
  const containerRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  useOnClickOutside(containerRef, () => setIsFocused(false))

  useEffect(() => {
    setQuery(initialValue)
  }, [initialValue])

  const showDropdown = isFocused && showSuggestions && recentSearches.length > 0

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (query.trim()) {
      onSearch(query.trim())
      setIsFocused(false)
    }
  }

  const handleSelectSuggestion = (suggestion: string) => {
    setQuery(suggestion)
    onSearch(suggestion)
    setIsFocused(false)
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (!showDropdown) return

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault()
        setSelectedIndex((prev) =>
          prev < recentSearches.length - 1 ? prev + 1 : prev
        )
        break
      case 'ArrowUp':
        e.preventDefault()
        setSelectedIndex((prev) => (prev > 0 ? prev - 1 : -1))
        break
      case 'Enter':
        if (selectedIndex >= 0 && selectedIndex < recentSearches.length) {
          e.preventDefault()
          handleSelectSuggestion(recentSearches[selectedIndex])
        }
        break
      case 'Escape':
        setIsFocused(false)
        setSelectedIndex(-1)
        break
    }
  }

  const handleClear = () => {
    setQuery('')
    inputRef.current?.focus()
  }

  return (
    <div ref={containerRef} className={cn('relative', className)}>
      <form onSubmit={handleSubmit}>
        <div className="relative">
          <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
            <svg
              className="h-5 w-5 text-muted-foreground"
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
            onFocus={() => setIsFocused(true)}
            onKeyDown={handleKeyDown}
            placeholder={placeholder}
            className={cn(
              'block w-full h-[52px] pl-11 pr-12 border border-input rounded-lg',
              'text-foreground placeholder-muted-foreground',
              'focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent',
              'transition-shadow'
            )}
            aria-label="検索"
            aria-autocomplete="list"
            aria-controls={showDropdown ? 'search-suggestions' : undefined}
            aria-activedescendant={
              selectedIndex >= 0 ? `suggestion-${selectedIndex}` : undefined
            }
          />
          {query && (
            <button
              type="button"
              onClick={handleClear}
              className="absolute inset-y-0 right-12 flex items-center pr-2 text-muted-foreground hover:text-foreground"
              aria-label="検索をクリア"
            >
              <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
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
            className="absolute inset-y-0 right-0 flex items-center pr-4 text-primary hover:text-primary/80"
            aria-label="検索を実行"
          >
            <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
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

      {showDropdown && (
        <div
          id="search-suggestions"
          className="absolute z-10 w-full mt-1 bg-popover border border-border rounded-lg shadow-lg"
          role="listbox"
        >
          <div className="flex items-center justify-between px-3 py-2 border-b border-border">
            <span className="text-xs text-muted-foreground font-medium">最近の検索</span>
            {onClearHistory && (
              <button
                type="button"
                onClick={onClearHistory}
                className="text-xs text-primary hover:text-primary/80"
              >
                履歴をクリア
              </button>
            )}
          </div>
          <ul className="py-1">
            {recentSearches.map((item, index) => (
              <li
                key={item}
                id={`suggestion-${index}`}
                role="option"
                aria-selected={index === selectedIndex}
                className={cn(
                  'px-3 py-2 cursor-pointer flex items-center gap-2',
                  index === selectedIndex ? 'bg-primary/10' : 'hover:bg-muted'
                )}
                onClick={() => handleSelectSuggestion(item)}
              >
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
                    d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
                <span className="text-sm text-foreground">{item}</span>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  )
}

export default SearchForm
