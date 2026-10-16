'use client'

import React from 'react'
import { Search } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from './Button'

export interface EmptyStateProps {
  title: string
  description?: string
  /** lucide-react のアイコン。stroke 1.75・text-slate-400で統一し、円（bg-slate-100）に収める */
  icon?: React.ReactNode
  action?: {
    label: string
    onClick: () => void
  }
  className?: string
}

/**
 * 空状態の共通表示。アイコンは72px（SP: 64px）の円に収め、見出し＋説明＋ボタンの
 * 3段構成に統一する（空状態UI刷新）。
 */
export const EmptyState: React.FC<EmptyStateProps> = ({
  title,
  description,
  icon,
  action,
  className,
}) => {
  return (
    <div className={cn('flex flex-col items-center justify-center py-12 text-center', className)}>
      {icon && (
        <div className="w-16 h-16 md:w-[72px] md:h-[72px] rounded-full bg-slate-100 flex items-center justify-center mb-[18px] md:mb-5 text-slate-400">
          {icon}
        </div>
      )}
      <h3 className="text-base md:text-[17px] font-bold text-foreground mb-1.5 md:mb-2">{title}</h3>
      {description && (
        <p className="text-[13.5px] md:text-sm leading-[1.7] text-muted-foreground mb-[22px] md:mb-6 max-w-md">
          {description}
        </p>
      )}
      {action && (
        <Button variant="primary" size="md" onClick={action.onClick} className="gap-2">
          <Search className="w-4 h-4" aria-hidden="true" />
          {action.label}
        </Button>
      )}
    </div>
  )
}

export default EmptyState
