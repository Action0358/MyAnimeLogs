'use client';

import React, { useState } from 'react';
import SearchTabs from './SearchTabs';
import SearchInputArea from './SearchInputArea';

export type SearchTab = 'keyword' | 'season' | 'genre';

const SearchContainer: React.FC = () => {
  const [activeTab, setActiveTab] = useState<SearchTab>('keyword');

  return (
    <div className="w-full max-w-[800px] relative z-0">
      <SearchTabs activeTab={activeTab} onTabChange={setActiveTab} />
      <SearchInputArea activeTab={activeTab} />
    </div>
  );
};

export default SearchContainer;
