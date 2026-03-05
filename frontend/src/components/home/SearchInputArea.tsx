'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import type { Route } from 'next';
import type { SearchTab } from './SearchContainer';
import { SEASONS, GENRES } from '@/constants/search';

interface SearchInputAreaProps {
  activeTab: SearchTab;
}

const SearchInputArea: React.FC<SearchInputAreaProps> = ({ activeTab }) => {
  const router = useRouter();
  const [keyword, setKeyword] = useState('');
  const [season, setSeason] = useState('');
  const [selectedGenres, setSelectedGenres] = useState<string[]>([]);

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
      router.push(`/search${searchQuery}` as Route);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSearch();
    }
  };

  const searchButtonBase = 'bg-blue-600 text-white font-bold text-base hover:bg-blue-700 active:bg-blue-800 transition-all duration-150 tracking-widest disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:bg-blue-600';

  const searchButton = (disabled = false) => (
    <button
      onClick={handleSearch}
      disabled={disabled}
      className={`px-6 md:px-10 py-4 whitespace-nowrap ${searchButtonBase}`}
    >
      検 索
    </button>
  );

  return (
    <div className="relative mt-3 backdrop-blur-sm bg-white/10 border border-white/20 rounded-2xl p-4 md:p-5 shadow-lg shadow-black/10">
      {/* キーワード検索 */}
      {activeTab === 'keyword' && (
        <div className="flex overflow-hidden rounded-xl shadow-sm">
          <div className="relative flex-1">
            <svg
              className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              aria-hidden="true"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
              />
            </svg>
            <input
              type="text"
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="作品を探す"
              className="w-full pl-12 pr-4 py-4 bg-white text-gray-900
                placeholder:text-gray-400 placeholder:font-bold focus:outline-none text-base font-bold"
            />
          </div>
          {searchButton()}
        </div>
      )}

      {/* 放送時期検索 */}
      {activeTab === 'season' && (
        <div className="flex overflow-hidden rounded-xl shadow-sm">
          <div className="relative flex-1">
            <svg
              className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400 pointer-events-none"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              aria-hidden="true"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
              />
            </svg>
            <select
              value={season}
              onChange={(e) => setSeason(e.target.value)}
              className={`w-full pl-12 pr-4 py-4 bg-white
                focus:outline-none text-base font-bold appearance-none cursor-pointer
                ${season ? 'text-gray-900' : 'text-gray-400'}`}
            >
              <option value="" disabled hidden>放送時期を選択</option>
              {SEASONS.map((s) => (
                <option key={s} value={s}>
                  {s}
                </option>
              ))}
            </select>
          </div>
          {searchButton(!season)}
        </div>
      )}

      {/* ジャンル検索 */}
      {activeTab === 'genre' && (
        <div className="space-y-4">
          <div className="flex flex-wrap justify-center gap-1.5">
            {GENRES.map((genre) => (
              <button
                key={genre}
                onClick={() => handleGenreToggle(genre)}
                className={`
                  px-3.5 py-1.5 rounded-full text-[13px] font-bold
                  transition-all duration-150
                  ${
                    selectedGenres.includes(genre)
                      ? 'bg-white text-blue-700 shadow-md scale-105'
                      : 'bg-white/15 text-white/90 hover:bg-white/25 hover:scale-105'
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
            className={`w-full py-4 rounded-xl shadow-sm ${searchButtonBase}`}
          >
            検 索
          </button>
        </div>
      )}
    </div>
  );
};

export default SearchInputArea;
