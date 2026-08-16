'use client'

import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import type { Route } from 'next'
import { Settings } from 'lucide-react'
import { useAuth } from '@/contexts/AuthContext'
import { Container } from '@/components/layout'
import { Button } from '@/components/ui/Button'
import {
  WatchStatusTabs,
  WatchStatusCard,
  WatchStatusTabKey,
  WATCH_STATUS_TABS,
  WATCH_STATUS_TAB_LABELS,
} from '@/components/favorites'
import { useWatchStatus, WatchStatusAnime } from '@/hooks/useWatchStatus'
import { WatchStatusStatus } from '@/lib/api/generated/myanimelogs'
import { resolveAvatarUrl } from '@/lib/avatars'
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

  // ユーザー情報から好きなアニメを設定
  useEffect(() => {
    if (user) {
      setFavoriteAnime(user.favorite_anime || '')
    }
  }, [user])

  const allItems = useMemo(() => filterValidItems(watchStatuses), [watchStatuses])

  // ステータスごとの件数。記録量の指標として見せるため、
  // 表示件数の上限ではなく全件から数える。
  const counts = useMemo(() => {
    const result: Record<WatchStatusTabKey, number> = {
      watching: 0,
      planned: 0,
      completed: 0,
    }
    allItems.forEach((anime) => {
      const status = anime.user_watch_status?.status
      if (status && status in result) {
        result[status as WatchStatusStatus]++
      }
    })
    return result
  }, [allItems])

  // グリッドの表示は上限を維持する
  const filteredItems = useMemo(() => {
    return allItems
      .filter((anime) => anime.user_watch_status?.status === activeTab)
      .slice(0, MAX_WATCH_STATUSES)
  }, [allItems, activeTab])

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

  // 未設定の項目は空欄を並べず、何を追加できるかを示して編集へ誘導する
  const profileFields = [
    {
      label: 'プロフィール',
      value: user?.description || '',
      emptyPrompt: '自己紹介を追加',
    },
    {
      label: '好きなアニメ',
      value: favoriteAnime,
      emptyPrompt: '作品を追加',
    },
  ]

  const avatar = (
    <div className="w-16 h-16 md:w-[88px] md:h-[88px] bg-muted rounded-full overflow-hidden flex-shrink-0">
      {user?.icon && !imageError ? (
        // eslint-disable-next-line @next/next/no-img-element
        <img
          src={resolveAvatarUrl(user.icon) || ''}
          alt={user?.name || 'ユーザーアイコン'}
          className="w-full h-full object-cover"
          onError={() => setImageError(true)}
        />
      ) : (
        <div className="w-full h-full flex items-center justify-center">
          <span className="text-muted-foreground text-3xl md:text-4xl">👤</span>
        </div>
      )}
    </div>
  )

  const settingsButton = (
    <button
      type="button"
      onClick={() => router.push('/mypage/edit')}
      className="flex items-center justify-center gap-1.5 h-9 w-9 md:h-[38px] md:w-auto md:px-3.5 flex-shrink-0 border border-border rounded-md bg-card text-sm font-semibold text-foreground hover:bg-muted transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
      aria-label="プロフィール設定"
    >
      <Settings className="w-[15px] h-[15px] flex-shrink-0" aria-hidden="true" />
      <span className="hidden md:inline">プロフィール設定</span>
    </button>
  )

  const stats = WATCH_STATUS_TABS.map((tab) => ({
    key: tab.key,
    label: tab.label,
    value: counts[tab.key],
  }))

  return (
    <Container size="xl" className="max-w-[1180px] py-8">
      {/* プロフィールヘッダー。アバター・名前・記録数・設定を1つの行にまとめる */}
      <div className="flex items-center md:items-start gap-4 md:gap-6 pb-5 md:pb-[26px] border-b border-border">
        {avatar}
        <div className="flex-1 min-w-0">
          <div className="flex flex-col md:flex-row md:items-center md:gap-3">
            <h1 className="text-lg md:text-2xl font-bold text-foreground truncate">
              {user?.name}
            </h1>
            <p className="text-[11.5px] md:text-[12.5px] text-muted-foreground mt-1 md:mt-0 flex-shrink-0">
              {registrationDate} 登録
            </p>
          </div>

          {/* 記録数。狭い画面では下のカードに出すため隠す */}
          <div className="hidden md:flex gap-7 mt-4">
            {stats.map((stat) => (
              <div key={stat.key}>
                <div className="text-xl font-bold leading-tight tabular-nums text-foreground">
                  {stat.value}
                </div>
                <div className="text-xs text-muted-foreground mt-[3px]">{stat.label}</div>
              </div>
            ))}
          </div>
        </div>
        {settingsButton}
      </div>

      {/* 記録数（狭い画面）。1行に収めてスクロール量を抑える */}
      <div className="grid grid-cols-3 gap-2 mt-4 md:hidden">
        {stats.map((stat) => (
          <div
            key={stat.key}
            className="border border-border rounded-lg py-3 px-2.5 text-center"
          >
            <div className="text-lg font-bold leading-tight tabular-nums text-foreground">
              {stat.value}
            </div>
            <div className="text-[11px] text-muted-foreground mt-[3px]">{stat.label}</div>
          </div>
        ))}
      </div>

      {/* プロフィール項目。未設定は編集への導線にする */}
      <dl className="grid grid-cols-[88px_1fr] md:grid-cols-[104px_1fr] gap-x-4 md:gap-x-5 gap-y-3 md:gap-y-3.5 py-5 md:py-[22px] mt-4 md:mt-0 border-b border-border text-[12.5px] md:text-[13.5px]">
        {profileFields.map((field) => (
          <React.Fragment key={field.label}>
            <dt className="text-muted-foreground">{field.label}</dt>
            <dd className="min-w-0">
              {field.value ? (
                <span className="text-foreground whitespace-pre-wrap break-words">
                  {field.value}
                </span>
              ) : (
                <Link
                  href={'/mypage/edit' as Route}
                  className="text-primary font-semibold hover:underline"
                >
                  {field.emptyPrompt}
                </Link>
              )}
            </dd>
          </React.Fragment>
        ))}
      </dl>

      {/* 視聴ステータス別タブナビゲーション */}
      <div className="mt-5 md:mt-[26px] mb-4 md:mb-6">
        <WatchStatusTabs
          activeTab={activeTab}
          onTabChange={setActiveTab}
          counts={counts}
        />
      </div>

      {/* タブコンテンツ */}
      {filteredItems.length > 0 ? (
        <div className="grid grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-x-2.5 gap-y-3.5 md:gap-x-[18px] md:gap-y-5">
          {filteredItems.map((anime) => (
            <WatchStatusCard key={anime.id} anime={anime} />
          ))}
        </div>
      ) : (
        <div className="text-center py-12">
          <div className="w-16 h-16 bg-muted rounded-full flex items-center justify-center mx-auto mb-4">
            <span className="text-muted-foreground text-3xl">📺</span>
          </div>
          <p className="text-muted-foreground mb-4">
            {`「${WATCH_STATUS_TAB_LABELS[activeTab]}」のアニメがありません`}
          </p>
          <Button variant="primary" size="md" onClick={() => router.push('/')}>
            アニメを探す
          </Button>
        </div>
      )}
    </Container>
  )
}
