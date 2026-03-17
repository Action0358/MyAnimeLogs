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
      'flex gap-1 border-b border-gray-200 overflow-hidden',
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
            'relative px-3 py-2 md:px-4 md:py-3 text-xs md:text-sm font-medium',
            'whitespace-nowrap flex-shrink-0 transition-colors duration-200',
            'border-b-2 hover:text-blue-600 hover:bg-blue-50/50',
            'focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2',
            isActive
              ? 'text-blue-600 border-blue-600 bg-blue-50/30'
              : 'text-gray-600 border-transparent'
          )}
        >
          <span>{tab.label}</span>
        </button>
      )
    })}
  </div>
)

export default MediaTabs
