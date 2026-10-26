'use client'

import React from 'react'
import { cn } from '@/lib/utils'
import { WatchStatusStatus } from '@/lib/api/generated/myanimelogs'

// 視聴ステータスのタブ（視聴中 / 視聴予定 / 視聴済）。「すべて」「中断」は持たない。
export type WatchStatusTabKey = WatchStatusStatus

interface TabConfig {
  key: WatchStatusTabKey
  label: string
  color: string
}

export const WATCH_STATUS_TABS: TabConfig[] = [
  { key: 'planned', label: '視聴予定', color: 'bg-yellow-500' },
  { key: 'watching', label: '視聴中', color: 'bg-blue-500' },
  { key: 'completed', label: '視聴済', color: 'bg-green-500' },
]

export const WATCH_STATUS_TAB_LABELS: Record<WatchStatusTabKey, string> = {
  watching: '視聴中',
  planned: '視聴予定',
  completed: '視聴済',
}

interface WatchStatusTabsProps {
  activeTab: WatchStatusTabKey
  onTabChange: (tab: WatchStatusTabKey) => void
  counts: Record<WatchStatusTabKey, number>
}

export const WatchStatusTabs: React.FC<WatchStatusTabsProps> = ({
  activeTab,
  onTabChange,
  counts,
}) => {
  // 下線タブ型。件数は各タブにピルとして添え、ヘッダーの記録数カードとの
  // 二重表示をなくす（マイページUI刷新v2）
  return (
    <nav
      className="grid grid-cols-3 md:flex md:gap-7 border-b border-border"
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
              'flex items-center justify-center md:justify-start gap-1.5 md:gap-2',
              '-mb-px border-b-2 pb-3 md:pb-[13px] transition-colors duration-200',
              'focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
              isActive ? 'border-primary' : 'border-transparent'
            )}
            aria-current={isActive ? 'page' : undefined}
          >
            <span
              className={cn(
                'text-[13px] md:text-[14px]',
                isActive ? 'font-bold text-foreground' : 'font-medium text-muted-foreground'
              )}
            >
              {tab.label}
            </span>
            <span
              className={cn(
                'flex h-[18px] min-w-[19px] items-center justify-center rounded-full px-1.5',
                'md:h-5 md:min-w-[22px] md:px-[7px]',
                'text-[10.5px] md:text-[11.5px] font-bold tabular-nums',
                isActive ? 'bg-primary/10 text-primary' : 'bg-muted text-muted-foreground'
              )}
            >
              {counts[tab.key]}
            </span>
          </button>
        )
      })}
    </nav>
  )
}
