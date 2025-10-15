'use client'

import React, { useCallback, useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/contexts/AuthContext'
import { Container } from '@/components/layout'
import { Button } from '@/components/ui/Button'
import { getAnimeStatusConfig } from '@/utils/animeStatus'
import {
  DEFAULT_USER_STATS,
  ERROR_MESSAGES,
  type UserStats,
} from '@/constants/stats'

interface FavoriteAnime {
  id: number
  title: string
  status: string
  imageUrl?: string
}

export const MyPageContent: React.FC = () => {
  const { user, token } = useAuth()
  const router = useRouter()
  const [stats, setStats] = useState<UserStats | null>(null)
  const [favorites, setFavorites] = useState<FavoriteAnime[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const API_BASE_URL =
    process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

  // 統計情報とお気に入りアニメを並列で取得
  const fetchData = useCallback(async () => {
    setIsLoading(true)
    setError(null)

    try {
      const [statsResponse, favoritesResponse] = await Promise.all([
        fetch(`${API_BASE_URL}/api/users/me/stats`, {
          headers: { Authorization: `Bearer ${token}` },
        }),
        fetch(`${API_BASE_URL}/api/users/me/favorites`, {
          headers: { Authorization: `Bearer ${token}` },
        }),
      ])

      // 認証エラーの場合はログインページにリダイレクト
      if (statsResponse.status === 401 || favoritesResponse.status === 401) {
        router.push('/login')
        return
      }

      // 両方のレスポンスを処理してから状態を更新（整合性を保つ）
      let statsData: UserStats
      let favoritesData: FavoriteAnime[]

      // 統計情報の処理
      if (statsResponse.status === 404) {
        statsData = DEFAULT_USER_STATS
      } else if (!statsResponse.ok) {
        throw new Error(ERROR_MESSAGES.STATS_FETCH_FAILED)
      } else {
        statsData = await statsResponse.json()
      }

      // お気に入り情報の処理
      if (favoritesResponse.status === 404) {
        favoritesData = []
      } else if (!favoritesResponse.ok) {
        throw new Error(ERROR_MESSAGES.FAVORITES_FETCH_FAILED)
      } else {
        const data = await favoritesResponse.json()
        // 7件取得して「すべて見る」ボタンの表示判定に使用し、表示は6件のみ
        favoritesData = data.slice(0, 7)
      }

      // 両方のデータが正常に取得できた場合のみ状態を更新
      setStats(statsData)
      setFavorites(favoritesData)
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : ERROR_MESSAGES.DATA_FETCH_FAILED
      setError(errorMessage)
      console.error('Failed to fetch data:', err)
    } finally {
      setIsLoading(false)
    }
  }, [token, API_BASE_URL, router])

  useEffect(() => {
    if (user && token) {
      fetchData()
    }
  }, [user, token, fetchData])

  if (isLoading) {
    return (
      <Container className="py-12">
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
            <p className="text-gray-600">読み込み中...</p>
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
          <Button variant="primary" size="md" onClick={fetchData}>
            再読み込み
          </Button>
        </div>
      </Container>
    )
  }

  return (
    <Container className="py-12">
      <div className="max-w-6xl mx-auto">
        {/* ヘッダーセクション */}
        <div className="bg-white rounded-lg border border-gray-200 p-8 shadow-sm mb-8">
          <div className="flex items-center space-x-6">
            {/* ユーザーアイコン */}
            <div className="w-20 h-20 bg-blue-100 rounded-full flex items-center justify-center flex-shrink-0">
              {user?.icon ? (
                <img
                  src={user.icon}
                  alt={user.name}
                  className="w-20 h-20 rounded-full object-cover"
                />
              ) : (
                <span className="text-blue-600 text-3xl">👤</span>
              )}
            </div>

            {/* ユーザー情報 */}
            <div className="flex-1">
              <h1 className="text-2xl font-bold text-gray-900 mb-2">
                {user?.name}
              </h1>
              {user?.description && (
                <p className="text-gray-700">{user.description}</p>
              )}
            </div>

            {/* 編集ボタン */}
            <div>
              <Button
                variant="outline"
                size="md"
                onClick={() => {
                  // TODO: プロフィール編集ページ実装後に有効化
                  console.info('プロフィール編集機能は準備中です')
                }}
                disabled
              >
                プロフィール編集
              </Button>
            </div>
          </div>
        </div>

        {/* 統計情報セクション */}
        <div className="mb-8">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-2xl font-bold text-gray-900">視聴統計</h2>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => router.push('/stats')}
            >
              詳細を見る →
            </Button>
          </div>

          {stats ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
              {/* 総視聴エピソード数 */}
              <div className="bg-white rounded-lg border border-gray-200 p-6 shadow-sm">
                <div className="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center mb-4">
                  <span className="text-blue-600 text-xl">📺</span>
                </div>
                <p className="text-sm text-gray-600 mb-1">総視聴数</p>
                <p className="text-2xl font-bold text-gray-900">
                  {stats.totalEpisodesWatched || 0}
                </p>
              </div>

              {/* 完了アニメ数 */}
              <div className="bg-white rounded-lg border border-gray-200 p-6 shadow-sm">
                <div className="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center mb-4">
                  <span className="text-green-600 text-xl">✅</span>
                </div>
                <p className="text-sm text-gray-600 mb-1">完了アニメ</p>
                <p className="text-2xl font-bold text-gray-900">
                  {stats.totalAnimesCompleted || 0}
                </p>
              </div>

              {/* 総視聴時間 */}
              <div className="bg-white rounded-lg border border-gray-200 p-6 shadow-sm">
                <div className="w-12 h-12 bg-purple-100 rounded-lg flex items-center justify-center mb-4">
                  <span className="text-purple-600 text-xl">⏱️</span>
                </div>
                <p className="text-sm text-gray-600 mb-1">総視聴時間</p>
                <p className="text-2xl font-bold text-gray-900">
                  {(stats.totalWatchTimeHours || 0).toFixed(1)}h
                </p>
              </div>

              {/* 今月の視聴 */}
              <div className="bg-white rounded-lg border border-gray-200 p-6 shadow-sm">
                <div className="w-12 h-12 bg-orange-100 rounded-lg flex items-center justify-center mb-4">
                  <span className="text-orange-600 text-xl">📅</span>
                </div>
                <p className="text-sm text-gray-600 mb-1">今月の視聴数</p>
                <p className="text-2xl font-bold text-gray-900">
                  {stats.thisMonthEpisodes || 0}
                </p>
              </div>

              {/* お気に入りジャンル */}
              <div className="bg-white rounded-lg border border-gray-200 p-6 shadow-sm">
                <div className="w-12 h-12 bg-pink-100 rounded-lg flex items-center justify-center mb-4">
                  <span className="text-pink-600 text-xl">❤️</span>
                </div>
                <p className="text-sm text-gray-600 mb-1">好きなジャンル</p>
                <div className="text-sm font-medium text-gray-900">
                  {stats.favoriteGenres && stats.favoriteGenres.length > 0 ? (
                    stats.favoriteGenres.slice(0, 2).map((genre, index) => (
                      <span key={index} className="block">
                        {genre}
                      </span>
                    ))
                  ) : (
                    <span className="text-gray-500">-</span>
                  )}
                </div>
              </div>
            </div>
          ) : (
            <div className="bg-white rounded-lg border border-gray-200 p-8 shadow-sm text-center">
              <p className="text-gray-600">統計情報を取得できませんでした</p>
            </div>
          )}
        </div>

        {/* お気に入りアニメセクション */}
        <div>
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-2xl font-bold text-gray-900">
              お気に入りアニメ
            </h2>
            {favorites.length > 6 && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  // TODO: お気に入り一覧ページ実装後に有効化
                  console.info('お気に入り一覧ページは準備中です')
                }}
                disabled
              >
                すべて見る →
              </Button>
            )}
          </div>

          {favorites.length > 0 ? (
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
              {favorites.slice(0, 6).map((anime) => {
                const statusConfig = getAnimeStatusConfig(anime.status)
                return (
                  <div
                    key={anime.id}
                    className="bg-white rounded-lg border border-gray-200 overflow-hidden shadow-sm hover:shadow-md transition-shadow cursor-pointer"
                    onClick={() => {
                      // TODO: アニメ詳細ページ実装後に有効化
                      console.info(
                        `アニメ詳細ページは準備中です (ID: ${anime.id})`
                      )
                    }}
                    role="button"
                    tabIndex={0}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        console.info(
                          `アニメ詳細ページは準備中です (ID: ${anime.id})`
                        )
                      }
                    }}
                  >
                    {anime.imageUrl ? (
                      <img
                        src={anime.imageUrl}
                        alt={anime.title}
                        className="w-full h-48 object-cover"
                      />
                    ) : (
                      <div className="w-full h-48 bg-gray-200 flex items-center justify-center">
                        <span className="text-gray-400 text-4xl">📺</span>
                      </div>
                    )}
                    <div className="p-3">
                      <h3 className="text-sm font-medium text-gray-900 line-clamp-2 mb-2">
                        {anime.title}
                      </h3>
                      <span
                        className={`inline-block px-2 py-1 text-xs rounded-full ${statusConfig.bgColor} ${statusConfig.textColor}`}
                      >
                        {statusConfig.label}
                      </span>
                    </div>
                  </div>
                )
              })}
            </div>
          ) : (
            <div className="bg-white rounded-lg border border-gray-200 p-12 shadow-sm text-center">
              <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <span className="text-gray-400 text-3xl">📺</span>
              </div>
              <p className="text-gray-600 mb-4">
                まだお気に入りのアニメがありません
              </p>
              <Button
                variant="primary"
                size="md"
                onClick={() => router.push('/')}
              >
                アニメを探す
              </Button>
            </div>
          )}
        </div>
      </div>
    </Container>
  )
}
