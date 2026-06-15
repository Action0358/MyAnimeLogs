'use client'

import React from 'react'
import { cn } from '@/lib/utils'
import { WatchStatusStatus } from '@/lib/api/generated/myanimelogs'

// 視聴ステータスのタブ（視聴中 / 視聴予定 / 完了）。「すべて」「中断」は持たない。
export type WatchStatusTabKey = WatchStatusStatus

interface TabConfig {
  key: WatchStatusTabKey
  label: string
  color: string
}

export const WATCH_STATUS_TABS: TabConfig[] = [
  { key: 'planned', label: '視聴予定', color: 'bg-yellow-500' },
  { key: 'watching', label: '視聴中', color: 'bg-blue-500' },
  { key: 'completed', label: '完了', color: 'bg-green-500' },
]

export const WATCH_STATUS_TAB_LABELS: Record<WatchStatusTabKey, string> = {
  watching: '視聴中',
  planned: '視聴予定',
  completed: '完了',
}

interface WatchStatusTabsProps {
  activeTab: WatchStatusTabKey
  onTabChange: (tab: WatchStatusTabKey) => void
  counts?: Record<WatchStatusTabKey, number>
}

export const WatchStatusTabs: React.FC<WatchStatusTabsProps> = ({
  activeTab,
  onTabChange,
  counts,
}) => {
  return (
    <div className="border-b border-gray-200">
      <nav className="-mb-px flex space-x-4 overflow-x-auto" aria-label="視聴ステータスタブ">
        {WATCH_STATUS_TABS.map((tab) => {
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
