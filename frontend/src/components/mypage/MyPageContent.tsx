'use client'

import React, { useCallback, useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/contexts/AuthContext'
import { Container } from '@/components/layout'
import { Button } from '@/components/ui/Button'
import { getAnimeStatusConfig } from '@/utils/animeStatus'
import { buildIconUrl } from '@/lib/utils'
import {
  DEFAULT_USER_STATS,
  ERROR_MESSAGES,
  type UserStats,
} from '@/constants/stats'
import { get, ApiClientError } from '@/lib/api/client'

interface FavoriteAnime {
  id: number
  title: string
  status: string
  imageUrl?: string
}

type TabType = 'favorites' | 'stats' | 'history'

export const MyPageContent: React.FC = () => {
  const {
    user: authUser,
    token,
    isLoading: authLoading,
    isAuthenticated,
  } = useAuth()
  const router = useRouter()
  const [user, setUser] = useState(authUser)
  const [stats, setStats] = useState<UserStats | null>(null)
  const [favorites, setFavorites] = useState<FavoriteAnime[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<TabType>('favorites')
  const [favoriteAnime, setFavoriteAnime] = useState('')
  const [snsLink, setSnsLink] = useState('')
  const [imageError, setImageError] = useState(false)

  const fetchData = useCallback(async () => {
    setIsLoading(true)
    setError(null)

    try {
      // 並列でデータを取得
      const [userData, statsData, favoritesData] = await Promise.all([
        get('/api/users/me'),
        get<UserStats>('/api/users/me/stats').catch((err) => {
          // 404エラーの場合のみデフォルト値を返す
          if (err instanceof ApiClientError && err.status === 404) {
            return DEFAULT_USER_STATS
          }
          // その他のエラー（401を含む）は再スローして上位で処理
          throw err
        }),
        get<FavoriteAnime[]>('/api/users/me/favorites').catch((err) => {
          // 404エラーの場合のみ空配列を返す
          if (err instanceof ApiClientError && err.status === 404) {
            return []
          }
          // その他のエラー（401を含む）は再スローして上位で処理
          throw err
        }),
      ])

      setUser(userData)
      setStats(statsData)
      setFavorites(favoritesData.slice(0, 20))
    } catch (err) {
      // 401エラーの場合はグローバルハンドラーが認証状態をクリアし、
      // ProtectedRouteがリダイレクトを実行するため、エラー表示せずに処理を終了
      if (err instanceof ApiClientError && err.status === 401) {
        return
      }
      
      // その他のエラーはユーザーに通知
      const errorMessage =
        err instanceof Error ? err.message : ERROR_MESSAGES.DATA_FETCH_FAILED
      setError(errorMessage)
      console.error('Failed to fetch data:', err)
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    // 認証チェックが完了し、認証されている場合のみデータを取得
    if (!authLoading && isAuthenticated && authUser && token) {
      fetchData()
    } else if (!authLoading && !isAuthenticated) {
      // 認証されていない場合はローディングを終了
      setIsLoading(false)
    }
  }, [authUser, token, authLoading, isAuthenticated, fetchData])

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
            <div className="p-4 sm:p-6 border-b border-gray-200">
              <div className="flex items-center justify-between mb-4">
                <p className="text-xs sm:text-sm text-gray-500">{registrationDate} 登録</p>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => router.push('/mypage/edit')}
                  className="text-xs sm:text-sm"
                >
                  ⚙️ 設定
                </Button>
              </div>

              {/* プロフィール情報 */}
              <div className="flex flex-col sm:flex-row items-start sm:space-x-6 space-y-4 sm:space-y-0">
                {/* 左側: アイコン */}
                <div className="flex-shrink-0 mx-auto sm:mx-0">
                  <div className="w-24 h-24 sm:w-32 sm:h-32 bg-gray-100 rounded-full overflow-hidden">
                    {user?.icon && !imageError ? (
                      <img
                        src={buildIconUrl(user.icon) || ''}
                        alt={user.name}
                        className="w-full h-full object-cover"
                        onError={() => setImageError(true)}
                      />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center">
                        <span className="text-gray-400 text-4xl sm:text-5xl">👤</span>
                      </div>
                    )}
                  </div>
                </div>

                {/* 右側: ユーザー情報とプロフィール詳細 */}
                <div className="flex-1 w-full">
                  {/* ユーザー名と統計 */}
                  <div className="mb-4 sm:mb-6">
                    <h1 className="text-2xl sm:text-3xl font-bold text-blue-900 mb-3 text-center sm:text-left">
                      {user?.name}
                    </h1>

                    {/* 統計バッジ */}
                    <div className="flex items-center justify-center sm:justify-start space-x-4 sm:space-x-6">
                      <div className="flex items-center space-x-2">
                        <span className="text-lg sm:text-xl">📺</span>
                        <div>
                          <p className="text-xs text-gray-500">視聴数</p>
                          <p className="text-sm sm:text-base font-bold text-gray-900">
                            {stats?.totalEpisodesWatched || 0}
                          </p>
                        </div>
                      </div>
                      <div className="flex items-center space-x-2">
                        <span className="text-lg sm:text-xl">✅</span>
                        <div>
                          <p className="text-xs text-gray-500">完了</p>
                          <p className="text-sm sm:text-base font-bold text-gray-900">
                            {stats?.totalAnimesCompleted || 0}
                          </p>
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* プロフィール詳細 */}
                  <div className="space-y-3 sm:space-y-4 max-w-3xl">
                    <div className="bg-white border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow">
                      <h3 className="text-base sm:text-lg font-bold text-gray-900 mb-2">
                        プロフィール
                      </h3>
                      <p className="text-gray-900 whitespace-pre-wrap text-sm leading-relaxed">
                        {user?.description || '未設定'}
                      </p>
                    </div>

                    <div className="bg-white border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow">
                      <h3 className="text-base sm:text-lg font-bold text-gray-900 mb-2">
                        好きなアニメ
                      </h3>
                      <p className="text-gray-900 text-sm leading-relaxed">
                        {favoriteAnime || '未設定'}
                      </p>
                    </div>

                    <div className="bg-white border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow">
                      <h3 className="text-base sm:text-lg font-bold text-gray-900 mb-2">
                        SNSリンク
                      </h3>
                      <div>
                        {snsLink ? (
                          <a
                            href={snsLink}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-blue-600 hover:underline text-sm break-all"
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
            </div>

            {/* タブナビゲーション */}
            <div className="border-b border-gray-200 overflow-x-auto">
              <div className="flex space-x-4 sm:space-x-8 px-4 sm:px-6 min-w-max">
                <button
                  onClick={() => setActiveTab('favorites')}
                  className={`py-3 sm:py-4 px-2 border-b-2 font-medium text-xs sm:text-sm transition-colors whitespace-nowrap ${
                    activeTab === 'favorites'
                      ? 'border-blue-600 text-blue-600'
                      : 'border-transparent text-gray-500 hover:text-gray-700'
                  }`}
                >
                  お気に入り
                  <span className="ml-1 sm:ml-2 text-xs">{favorites.length}</span>
                </button>
                <button
                  onClick={() => setActiveTab('stats')}
                  className={`py-3 sm:py-4 px-2 border-b-2 font-medium text-xs sm:text-sm transition-colors whitespace-nowrap ${
                    activeTab === 'stats'
                      ? 'border-blue-600 text-blue-600'
                      : 'border-transparent text-gray-500 hover:text-gray-700'
                  }`}
                >
                  統計レポート
                </button>
                <button
                  onClick={() => setActiveTab('history')}
                  className={`py-3 sm:py-4 px-2 border-b-2 font-medium text-xs sm:text-sm transition-colors whitespace-nowrap ${
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
            <div className="p-4 sm:p-6">
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
