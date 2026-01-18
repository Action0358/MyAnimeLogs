'use client';

import React from 'react';
import type { SearchTab } from './SearchContainer';

interface SearchTabsProps {
  activeTab: SearchTab;
  onTabChange: (tab: SearchTab) => void;
}

const SearchTabs: React.FC<SearchTabsProps> = ({ activeTab, onTabChange }) => {
  const tabs: { id: SearchTab; label: string }[] = [
    { id: 'keyword', label: 'キーワードから探す' },
    { id: 'season', label: '放送時期から探す' },
    { id: 'genre', label: 'ジャンル・特徴' },
  ];

  return (
    <div className="flex gap-2 mb-4 flex-wrap">
      {tabs.map((tab) => (
        <button
          key={tab.id}
          onClick={() => onTabChange(tab.id)}
          className={`
            px-4 py-2 rounded-md text-sm md:text-base font-bold transition-colors
            ${
              activeTab === tab.id
                ? 'bg-blue-600 text-white'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }
          `}
        >
          {tab.label}
        </button>
      ))}
    </div>
  );
};

export default SearchTabs;
