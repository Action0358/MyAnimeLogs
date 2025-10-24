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

type TabType = 'favorites' | 'stats' | 'history'

export const MyPageContent: React.FC = () => {
  const { user: authUser, token } = useAuth()
  const router = useRouter()
  const [user, setUser] = useState(authUser)
  const [stats, setStats] = useState<UserStats | null>(null)
  const [favorites, setFavorites] = useState<FavoriteAnime[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<TabType>('favorites')
  const [favoriteAnime, setFavoriteAnime] = useState('')
  const [snsLink, setSnsLink] = useState('')

  const API_BASE_URL =
    process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

  const fetchData = useCallback(async () => {
    setIsLoading(true)
    setError(null)

    try {
      const [userResponse, statsResponse, favoritesResponse] = await Promise.all([
        fetch(`${API_BASE_URL}/api/users/me`, {
          headers: { Authorization: `Bearer ${token}` },
        }),
        fetch(`${API_BASE_URL}/api/users/me/stats`, {
          headers: { Authorization: `Bearer ${token}` },
        }),
        fetch(`${API_BASE_URL}/api/users/me/favorites`, {
          headers: { Authorization: `Bearer ${token}` },
        }),
      ])

      if (userResponse.status === 401 || statsResponse.status === 401 || favoritesResponse.status === 401) {
        router.push('/login')
        return
      }

      // ユーザー情報を取得
      let userData
      if (userResponse.ok) {
        userData = await userResponse.json()
        setUser(userData)
      }

      let statsData: UserStats
      let favoritesData: FavoriteAnime[]

      if (statsResponse.status === 404) {
        statsData = DEFAULT_USER_STATS
      } else if (!statsResponse.ok) {
        throw new Error(ERROR_MESSAGES.STATS_FETCH_FAILED)
      } else {
        statsData = await statsResponse.json()
      }

      if (favoritesResponse.status === 404) {
        favoritesData = []
      } else if (!favoritesResponse.ok) {
        throw new Error(ERROR_MESSAGES.FAVORITES_FETCH_FAILED)
      } else {
        const data = await favoritesResponse.json()
        favoritesData = data.slice(0, 20)
      }

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
    if (authUser && token) {
      fetchData()
    }
  }, [authUser, token, fetchData])

  // ユーザー情報から好きなアニメとSNSリンクを設定
  useEffect(() => {
    if (user) {
      setFavoriteAnime(user.favorite_anime || '')
      setSnsLink(user.sns_link || '')
    }
  }, [user])

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

  const registrationDate = (() => {
    const formatDate = (date: Date) =>
      date.toLocaleDateString('ja-JP', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
      })

    if (!user?.created_at) {
      return formatDate(new Date())
    }

    const date = new Date(user.created_at)
    // 無効な日付またはゼロ値の場合はフォールバック
    if (isNaN(date.getTime()) || date.getFullYear() === 1) {
      return formatDate(new Date())
    }

    return formatDate(date)
  })()

  return (
    <div className="bg-gray-50 min-h-screen">
      <Container className="py-8">
        <div className="max-w-6xl mx-auto">
          <div className="bg-white rounded-lg shadow-sm">
            {/* ヘッダー部分 */}
            <div className="p-6 border-b border-gray-200">
              <div className="flex items-center justify-between mb-4">
                <p className="text-sm text-gray-500">{registrationDate} 登録</p>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => router.push('/mypage/edit')}
                >
                  ⚙️ 設定
                </Button>
              </div>

              {/* プロフィール情報 */}
              <div className="flex items-start space-x-6">
                {/* 左側: アイコン */}
                <div className="flex-shrink-0">
                  <div className="w-32 h-32 bg-gray-100 rounded-full overflow-hidden">
                    {user?.icon ? (
                      <img
                        src={`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}${user.icon}`}
                        alt={user.name}
                        className="w-full h-full object-cover"
                      />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center">
                        <span className="text-gray-400 text-5xl">👤</span>
                      </div>
                    )}
                  </div>
                </div>

                {/* 右側: ユーザー情報とプロフィール詳細 */}
                <div className="flex-1">
                  {/* ユーザー名と統計 */}
                  <div className="mb-6">
                    <h1 className="text-3xl font-bold text-gray-900 mb-3">
                      {user?.name}
                    </h1>

                    {/* 統計バッジ */}
                    <div className="flex items-center space-x-6">
                      <div className="flex items-center space-x-2">
                        <span className="text-xl">📺</span>
                        <div>
                          <p className="text-xs text-gray-500">視聴数</p>
                          <p className="text-base font-bold text-gray-900">
                            {stats?.totalEpisodesWatched || 0}
                          </p>
                        </div>
                      </div>
                      <div className="flex items-center space-x-2">
                        <span className="text-xl">✅</span>
                        <div>
                          <p className="text-xs text-gray-500">完了</p>
                          <p className="text-base font-bold text-gray-900">
                            {stats?.totalAnimesCompleted || 0}
                          </p>
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* プロフィール詳細 */}
                  <div className="space-y-4 max-w-3xl">
                    <div>
                      <h3 className="text-sm font-medium text-gray-500 mb-2">
                        プロフィール
                      </h3>
                      <p className="text-gray-900 whitespace-pre-wrap text-sm leading-relaxed">
                        {user?.description || '未設定'}
                      </p>
                    </div>

                    <div>
                      <h3 className="text-sm font-medium text-gray-500 mb-2">
                        好きなアニメ
                      </h3>
                      <p className="text-gray-900 text-sm leading-relaxed">
                        {favoriteAnime || '未設定'}
                      </p>
                    </div>

                    <div>
                      <h3 className="text-sm font-medium text-gray-500 mb-2">
                        リンク
                      </h3>
                      {snsLink ? (
                        <a
                          href={snsLink}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-blue-600 hover:underline text-sm"
                        >
                          {snsLink}
                        </a>
                      ) : (
                        <p className="text-gray-500 text-sm">未設定</p>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* タブナビゲーション */}
            <div className="border-b border-gray-200">
              <div className="flex space-x-8 px-6">
                <button
                  onClick={() => setActiveTab('favorites')}
                  className={`py-4 px-2 border-b-2 font-medium text-sm transition-colors ${
                    activeTab === 'favorites'
                      ? 'border-blue-600 text-blue-600'
                      : 'border-transparent text-gray-500 hover:text-gray-700'
                  }`}
                >
                  お気に入り
                  <span className="ml-2 text-xs">{favorites.length}</span>
                </button>
                <button
                  onClick={() => setActiveTab('stats')}
                  className={`py-4 px-2 border-b-2 font-medium text-sm transition-colors ${
                    activeTab === 'stats'
                      ? 'border-blue-600 text-blue-600'
                      : 'border-transparent text-gray-500 hover:text-gray-700'
                  }`}
                >
                  統計レポート
                </button>
                <button
                  onClick={() => setActiveTab('history')}
                  className={`py-4 px-2 border-b-2 font-medium text-sm transition-colors ${
                    activeTab === 'history'
                      ? 'border-blue-600 text-blue-600'
                      : 'border-transparent text-gray-500 hover:text-gray-700'
                  }`}
                >
                  視聴履歴
                </button>
              </div>
            </div>

            {/* タブコンテンツ */}
            <div className="p-6">
              {activeTab === 'favorites' && (
                <div>
                  {favorites.length > 0 ? (
                    <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
                      {favorites.map((anime) => {
                        const statusConfig = getAnimeStatusConfig(anime.status)
                        return (
                          <div
                            key={anime.id}
                            className="bg-white rounded-lg border border-gray-200 overflow-hidden hover:shadow-md transition-shadow cursor-pointer"
                            onClick={() => {
                              // アニメ詳細ページは準備中
                              // TODO: router.push(`/anime/${anime.id}`)
                            }}
                          >
                            {anime.imageUrl ? (
                              <img
                                src={anime.imageUrl}
                                alt={anime.title}
                                className="w-full h-40 object-cover"
                              />
                            ) : (
                              <div className="w-full h-40 bg-gray-200 flex items-center justify-center">
                                <span className="text-gray-400 text-3xl">
                                  📺
                                </span>
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
                    <div className="text-center py-12">
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
              )}

              {activeTab === 'stats' && (
                <div className="text-center py-12">
                  <p className="text-gray-600 mb-4">
                    統計レポート機能は準備中です
                  </p>
                  <Button
                    variant="primary"
                    size="md"
                    onClick={() => router.push('/stats')}
                  >
                    詳細統計を見る
                  </Button>
                </div>
              )}

              {activeTab === 'history' && (
                <div className="text-center py-12">
                  <p className="text-gray-600">視聴履歴機能は準備中です</p>
                </div>
              )}
            </div>
          </div>
        </div>
      </Container>
    </div>
  )
}
