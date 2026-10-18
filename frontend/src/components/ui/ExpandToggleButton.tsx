'use client'

import React from 'react'
import { ChevronDown, ChevronUp } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface ExpandToggleButtonProps {
  isExpanded: boolean
  hiddenCount: number
  onClick: () => void
  className?: string
}

/** 件数の多いグリッドの折りたたみ/展開ボタン。関連作品(RelatedAnimeGrid)と見た目を揃える */
export const ExpandToggleButton: React.FC<ExpandToggleButtonProps> = ({
  isExpanded,
  hiddenCount,
  onClick,
  className,
}) => (
  <button
    type="button"
    onClick={onClick}
    aria-expanded={isExpanded}
    className={cn(
      'mt-4 inline-flex items-center gap-1 rounded text-sm font-medium text-primary hover:underline focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
      className
    )}
  >
    {isExpanded ? (
      <>
        折りたたむ
        <ChevronUp className="w-4 h-4" />
      </>
    ) : (
      <>
        他{hiddenCount}件を表示
        <ChevronDown className="w-4 h-4" />
      </>
    )}
  </button>
)

export default ExpandToggleButton
