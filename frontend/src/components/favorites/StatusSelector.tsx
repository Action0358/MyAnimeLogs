'use client'

import React, { useState, useRef, useCallback } from 'react'
import { Plus, Bookmark, CirclePlay, CircleCheck, Check, X, ChevronDown } from 'lucide-react'
import { cn } from '@/lib/utils'
import { WatchStatusStatus } from '@/lib/api/generated/myanimelogs'
import { useOnClickOutside } from '@/hooks/useOnClickOutside'

// 視聴ステータスは 視聴予定 / 視聴中 / 視聴済 の3つ。
// 色は全ステータス共通のインディゴで統一し、違いはアイコンとラベルのみで示す
// （アニメ詳細UI刷新v2「ステータスボタン」）
export const STATUS_CONFIG = {
  planned: { label: '視聴予定', icon: Bookmark },
  watching: { label: '視聴中', icon: CirclePlay },
  completed: { label: '視聴済', icon: CircleCheck },
} as const

interface StatusSelectorProps {
  currentStatus: WatchStatusStatus | null
  onStatusChange: (status: WatchStatusStatus) => void
  onRemove?: () => void
  disabled?: boolean
  size?: 'sm' | 'md' | 'lg'
  /** outline: 枠線のみ（既定）。primary: 未選択=枠線／選択済み=インディゴ塗り
   *  （ページで一番重要な操作として扱う場所向け） */
  variant?: 'outline' | 'primary'
  /** 外側のラッパーに付与する。幅・高さを呼び出し側で揃えたいときに渡す */
  className?: string
}

export const StatusSelector: React.FC<StatusSelectorProps> = ({
  currentStatus,
  onStatusChange,
  onRemove,
  disabled = false,
  size = 'md',
  variant = 'outline',
  className,
}) => {
  const [isOpen, setIsOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)

  const closeDropdown = useCallback(() => setIsOpen(false), [])
  useOnClickOutside(dropdownRef, closeDropdown)

  const sizeStyles = {
    sm: 'text-xs px-2 py-1',
    md: 'text-sm px-3 py-2',
    lg: 'text-base px-4 py-2',
  }

  const statuses = Object.entries(STATUS_CONFIG) as [
    WatchStatusStatus,
    (typeof STATUS_CONFIG)[WatchStatusStatus],
  ][]
  const TriggerIcon = currentStatus ? STATUS_CONFIG[currentStatus].icon : Plus
  // primaryバリアントは未選択=枠線、選択済み=インディゴ塗りに切り替わる。
  // 記録済みであることが一目で分かるようにする
  const isFilled = variant === 'primary' && currentStatus !== null

  return (
    <div className={cn('relative', className)} ref={dropdownRef}>
      <button
        type="button"
        onClick={() => !disabled && setIsOpen(!isOpen)}
        disabled={disabled}
        className={cn(
          // ラッパーに幅・高さが指定された場合はそれに追従する
          'flex w-full h-full items-center justify-center gap-2 rounded-md font-semibold transition-colors',
          'focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2',
          'disabled:opacity-50 disabled:cursor-not-allowed',
          sizeStyles[size],
          isFilled
            ? 'border-none bg-primary text-primary-foreground hover:bg-primary/90'
            : 'border border-gray-300 bg-white text-gray-800 hover:bg-gray-50'
        )}
        aria-haspopup="listbox"
        aria-expanded={isOpen}
      >
        <TriggerIcon className="w-[17px] h-[17px]" aria-hidden="true" />
        <span className={currentStatus ? undefined : 'text-gray-500 font-normal'}>
          {currentStatus ? STATUS_CONFIG[currentStatus].label : 'ステータスを選択'}
        </span>
        {/* 開いている間は上向きにして、押すと閉じることを示す（検索のFilterPanelと同じ） */}
        <ChevronDown
          className={cn('w-4 h-4 ml-1 transition-transform duration-150', isOpen && 'rotate-180')}
          aria-hidden="true"
        />
      </button>

      {/* 常時マウントしてopacity/transformで開閉する（UserMenuと同じフェード＋6pxスライド） */}
      <div
        role="listbox"
        aria-hidden={!isOpen}
        className={cn(
          'absolute z-50 mt-1.5 w-full min-w-[140px] origin-top rounded-lg border border-border bg-popover p-1.5 shadow-lg',
          'transition-[opacity,transform] duration-150 ease-out',
          isOpen ? 'translate-y-0 opacity-100' : 'pointer-events-none -translate-y-1.5 opacity-0'
        )}
      >
        {statuses.map(([status, config]) => {
          const Icon = config.icon
          const isSelected = currentStatus === status
          return (
            <button
              key={status}
              type="button"
              onClick={() => {
                onStatusChange(status)
                setIsOpen(false)
              }}
              className={cn(
                'w-full flex items-center gap-2.5 h-10 px-2.5 rounded-md text-left text-sm transition-colors',
                'hover:bg-muted',
                isSelected
                  ? 'font-bold text-foreground bg-primary/5'
                  : 'font-medium text-foreground/90'
              )}
              role="option"
              aria-selected={isSelected}
            >
              <Icon
                className={cn(
                  'w-[17px] h-[17px]',
                  isSelected ? 'text-primary' : 'text-muted-foreground'
                )}
                aria-hidden="true"
              />
              <span className="flex-1">{config.label}</span>
              {isSelected && <Check className="w-4 h-4 text-primary" aria-hidden="true" />}
            </button>
          )
        })}
        {onRemove && currentStatus && (
          <>
            <div className="h-px bg-border mx-1 my-1.5" />
            <button
              type="button"
              onClick={() => {
                onRemove()
                setIsOpen(false)
              }}
              className="w-full flex items-center gap-2.5 h-10 px-2.5 rounded-md text-left text-sm font-medium text-muted-foreground hover:bg-muted transition-colors"
            >
              <X className="w-4 h-4" aria-hidden="true" />
              <span>ステータスを外す</span>
            </button>
          </>
        )}
      </div>
    </div>
  )
}
