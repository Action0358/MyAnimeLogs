'use client'

import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/contexts/AuthContext'
import { Container } from '@/components/layout'
import { Button } from '@/components/ui/Button'
import {
  WatchStatusTabs,
  WatchStatusCard,
  WatchStatusTabKey,
  WATCH_STATUS_TAB_LABELS,
} from '@/components/favorites'
import { useWatchStatus, WatchStatusAnime } from '@/hooks/useWatchStatus'
import { WatchStatusStatus } from '@/lib/api/generated/myanimelogs'
import { buildIconUrl } from '@/lib/utils'
import { getProfileFieldValues } from '@/lib/profile'
import { get, ApiClientError } from '@/lib/api/client'

// マイページに表示する視聴ステータスの最大件数
const MAX_WATCH_STATUSES = 20

// id が number の作品のみを抽出（WatchStatusCard は id 必須）
const filterValidItems = (
  items: WatchStatusAnime[]
): (WatchStatusAnime & { id: number })[] => {
  return items.filter(
    (anime): anime is WatchStatusAnime & { id: number } =>
      typeof anime.id === 'number'
  )
}

export const MyPageContent: React.FC = () => {
  const {
    user: authUser,
    token,
    isLoading: authLoading,
    isAuthenticated,
  } = useAuth()
  const router = useRouter()
  const [user, setUser] = useState(authUser)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<WatchStatusTabKey>('planned')
  const [favoriteAnime, setFavoriteAnime] = useState('')
  const [snsLink, setSnsLink] = useState('')
  const [imageError, setImageError] = useState(false)

  // 視聴ステータス一覧（お気に入りとは独立）
  const {
    watchStatuses,
    isLoading: watchStatusesLoading,
    error: watchStatusesError,
  } = useWatchStatus()

  const fetchData = useCallback(async () => {
    setIsLoading(true)
    setError(null)

    try {
      const userData = await get('/api/users/me')
      setUser(userData)
    } catch (err) {
      // 401エラーの場合はグローバルハンドラーが認証状態をクリアし、
      // ProtectedRouteがリダイレクトを実行するため、エラー表示せずに処理を終了
      if (err instanceof ApiClientError && err.status === 401) {
        return
      }

      // その他のエラーはユーザーに通知
      const errorMessage =
        err instanceof Error ? err.message : 'データの取得に失敗しました'
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
      const { favoriteAnime, snsLink } = getProfileFieldValues(user)
      setFavoriteAnime(favoriteAnime)
      setSnsLink(snsLink)
    }
  }, [user])

  // 件数上限を維持したうえで有効な作品を抽出
  const validItems = useMemo(
    () => filterValidItems(watchStatuses).slice(0, MAX_WATCH_STATUSES),
    [watchStatuses]
  )

  // ステータスタブごとの件数
  const counts = useMemo(() => {
    const result: Record<WatchStatusTabKey, number> = {
      watching: 0,
      planned: 0,
      completed: 0,
    }
    validItems.forEach((anime) => {
      const status = anime.user_watch_status?.status
      if (status && status in result) {
        result[status as WatchStatusStatus]++
      }
    })
    return result
  }, [validItems])

  const filteredItems = useMemo(() => {
    return validItems.filter(
      (anime) => anime.user_watch_status?.status === activeTab
    )
  }, [validItems, activeTab])

  if (isLoading || watchStatusesLoading) {
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

  if (error || watchStatusesError) {
    return (
      <Container className="py-12">
        <div className="bg-white rounded-lg border border-gray-200 p-12 shadow-sm text-center">
          <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <span className="text-red-600 text-3xl">⚠️</span>
          </div>
          <p className="text-red-600 mb-4">
            {error || watchStatusesError?.message || 'データの取得に失敗しました'}
          </p>
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
                  {/* ユーザー名 */}
                  <div className="mb-4 sm:mb-6">
                    <h1 className="text-2xl sm:text-3xl font-bold text-blue-900 mb-3 text-center sm:text-left">
                      {user?.name}
                    </h1>
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

            {/* 視聴ステータス別タブナビゲーション */}
            <div className="px-4 sm:px-6">
              <WatchStatusTabs
                activeTab={activeTab}
                onTabChange={setActiveTab}
                counts={counts}
              />
            </div>

            {/* タブコンテンツ */}
            <div className="p-4 sm:p-6">
              {filteredItems.length > 0 ? (
                <div className="grid grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-2 sm:gap-4">
                  {filteredItems.map((anime) => (
                    <WatchStatusCard
                      key={anime.id}
                      anime={anime}
                      status={anime.user_watch_status?.status || 'planned'}
                    />
                  ))}
                </div>
              ) : (
                <div className="text-center py-12">
                  <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
                    <span className="text-gray-400 text-3xl">📺</span>
                  </div>
                  <p className="text-gray-600 mb-4">
                    {`「${WATCH_STATUS_TAB_LABELS[activeTab]}」のアニメがありません`}
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
        </div>
      </Container>
    </div>
  )
}
