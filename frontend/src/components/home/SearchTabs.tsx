'use client';

import React from 'react';
import type { SearchTab } from './SearchContainer';

interface SearchTabsProps {
  activeTab: SearchTab;
  onTabChange: (tab: SearchTab) => void;
}

const SearchTabs: React.FC<SearchTabsProps> = ({ activeTab, onTabChange }) => {
  const tabs: { id: SearchTab; label: string }[] = [
    { id: 'keyword', label: 'キーワード' },
    { id: 'season', label: '放送時期' },
    { id: 'genre', label: 'ジャンル' },
  ];

  return (
    <div className="flex gap-1 md:gap-2">
      {tabs.map((tab) => {
        const isActive = activeTab === tab.id;
        return (
          <button
            key={tab.id}
            onClick={() => onTabChange(tab.id)}
            className={`
              relative flex-1 flex items-center justify-center gap-1.5
              px-3 py-3 md:py-3.5 text-xs md:text-sm font-bold
              transition-all duration-200 ease-out rounded-t-xl
              ${
                isActive
                  ? 'bg-white text-blue-700 shadow-lg shadow-black/10 z-10'
                  : 'bg-white/10 text-white/80 hover:bg-white/20 hover:text-white'
              }
            `}
          >
            <span>{tab.label}</span>
            {/* 吹き出し三角形 */}
            {isActive && (
              <span
                className="absolute left-1/2 -translate-x-1/2 -bottom-[6px] w-3 h-3 bg-white rotate-45 shadow-sm"
              />
            )}
          </button>
        );
      })}
    </div>
  );
};

export default SearchTabs;
