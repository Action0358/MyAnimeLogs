'use client'

import React, { useRef } from 'react'
import { X } from 'lucide-react'
import { useOnClickOutside } from '@/hooks/useOnClickOutside'
import { FilterPanel, type FilterPanelProps } from './FilterPanel'

/**
 * フィルターパネル（画面幅を問わず同じ「ボタン→開閉パネル」の体験に揃える）。
 *
 * モバイルは下からのスライドシート、デスクトップ（lg以上）は画面中央の
 * ダイアログとして表示する。検索・今期アニメ・人気アニメの3ページで共有する。
 */
export function FilterSheet({
  isOpen,
  onClose,
  ...filterProps
}: {
  isOpen: boolean
  onClose: () => void
} & FilterPanelProps) {
  const sheetRef = useRef<HTMLDivElement>(null)
  useOnClickOutside(sheetRef, onClose)

  if (!isOpen) return null

  return (
    <>
      <div className="fixed inset-0 bg-black/50 z-40" aria-hidden="true" />
      <div
        ref={sheetRef}
        className="fixed inset-x-0 bottom-0 z-50 bg-background rounded-t-2xl shadow-xl max-h-[80vh] overflow-y-auto
          lg:inset-x-auto lg:bottom-auto lg:left-1/2 lg:top-1/2 lg:w-full lg:max-w-md
          lg:-translate-x-1/2 lg:-translate-y-1/2 lg:rounded-2xl"
        role="dialog"
        aria-modal="true"
        aria-label="フィルター"
      >
        <div className="flex items-center justify-between px-5 py-4 border-b border-border">
          <h2 className="text-base font-semibold text-foreground">フィルター</h2>
          <button
            onClick={onClose}
            className="p-1.5 rounded-md text-muted-foreground hover:bg-muted focus:outline-none focus:ring-2 focus:ring-ring"
            aria-label="閉じる"
          >
            <X className="h-5 w-5" />
          </button>
        </div>
        <div className="p-5">
          <FilterPanel {...filterProps} />
        </div>
      </div>
    </>
  )
}

export default FilterSheet
