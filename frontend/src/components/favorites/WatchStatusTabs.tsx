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
}

export const WatchStatusTabs: React.FC<WatchStatusTabsProps> = ({
  activeTab,
  onTabChange,
}) => {
  // 一覧・検索ページの MediaTabs と同じピル型に揃える。
  // 件数はヘッダーの記録数カードに表示済みのため、ここでは出さない。
  return (
    <nav
      className="flex gap-1 overflow-x-auto scrollbar-hide touch-pan-x"
      aria-label="視聴ステータスタブ"
    >
      {WATCH_STATUS_TABS.map((tab) => {
        const isActive = activeTab === tab.key

        return (
          <button
            key={tab.key}
            type="button"
            onClick={() => onTabChange(tab.key)}
            className={cn(
              'relative rounded-md px-3 py-1.5 md:px-4 md:py-2 text-[12.5px] md:text-[13.5px]',
              'whitespace-nowrap flex-shrink-0 transition-colors duration-200',
              'focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
              isActive
                ? 'bg-primary/10 text-primary font-semibold'
                : 'text-muted-foreground font-medium hover:text-foreground hover:bg-muted'
            )}
            aria-current={isActive ? 'page' : undefined}
          >
            {tab.label}
          </button>
        )
      })}
    </nav>
  )
}
