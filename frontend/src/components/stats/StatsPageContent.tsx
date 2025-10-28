'use client'

import React, { useCallback, useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/contexts/AuthContext'
import { Container } from '@/components/layout'
import { Button } from '@/components/ui/Button'
import {
  DEFAULT_USER_STATS,
  ERROR_MESSAGES,
  HOURS_PER_EPISODE,
  MINUTES_PER_EPISODE,
  type UserStats,
} from '@/constants/stats'
import { get, ApiClientError } from '@/lib/api/client'

export const StatsPageContent: React.FC = () => {
  const { token } = useAuth()
  const router = useRouter()
  const [stats, setStats] = useState<UserStats | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchStats = useCallback(async () => {
    setIsLoading(true)
    setError(null)

    try {
      const statsData = await get<UserStats>('/api/users/me/stats').catch(
        (err) => {
          if (err instanceof ApiClientError) {
            if (err.status === 401) {
              router.push('/login')
              return null
            }
            if (err.status === 404) {
              return DEFAULT_USER_STATS
            }
          }
          throw new Error(ERROR_MESSAGES.STATS_FETCH_FAILED)
        }
      )

      // データが正常に取得できた場合のみ状態を更新
      if (statsData) {
        setStats(statsData)
      }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : ERROR_MESSAGES.STATS_FETCH_FAILED
      setError(errorMessage)
    } finally {
      setIsLoading(false)
    }
  }, [router])

  useEffect(() => {
    if (token) {
      fetchStats()
    }
  }, [token, fetchStats])

  if (isLoading) {
    return (
      <Container className="py-12">
        <div className="bg-white rounded-lg border border-gray-200 p-12 shadow-sm">
          <div className="flex items-center justify-center">
            <div className="text-center">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
              <p className="text-gray-600">統計情報を読み込み中...</p>
            </div>
          </div>
        </div>
      </Container>
    )
  }

  if (error) {
    return (
      <Container className="py-12">
        <div className="bg-white rounded-lg border border-gray-200 p-12 shadow-sm text-center">
          <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <span className="text-red-600 text-3xl">⚠️</span>
          </div>
          <p className="text-red-600 mb-4">{error}</p>
          <Button variant="primary" size="md" onClick={fetchStats}>
            再読み込み
          </Button>
        </div>
      </Container>
    )
  }

  if (!stats) {
    return null
  }

  return (
    <Container className="py-12">
      <div className="max-w-6xl mx-auto">
        <div className="mb-8">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 mb-2">
                視聴統計
              </h1>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => router.push('/mypage')}
            >
              ← マイページに戻る
            </Button>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <div className="bg-white rounded-lg border border-gray-200 p-6 shadow-sm">
            <div className="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center mb-4">
              <span className="text-blue-600 text-2xl">📺</span>
            </div>
            <p className="text-sm text-gray-600 mb-1">エピソード総数</p>
            <p className="text-3xl font-bold text-gray-900 mb-2">
              {stats.totalEpisodesWatched || 0}
            </p>
            <p className="text-xs text-gray-500">
              視聴したエピソード総数
            </p>
          </div>

          <div className="bg-white rounded-lg border border-gray-200 p-6 shadow-sm">
            <div className="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center mb-4">
              <span className="text-green-600 text-2xl">✅</span>
            </div>
            <p className="text-sm text-gray-600 mb-1">完了アニメ</p>
            <p className="text-3xl font-bold text-gray-900 mb-2">
              {stats.totalAnimesCompleted || 0}
            </p>
            <p className="text-xs text-gray-500">
              最後まで視聴した作品数
            </p>
          </div>

          <div className="bg-white rounded-lg border border-gray-200 p-6 shadow-sm">
            <div className="w-12 h-12 bg-purple-100 rounded-lg flex items-center justify-center mb-4">
              <span className="text-purple-600 text-2xl">⏱️</span>
            </div>
            <p className="text-sm text-gray-600 mb-1">総視聴時間</p>
            <p className="text-3xl font-bold text-gray-900 mb-2">
              {(stats.totalWatchTimeHours || 0).toFixed(1)}
              <span className="text-lg text-gray-600 ml-1">時間</span>
            </p>
            <p className="text-xs text-gray-500">
              約{Math.floor((stats.totalWatchTimeHours || 0) / 24)}
              日分の視聴時間
            </p>
          </div>

          <div className="bg-white rounded-lg border border-gray-200 p-6 shadow-sm">
            <div className="w-12 h-12 bg-orange-100 rounded-lg flex items-center justify-center mb-4">
              <span className="text-orange-600 text-2xl">📅</span>
            </div>
            <p className="text-sm text-gray-600 mb-1">今月のエピソード数</p>
            <p className="text-3xl font-bold text-gray-900 mb-2">
              {stats.thisMonthEpisodes || 0}
            </p>
            <p className="text-xs text-gray-500">視聴したエピソード数</p>
          </div>
        </div>

        <div className="bg-white rounded-lg border border-gray-200 p-8 shadow-sm mb-8">
          <div className="flex items-center mb-6">
            <div className="w-12 h-12 bg-pink-100 rounded-lg flex items-center justify-center mr-4">
              <span className="text-pink-600 text-2xl">❤️</span>
            </div>
            <div>
              <h2 className="text-xl font-bold text-gray-900">
                お気に入りジャンル
              </h2>
              <p className="text-sm text-gray-600">よく視聴するジャンルTOP3</p>
            </div>
          </div>

          {stats.favoriteGenres && stats.favoriteGenres.length > 0 ? (
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {stats.favoriteGenres.map((genre, index) => (
                <div
                  key={index}
                  className="bg-gradient-to-br from-blue-50 to-purple-50 rounded-lg p-6 border border-blue-100"
                >
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-2xl font-bold text-blue-600">
                      #{index + 1}
                    </span>
                    <span className="text-3xl">
                      {index === 0 ? '🥇' : index === 1 ? '🥈' : '🥉'}
                    </span>
                  </div>
                  <p className="text-lg font-semibold text-gray-900">{genre}</p>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8">
              <p className="text-gray-600">まだジャンルデータがありません</p>
            </div>
          )}
        </div>

        <div className="bg-white rounded-lg border border-gray-200 p-8 shadow-sm">
          <div className="flex items-center mb-6">
            <div className="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center mr-4">
              <span className="text-green-600 text-2xl">📈</span>
            </div>
            <div>
              <h2 className="text-xl font-bold text-gray-900">視聴ペース</h2>
              <p className="text-sm text-gray-600">あなたの視聴傾向</p>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="text-center p-6 bg-gray-50 rounded-lg">
              <p className="text-sm text-gray-600 mb-2">今月の平均視聴時間</p>
              <p className="text-2xl font-bold text-gray-900">
                {((stats.thisMonthEpisodes || 0) * HOURS_PER_EPISODE).toFixed(
                  1
                )}
                <span className="text-sm text-gray-600 ml-1">時間</span>
              </p>
              <p className="text-xs text-gray-500 mt-1">
                ※1話{MINUTES_PER_EPISODE}分として計算
              </p>
            </div>

            <div className="text-center p-6 bg-gray-50 rounded-lg">
              <p className="text-sm text-gray-600 mb-2">
                1作品の平均エピソード数
              </p>
              <p className="text-2xl font-bold text-gray-900">
                {stats.totalAnimesCompleted > 0
                  ? Math.floor(
                      (stats.totalEpisodesWatched ?? 0) /
                        stats.totalAnimesCompleted
                    )
                  : 0}
                <span className="text-sm text-gray-600 ml-1">話</span>
              </p>
            </div>

            <div className="text-center p-6 bg-gray-50 rounded-lg">
              <p className="text-sm text-gray-600 mb-2">今月の視聴ペース</p>
              <p className="text-2xl font-bold text-gray-900">
                {(stats.thisMonthEpisodes || 0) > 0 ? '📈 活発' : '📉 低調'}
              </p>
            </div>
          </div>
        </div>

        <div className="mt-8 text-center">
          <Button variant="primary" size="lg" onClick={() => router.push('/')}>
            もっとアニメを探す
          </Button>
        </div>
      </div>
    </Container>
  )
}
