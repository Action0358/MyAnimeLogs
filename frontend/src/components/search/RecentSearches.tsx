'use client'

import React, { useRef, useState } from 'react'
import { Clock, X } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useOnClickOutside } from '@/hooks/useOnClickOutside'

const LISTBOX_ID = 'search-suggestions'

/** ドロップダウンに出す履歴の件数。履歴自体は最大10件まで保存している */
export const MAX_VISIBLE_RECENT_SEARCHES = 5

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

  const items = recentSearches.slice(0, MAX_VISIBLE_RECENT_SEARCHES)
  const isOpen = isFocused && enabled && items.length > 0

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
        setSelectedIndex((prev) => (prev < items.length - 1 ? prev + 1 : prev))
        break
      case 'ArrowUp':
        e.preventDefault()
        setSelectedIndex((prev) => (prev > 0 ? prev - 1 : -1))
        break
      case 'Enter':
        if (selectedIndex >= 0 && selectedIndex < items.length) {
          e.preventDefault()
          select(items[selectedIndex])
        }
        break
      case 'Escape':
        close()
        break
    }
  }

  return {
    containerRef,
    /** ドロップダウンに出す履歴（新しい順に最大5件） */
    items,
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
  isOpen: boolean
  selectedIndex: number
  onSelect: (item: string) => void
  /** 指定すると各行に×を出し、1件ずつ削除できるようにする */
  onRemove?: (item: string) => void
  onClear?: () => void
  className?: string
}

/**
 * 「最近の検索」の候補リスト。検索欄から少し離して浮かせ、ユーザーメニューと同じ
 * 角丸・枠線・影・背景で表示する。
 * 開閉をフェード＋スライドで見せるため、履歴がある間は常時マウントし
 * opacity/transform で切り替える（UserMenu と同じパターン）。
 */
export const RecentSearchesDropdown: React.FC<RecentSearchesDropdownProps> = ({
  items,
  isOpen,
  selectedIndex,
  onSelect,
  onRemove,
  onClear,
  className,
}) => {
  if (items.length === 0) return null

  return (
    <div
      aria-hidden={!isOpen}
      className={cn(
        'absolute left-0 top-full z-10 mt-1.5 w-full rounded-lg border border-border bg-popover p-1.5 text-left text-popover-foreground shadow-lg',
        'transition-[opacity,transform] duration-150 ease-out',
        isOpen ? 'translate-y-0 opacity-100' : 'pointer-events-none -translate-y-1.5 opacity-0',
        className
      )}
    >
      <div className="flex items-center justify-between px-3 pb-1 pt-1.5">
        <span className="text-xs font-semibold text-muted-foreground">最近の検索</span>
        {onClear && (
          <button
            type="button"
            onClick={onClear}
            tabIndex={isOpen ? undefined : -1}
            className="text-xs text-muted-foreground transition-colors hover:text-foreground"
          >
            すべて削除
          </button>
        )}
      </div>
      <ul id={LISTBOX_ID} role="listbox" aria-label="最近の検索">
        {items.map((item, index) => (
          <li
            key={item}
            id={`suggestion-${index}`}
            role="option"
            aria-selected={index === selectedIndex}
            className={cn(
              'flex h-10 cursor-pointer items-center gap-2.5 rounded-md px-3 transition-colors max-sm:h-11',
              index === selectedIndex ? 'bg-muted' : 'hover:bg-muted'
            )}
            onClick={() => onSelect(item)}
          >
            <Clock className="h-4 w-4 flex-shrink-0 text-muted-foreground/70" aria-hidden="true" />
            <span className="min-w-0 flex-1 truncate text-sm text-foreground">{item}</span>
            {onRemove && (
              <button
                type="button"
                onClick={(e) => {
                  // 行のクリック（その語で検索）にしない
                  e.stopPropagation()
                  onRemove(item)
                }}
                tabIndex={isOpen ? undefined : -1}
                className="-mr-1.5 flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-md text-muted-foreground/70 transition-colors hover:bg-background hover:text-foreground max-sm:-mr-2 max-sm:h-9 max-sm:w-9"
                aria-label={`「${item}」を履歴から削除`}
              >
                <X className="h-4 w-4" aria-hidden="true" />
              </button>
            )}
          </li>
        ))}
      </ul>
    </div>
  )
}
