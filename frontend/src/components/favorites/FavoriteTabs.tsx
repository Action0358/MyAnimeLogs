'use client'

import React from 'react'
import { cn } from '@/lib/utils'
import { FavoriteStatus } from '@/lib/api/generated/myanimelogs'
import { STATUS_CONFIG } from './StatusSelector'

export type TabKey = FavoriteStatus | 'all'

interface TabConfig {
  key: TabKey
  label: string
  color?: string
}

const TABS: TabConfig[] = [
  { key: 'all', label: 'すべて' },
  { key: 'watching', label: '視聴中', color: 'bg-blue-500' },
  { key: 'completed', label: '完了', color: 'bg-green-500' },
  { key: 'planned', label: '視聴予定', color: 'bg-yellow-500' },
  { key: 'dropped', label: '中断', color: 'bg-red-500' },
]

interface FavoriteTabsProps {
  activeTab: TabKey
  onTabChange: (tab: TabKey) => void
  counts?: Record<TabKey, number>
}

export const FavoriteTabs: React.FC<FavoriteTabsProps> = ({
  activeTab,
  onTabChange,
  counts,
}) => {
  return (
    <div className="border-b border-gray-200">
      <nav className="-mb-px flex space-x-4 overflow-x-auto" aria-label="お気に入りタブ">
        {TABS.map((tab) => {
          const isActive = activeTab === tab.key
          const count = counts?.[tab.key]
          
          return (
            <button
              key={tab.key}
              type="button"
              onClick={() => onTabChange(tab.key)}
              className={cn(
                'flex items-center gap-2 whitespace-nowrap py-3 px-1 border-b-2 text-sm font-medium transition-colors',
                isActive
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              )}
              aria-current={isActive ? 'page' : undefined}
            >
              {tab.color && (
                <span className={cn('w-2 h-2 rounded-full', tab.color)} />
              )}
              <span>{tab.label}</span>
              {count !== undefined && (
                <span
                  className={cn(
                    'inline-flex items-center justify-center px-2 py-0.5 text-xs rounded-full',
                    isActive
                      ? 'bg-blue-100 text-blue-600'
                      : 'bg-gray-100 text-gray-600'
                  )}
                >
                  {count}
                </span>
              )}
            </button>
          )
        })}
      </nav>
    </div>
  )
}
