'use client'

import React from 'react'
import { cn } from '@/lib/utils'
import type { MediaType } from '@/utils/mediaFilter'

export interface MediaTabDefinition {
  id: MediaType
  label: string
}

interface MediaTabsProps {
  tabs: MediaTabDefinition[]
  activeTab: MediaType
  onTabChange: (tab: MediaType) => void
  className?: string
}

export const MediaTabs: React.FC<MediaTabsProps> = ({
  tabs,
  activeTab,
  onTabChange,
  className,
}) => (
  <div
    className={cn(
      'flex gap-1.5 overflow-x-auto scrollbar-hide touch-pan-x',
      className
    )}
    role="tablist"
    aria-label="メディアタイプフィルター"
  >
    {tabs.map((tab) => {
      const isActive = activeTab === tab.id
      return (
        <button
          key={tab.id}
          role="tab"
          aria-selected={isActive}
          onClick={() => onTabChange(tab.id)}
          className={cn(
            'relative rounded-md px-3 py-1.5 md:px-3.5 md:py-[7px] text-[12.5px] md:text-[13px]',
            'whitespace-nowrap flex-shrink-0 transition-colors duration-200',
            'focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
            isActive
              ? 'bg-primary/10 text-primary font-semibold'
              : 'text-muted-foreground font-medium hover:text-foreground hover:bg-muted'
          )}
        >
          <span>{tab.label}</span>
        </button>
      )
    })}
  </div>
)

export default MediaTabs
