'use client'

import React from 'react'
import type { SearchTab } from './SearchContainer'

interface SearchTabsProps {
  activeTab: SearchTab
  onTabChange: (tab: SearchTab) => void
}

const SearchTabs: React.FC<SearchTabsProps> = ({ activeTab, onTabChange }) => {
  const tabs: { id: SearchTab; label: string }[] = [
    { id: 'keyword', label: 'キーワード' },
    { id: 'genre', label: 'ジャンル' },
    { id: 'season', label: '放送時期' },
  ]

  return (
    <div className="flex gap-1 md:gap-2" role="tablist" aria-label="検索方法">
      {tabs.map((tab) => {
        const isActive = activeTab === tab.id
        return (
          <button
            key={tab.id}
            role="tab"
            aria-selected={isActive}
            onClick={() => onTabChange(tab.id)}
            className={`
              relative flex-1 flex h-11 select-none items-center justify-center gap-1.5
              px-3 text-xs md:text-sm font-bold
              transition-all duration-200 ease-out rounded-t-xl
              focus:outline-none focus-visible:ring-2 focus-visible:ring-primary-foreground/70
              ${
                isActive
                  ? 'bg-card text-primary shadow-lg shadow-black/10 z-10'
                  : 'bg-primary-foreground/10 text-primary-foreground/80 hover:bg-primary-foreground/20 hover:text-primary-foreground'
              }
            `}
          >
            <span>{tab.label}</span>
            {/* 吹き出し三角形 */}
            {isActive && (
              <span className="absolute left-1/2 -translate-x-1/2 -bottom-[6px] w-3 h-3 bg-card rotate-45 shadow-sm" />
            )}
          </button>
        )
      })}
    </div>
  )
}

export default SearchTabs
