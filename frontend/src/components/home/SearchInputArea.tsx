'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import type { SearchTab } from './SearchContainer';

interface SearchInputAreaProps {
  activeTab: SearchTab;
}

const SearchInputArea: React.FC<SearchInputAreaProps> = ({ activeTab }) => {
  const router = useRouter();
  const [keyword, setKeyword] = useState('');
  const [season, setSeason] = useState('');
  const [selectedGenres, setSelectedGenres] = useState<string[]>([]);

  const seasons = [
    '2025年冬',
    '2024年秋',
    '2024年夏',
    '2024年春',
    '2024年冬',
    '2023年秋',
  ];

  const genres = [
    'アクション',
    'ラブコメ',
    '異世界',
    'SF',
    'ファンタジー',
    'ドラマ',
    'コメディ',
    'ホラー',
  ];

  const handleGenreToggle = (genre: string) => {
    setSelectedGenres((prev) =>
      prev.includes(genre)
        ? prev.filter((g) => g !== genre)
        : [...prev, genre]
    );
  };

  const handleSearch = () => {
    let searchQuery = '';

    if (activeTab === 'keyword' && keyword.trim()) {
      searchQuery = `?keyword=${encodeURIComponent(keyword.trim())}`;
    } else if (activeTab === 'season' && season) {
      searchQuery = `?season=${encodeURIComponent(season)}`;
    } else if (activeTab === 'genre' && selectedGenres.length > 0) {
      searchQuery = `?genre=${selectedGenres.map(g => encodeURIComponent(g)).join(',')}`;
    }

    if (searchQuery) {
      router.push(`/search${searchQuery}` as any);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSearch();
    }
  };

  return (
    <div className="space-y-4">
      {/* キーワード検索 */}
      {activeTab === 'keyword' && (
        <div className="flex flex-col md:flex-row gap-2">
          <input
            type="text"
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            onKeyPress={handleKeyPress}
            placeholder="作品名"
            className="flex-1 px-4 py-3 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900"
          />
          <button
            onClick={handleSearch}
            className="px-6 py-3 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors font-medium flex items-center justify-center gap-2 w-full md:w-auto"
          >
            <span>🔍</span>
            <span>検索</span>
          </button>
        </div>
      )}

      {/* 放送時期検索 */}
      {activeTab === 'season' && (
        <div className="flex flex-col md:flex-row gap-2">
          <select
            value={season}
            onChange={(e) => setSeason(e.target.value)}
            className="flex-1 px-4 py-3 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900"
          >
            <option value="">放送時期を選択</option>
            {seasons.map((s) => (
              <option key={s} value={s}>
                {s}
              </option>
            ))}
          </select>
          <button
            onClick={handleSearch}
            disabled={!season}
            className="px-6 py-3 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors font-medium flex items-center justify-center gap-2 w-full md:w-auto disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <span>🔍</span>
            <span>検索</span>
          </button>
        </div>
      )}

      {/* ジャンル検索 */}
      {activeTab === 'genre' && (
        <div className="space-y-4">
          <div className="flex flex-wrap gap-2">
            {genres.map((genre) => (
              <button
                key={genre}
                onClick={() => handleGenreToggle(genre)}
                className={`
                  px-4 py-2 rounded-md text-sm font-medium transition-colors
                  ${
                    selectedGenres.includes(genre)
                      ? 'bg-blue-600 text-white'
                      : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                  }
                `}
              >
                {genre}
              </button>
            ))}
          </div>
          <button
            onClick={handleSearch}
            disabled={selectedGenres.length === 0}
            className="w-full px-6 py-3 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors font-medium flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <span>🔍</span>
            <span>検索</span>
          </button>
        </div>
      )}
    </div>
  );
};

export default SearchInputArea;
