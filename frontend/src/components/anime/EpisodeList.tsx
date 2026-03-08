'use client'

import React, { useState, useMemo } from 'react'
import { cn } from '@/lib/utils'
import { Pagination } from '@/components/ui/Pagination'
import { formatEpisodeDuration } from '@/utils/duration'
import type { Episode } from '@/lib/api/generated/myanimelogs'

// 1ページあたりのエピソード数
const EPISODES_PER_PAGE = 10

export interface EpisodeListProps {
  episodes: Episode[]
  isLoading?: boolean
  onWatchStatusChange?: (episodeId: number, watched: boolean) => void
  className?: string
  /** ページ変更時にスクロールするターゲット要素のID */
  scrollTargetId?: string
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
  onWatchStatusChange,
  className,
  scrollTargetId,
}) => {
  const [watchedEpisodes, setWatchedEpisodes] = useState<Set<number>>(new Set())
  const [currentPage, setCurrentPage] = useState(1)

  // ページネーション計算
  const { paginatedEpisodes, totalPages } = useMemo(() => {
    const total = episodes.length
    const pages = Math.ceil(total / EPISODES_PER_PAGE)
    const startIndex = (currentPage - 1) * EPISODES_PER_PAGE
    const endIndex = startIndex + EPISODES_PER_PAGE
    const paginated = episodes.slice(startIndex, endIndex)
    
    return {
      paginatedEpisodes: paginated,
      totalPages: pages,
    }
  }, [episodes, currentPage])

  // ページ変更ハンドラー（スクロール機能付き）
  const handlePageChange = (page: number) => {
    setCurrentPage(page)
    
    // ページ変更時にターゲット要素にスクロール
    if (scrollTargetId) {
      const targetElement = document.getElementById(scrollTargetId)
      if (targetElement) {
        targetElement.scrollIntoView({ behavior: 'smooth', block: 'start' })
      }
    }
  }

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

  // 全エピソードを視聴済みにする
  const handleMarkAllWatched = () => {
    const allIds = episodes.map((ep) => ep.id).filter((id): id is number => id !== undefined)
    const newWatched = new Set([...Array.from(watchedEpisodes), ...allIds])
    setWatchedEpisodes(newWatched)
    episodes.forEach((ep) => {
      if (ep.id !== undefined) {
        onWatchStatusChange?.(ep.id, true)
      }
    })
  }

  // 全エピソードをクリアする
  const handleClearAll = () => {
    const currentIds = new Set(episodes.map((ep) => ep.id).filter((id): id is number => id !== undefined))
    const newWatched = new Set(Array.from(watchedEpisodes).filter((id) => !currentIds.has(id)))
    setWatchedEpisodes(newWatched)
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

  // 視聴進捗は全エピソードを対象にする
  const watchedCount = episodes.filter((ep) => ep.id !== undefined && watchedEpisodes.has(ep.id)).length
  const totalCount = episodes.length
  const progressPercent = totalCount > 0 ? Math.round((watchedCount / totalCount) * 100) : 0

  return (
    <div className={className}>
      {/* ツールバー */}
      <div className="flex flex-col sm:flex-row gap-4 mb-4">
        {/* 一括操作 */}
        <div className="flex gap-2">
          <button
            onClick={handleMarkAllWatched}
            className="px-3 py-2 text-sm text-white rounded-lg transition-colors bg-blue-600 hover:bg-blue-700"
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
            className="h-full rounded-full transition-all duration-300 bg-blue-600"
            style={{ width: `${progressPercent}%` }}
          />
        </div>
      </div>

      {/* エピソードリスト */}
      <div className="space-y-2">
        {paginatedEpisodes.map((episode) => {
          const isWatched = episode.id !== undefined && watchedEpisodes.has(episode.id)

          return (
            <div
              key={episode.id}
              className={cn(
                'flex items-center gap-4 p-4 bg-white border rounded-lg transition-colors',
                isWatched
                  ? 'border-green-200 bg-green-50'
                  : 'border-gray-200 hover:border-gray-300'
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
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="text-sm font-medium text-gray-500">
                    第{episode.episode_number}話
                  </span>
                  {episode.duration_minutes && episode.duration_minutes > 0 && (
                    <span className="text-xs text-gray-500">
                      ({formatEpisodeDuration(episode.duration_minutes)})
                    </span>
                  )}
                  {isWatched && (
                    <span className="text-xs text-green-600 font-medium">視聴済み</span>
                  )}
                  {episode.filler && (
                    <span className="px-1.5 py-0.5 text-xs font-medium bg-yellow-100 text-yellow-700 rounded">
                      フィラー
                    </span>
                  )}
                  {episode.recap && (
                    <span className="px-1.5 py-0.5 text-xs font-medium bg-blue-100 text-blue-700 rounded">
                      総集編
                    </span>
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

      {/* ページネーション */}
      {totalPages > 1 && (
        <Pagination
          currentPage={currentPage}
          totalPages={totalPages}
          onPageChange={handlePageChange}
          className="mt-6"
        />
      )}
    </div>
  )
}

export default EpisodeList
