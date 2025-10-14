'use client'

import React, { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/contexts/AuthContext'
import { Container } from '@/components/layout'
import { Button } from '@/components/ui/Button'

interface UserStats {
  totalEpisodesWatched: number
  totalAnimesCompleted: number
  totalWatchTimeHours: number
  favoriteGenres: string[]
  thisMonthEpisodes: number
}

interface FavoriteAnime {
  id: number
  title: string
  status: string
  imageUrl?: string
}

const MyPage: React.FC = () => {
  const { user, token, isLoading: authLoading } = useAuth()
  const router = useRouter()
  const [stats, setStats] = useState<UserStats | null>(null)
  const [favorites, setFavorites] = useState<FavoriteAnime[]>([])
  const [isLoadingStats, setIsLoadingStats] = useState(true)
  const [isLoadingFavorites, setIsLoadingFavorites] = useState(true)

  const API_BASE_URL =
    process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

  // 認証チェック
  useEffect(() => {
    if (!authLoading && !user) {
      router.push('/login')
    }
  }, [user, authLoading, router])

  // 統計情報取得
  useEffect(() => {
    const fetchStats = async () => {
      if (!token) return

      try {
        const response = await fetch(`${API_BASE_URL}/api/users/me/stats`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        })

        if (response.ok) {
          const data = await response.json()
          setStats(data)
        }
      } catch (error) {
        console.error('Failed to fetch stats:', error)
      } finally {
        setIsLoadingStats(false)
      }
    }

    if (user && token) {
      fetchStats()
    }
  }, [user, token, API_BASE_URL])

  // お気に入りアニメ取得
  useEffect(() => {
    const fetchFavorites = async () => {
      if (!token) return

      try {
        const response = await fetch(`${API_BASE_URL}/api/users/me/favorites`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        })

        if (response.ok) {
          const data = await response.json()
          setFavorites(data.slice(0, 6)) // 最大6件表示
        }
      } catch (error) {
        console.error('Failed to fetch favorites:', error)
      } finally {
        setIsLoadingFavorites(false)
      }
    }

    if (user && token) {
      fetchFavorites()
    }
  }, [user, token, API_BASE_URL])

  if (authLoading || !user) {
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

  return (
    <Container className="py-12">
      <div className="max-w-6xl mx-auto">
        {/* ヘッダーセクション */}
        <div className="bg-white rounded-lg border border-gray-200 p-8 shadow-sm mb-8">
          <div className="flex items-center space-x-6">
            {/* ユーザーアイコン */}
            <div className="w-20 h-20 bg-blue-100 rounded-full flex items-center justify-center flex-shrink-0">
              {user.icon ? (
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
                {user.name}
              </h1>
              <p className="text-gray-600 mb-4">{user.email}</p>
              {user.description && (
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
                  alert('プロフィール編集機能は準備中です')
                }}
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

          {isLoadingStats ? (
            <div className="bg-white rounded-lg border border-gray-200 p-8 shadow-sm">
              <div className="flex items-center justify-center">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
              </div>
            </div>
          ) : stats ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
              {/* 総視聴エピソード数 */}
              <div className="bg-white rounded-lg border border-gray-200 p-6 shadow-sm">
                <div className="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center mb-4">
                  <span className="text-blue-600 text-xl">📺</span>
                </div>
                <p className="text-sm text-gray-600 mb-1">総視聴エピソード</p>
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
                <p className="text-sm text-gray-600 mb-1">今月の視聴</p>
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
                  alert('お気に入り一覧ページは準備中です')
                }}
              >
                すべて見る →
              </Button>
            )}
          </div>

          {isLoadingFavorites ? (
            <div className="bg-white rounded-lg border border-gray-200 p-8 shadow-sm">
              <div className="flex items-center justify-center">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
              </div>
            </div>
          ) : favorites.length > 0 ? (
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
              {favorites.map((anime) => (
                <div
                  key={anime.id}
                  className="bg-white rounded-lg border border-gray-200 overflow-hidden shadow-sm hover:shadow-md transition-shadow cursor-pointer"
                  onClick={() => {
                    // TODO: アニメ詳細ページ実装後に有効化
                    alert(`アニメ詳細ページは準備中です (ID: ${anime.id})`)
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
                      className={`inline-block px-2 py-1 text-xs rounded-full ${
                        anime.status === 'watching'
                          ? 'bg-blue-100 text-blue-700'
                          : anime.status === 'completed'
                            ? 'bg-green-100 text-green-700'
                            : anime.status === 'planned'
                              ? 'bg-gray-100 text-gray-700'
                              : 'bg-orange-100 text-orange-700'
                      }`}
                    >
                      {anime.status === 'watching'
                        ? '視聴中'
                        : anime.status === 'completed'
                          ? '完了'
                          : anime.status === 'planned'
                            ? '予定'
                            : anime.status}
                    </span>
                  </div>
                </div>
              ))}
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

export default MyPage
