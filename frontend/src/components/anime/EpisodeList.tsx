'use client'

import React, { useState } from 'react'
import { cn } from '@/lib/utils'
import type { Episode } from '@/lib/api/generated/myanimelogs'

export interface EpisodeListProps {
  episodes: Episode[]
  isLoading?: boolean
  animeId: number
  onWatchStatusChange?: (episodeId: number, watched: boolean) => void
  className?: string
}

const EpisodeSkeleton: React.FC = () => (
  <div className="flex items-center gap-4 p-4 bg-white border border-gray-200 rounded-lg animate-pulse">
    <div className="w-5 h-5 bg-gray-200 rounded" />
    <div className="flex-1">
      <div className="h-4 bg-gray-200 rounded w-1/4 mb-2" />
      <div className="h-3 bg-gray-200 rounded w-1/2" />
    </div>
  </div>
)

export const EpisodeList: React.FC<EpisodeListProps> = ({
  episodes,
  isLoading = false,
  animeId,
  onWatchStatusChange,
  className,
}) => {
  const [watchedEpisodes, setWatchedEpisodes] = useState<Set<number>>(new Set())
  const [searchQuery, setSearchQuery] = useState('')

  const filteredEpisodes = episodes.filter((episode) => {
    if (!searchQuery) return true
    const query = searchQuery.toLowerCase()
    return (
      episode.title?.toLowerCase().includes(query) ||
      String(episode.episode_number).includes(query)
    )
  })

  const handleToggleWatched = (episodeId: number) => {
    const newWatched = new Set(watchedEpisodes)
    const isWatched = newWatched.has(episodeId)

    if (isWatched) {
      newWatched.delete(episodeId)
    } else {
      newWatched.add(episodeId)
    }

    setWatchedEpisodes(newWatched)
    onWatchStatusChange?.(episodeId, !isWatched)
  }

  const handleMarkAllWatched = () => {
    const allIds = new Set(episodes.map((ep) => ep.id).filter((id): id is number => id !== undefined))
    setWatchedEpisodes(allIds)
    episodes.forEach((ep) => {
      if (ep.id !== undefined) {
        onWatchStatusChange?.(ep.id, true)
      }
    })
  }

  const handleClearAll = () => {
    setWatchedEpisodes(new Set())
    episodes.forEach((ep) => {
      if (ep.id !== undefined) {
        onWatchStatusChange?.(ep.id, false)
      }
    })
  }

  if (isLoading) {
    return (
      <div className={cn('space-y-2', className)}>
        {Array.from({ length: 5 }).map((_, i) => (
          <EpisodeSkeleton key={i} />
        ))}
      </div>
    )
  }

  if (episodes.length === 0) {
    return (
      <div className={cn('text-center py-8 bg-gray-50 rounded-lg', className)}>
        <span className="text-4xl mb-2 block">📭</span>
        <p className="text-gray-500">エピソード情報がありません</p>
      </div>
    )
  }

  const watchedCount = watchedEpisodes.size
  const totalCount = episodes.length
  const progressPercent = totalCount > 0 ? Math.round((watchedCount / totalCount) * 100) : 0

  return (
    <div className={className}>
      {/* ツールバー */}
      <div className="flex flex-col sm:flex-row gap-4 mb-4">
        {/* 検索 */}
        <div className="relative flex-1">
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="エピソードを検索..."
            className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <svg
            className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
            />
          </svg>
        </div>

        {/* 一括操作 */}
        <div className="flex gap-2">
          <button
            onClick={handleMarkAllWatched}
            className="px-3 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            すべて視聴済み
          </button>
          <button
            onClick={handleClearAll}
            className="px-3 py-2 text-sm bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors"
          >
            クリア
          </button>
        </div>
      </div>

      {/* 進捗バー */}
      <div className="mb-4 p-3 bg-gray-50 rounded-lg">
        <div className="flex items-center justify-between mb-2 text-sm">
          <span className="text-gray-600">
            視聴進捗: {watchedCount} / {totalCount} 話
          </span>
          <span className="font-medium text-gray-900">{progressPercent}%</span>
        </div>
        <div className="w-full h-2 bg-gray-200 rounded-full overflow-hidden">
          <div
            className="h-full bg-blue-600 rounded-full transition-all duration-300"
            style={{ width: `${progressPercent}%` }}
          />
        </div>
      </div>

      {/* エピソードリスト */}
      <div className="space-y-2">
        {filteredEpisodes.map((episode) => {
          const isWatched = episode.id !== undefined && watchedEpisodes.has(episode.id)

          return (
            <div
              key={episode.id}
              className={cn(
                'flex items-center gap-4 p-4 bg-white border rounded-lg transition-colors',
                isWatched ? 'border-green-200 bg-green-50' : 'border-gray-200 hover:border-gray-300'
              )}
            >
              {/* チェックボックス */}
              <button
                onClick={() => episode.id !== undefined && handleToggleWatched(episode.id)}
                className={cn(
                  'w-6 h-6 rounded border-2 flex items-center justify-center transition-colors',
                  isWatched
                    ? 'bg-green-500 border-green-500 text-white'
                    : 'border-gray-300 hover:border-blue-500'
                )}
                aria-label={isWatched ? '未視聴にする' : '視聴済みにする'}
              >
                {isWatched && (
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={3}
                      d="M5 13l4 4L19 7"
                    />
                  </svg>
                )}
              </button>

              {/* エピソード情報 */}
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-gray-500">
                    第{episode.episode_number}話
                  </span>
                  {isWatched && (
                    <span className="text-xs text-green-600 font-medium">視聴済み</span>
                  )}
                </div>
                {episode.title && (
                  <p className={cn('text-gray-900 truncate', isWatched && 'text-gray-600')}>
                    {episode.title}
                  </p>
                )}
              </div>
            </div>
          )
        })}
      </div>

      {/* 検索結果なし */}
      {filteredEpisodes.length === 0 && searchQuery && (
        <div className="text-center py-8">
          <p className="text-gray-500">「{searchQuery}」に一致するエピソードが見つかりません</p>
        </div>
      )}
    </div>
  )
}

export default EpisodeList
