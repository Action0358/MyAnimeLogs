'use client'

import React, { useRef, useState } from 'react'
import { cn } from '@/lib/utils'
import { useOnClickOutside } from '@/hooks/useOnClickOutside'

const LISTBOX_ID = 'search-suggestions'

/**
 * 検索欄の「最近の検索」ドロップダウンの開閉とキーボード操作。
 * 作品検索画面（SearchForm）とトップページの検索欄で同じ挙動にするため共通化している。
 * 入力欄にフォーカスすると履歴を開き、↑↓で選択・Enterで確定・Escで閉じる。
 */
export const useRecentSearches = (
  recentSearches: string[],
  onSelect: (item: string) => void,
  enabled = true
) => {
  const [isFocused, setIsFocused] = useState(false)
  const [selectedIndex, setSelectedIndex] = useState(-1)
  const containerRef = useRef<HTMLDivElement>(null)

  useOnClickOutside(containerRef, () => setIsFocused(false))

  const isOpen = isFocused && enabled && recentSearches.length > 0

  const close = () => {
    setIsFocused(false)
    setSelectedIndex(-1)
  }

  const select = (item: string) => {
    onSelect(item)
    close()
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (!isOpen) return

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault()
        setSelectedIndex((prev) => (prev < recentSearches.length - 1 ? prev + 1 : prev))
        break
      case 'ArrowUp':
        e.preventDefault()
        setSelectedIndex((prev) => (prev > 0 ? prev - 1 : -1))
        break
      case 'Enter':
        if (selectedIndex >= 0 && selectedIndex < recentSearches.length) {
          e.preventDefault()
          select(recentSearches[selectedIndex])
        }
        break
      case 'Escape':
        close()
        break
    }
  }

  return {
    containerRef,
    isOpen,
    selectedIndex,
    close,
    select,
    /** 入力欄に渡す属性 */
    inputProps: {
      onFocus: () => setIsFocused(true),
      onKeyDown: handleKeyDown,
      'aria-autocomplete': 'list' as const,
      'aria-controls': isOpen ? LISTBOX_ID : undefined,
      'aria-activedescendant': selectedIndex >= 0 ? `suggestion-${selectedIndex}` : undefined,
    },
  }
}

export interface RecentSearchesDropdownProps {
  items: string[]
  selectedIndex: number
  onSelect: (item: string) => void
  onClear?: () => void
  className?: string
}

/** 「最近の検索」の候補リスト。入力欄の直下に重ねて表示する */
export const RecentSearchesDropdown: React.FC<RecentSearchesDropdownProps> = ({
  items,
  selectedIndex,
  onSelect,
  onClear,
  className,
}) => (
  <div
    id={LISTBOX_ID}
    className={cn(
      'absolute z-10 w-full mt-1 bg-popover border border-border rounded-lg shadow-lg',
      className
    )}
    role="listbox"
  >
    <div className="flex items-center justify-between px-3 py-2 border-b border-border">
      <span className="text-xs text-muted-foreground font-medium">最近の検索</span>
      {onClear && (
        <button
          type="button"
          onClick={onClear}
          className="text-xs text-primary hover:text-primary/80"
        >
          履歴をクリア
        </button>
      )}
    </div>
    <ul className="py-1">
      {items.map((item, index) => (
        <li
          key={item}
          id={`suggestion-${index}`}
          role="option"
          aria-selected={index === selectedIndex}
          className={cn(
            'px-3 py-2 cursor-pointer flex items-center gap-2',
            index === selectedIndex ? 'bg-primary/10' : 'hover:bg-muted'
          )}
          onClick={() => onSelect(item)}
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
)
