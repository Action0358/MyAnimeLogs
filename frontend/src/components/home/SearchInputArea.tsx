'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import type { Route } from 'next';
import type { SearchTab } from './SearchContainer';
import { GENRES, HERO_SEASON_OPTIONS } from '@/constants/search';

interface SearchInputAreaProps {
  activeTab: SearchTab;
}

const SearchInputArea: React.FC<SearchInputAreaProps> = ({ activeTab }) => {
  const router = useRouter();
  const [keyword, setKeyword] = useState('');
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
    } else if (activeTab === 'genre' && selectedGenres.length > 0) {
      searchQuery = `?genre=${selectedGenres.map(g => encodeURIComponent(g)).join(',')}`;
    }

    if (searchQuery) {
      router.push(`/search${searchQuery}` as Route);
    }
  };

  const handleSeasonSelect = (season: string) => {
    router.push(`/search?season=${encodeURIComponent(season)}` as Route);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSearch();
    }
  };

  const searchButtonBase =
    'bg-primary text-primary-foreground font-bold text-base hover:bg-primary/90 active:bg-primary/80 transition-all duration-150 tracking-widest focus:outline-none focus-visible:ring-2 focus-visible:ring-primary-foreground/70 disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:bg-primary';

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
    <div className="relative mt-3 backdrop-blur-md bg-primary-foreground/10 border border-primary-foreground/20 rounded-2xl p-4 md:p-5 shadow-xl shadow-black/20">
      {/* キーワード検索 */}
      {activeTab === 'keyword' && (
        <div className="flex overflow-hidden rounded-xl shadow-sm">
          <div className="relative flex-1">
            <svg
              className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground"
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
              className="w-full pl-12 pr-4 py-4 bg-card text-foreground
                placeholder:text-muted-foreground placeholder:font-bold focus:outline-none text-base font-bold"
            />
          </div>
          {searchButton(!keyword.trim())}
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
                  focus:outline-none focus-visible:ring-2 focus-visible:ring-primary-foreground/70
                  ${
                    selectedGenres.includes(genre)
                      ? 'bg-card text-primary shadow-md scale-105'
                      : 'bg-primary-foreground/15 text-primary-foreground/90 hover:bg-primary-foreground/25 hover:scale-105'
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

      {/* 放送時期検索 */}
      {activeTab === 'season' && (
        <div className="grid grid-cols-3 gap-2">
          {HERO_SEASON_OPTIONS.map(({ tag, season }) => (
            <button
              key={season}
              onClick={() => handleSeasonSelect(season)}
              className="flex flex-col items-center justify-center gap-1 py-4 rounded-xl bg-card text-foreground hover:bg-card/80 transition-all duration-150 shadow-sm focus:outline-none focus-visible:ring-2 focus-visible:ring-primary-foreground/70"
            >
              <span className="text-xs font-bold text-primary">{tag}</span>
              <span className="text-sm font-bold">{season}</span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
};

export default SearchInputArea;
