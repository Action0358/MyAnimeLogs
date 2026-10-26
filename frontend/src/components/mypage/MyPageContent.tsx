'use client'

import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import type { Route } from 'next'
import { Settings, Bookmark, CirclePlay, CircleCheck } from 'lucide-react'
import { useAuth } from '@/contexts/AuthContext'
import { Container } from '@/components/layout'
import { Button } from '@/components/ui/Button'
import { EmptyState } from '@/components/ui/EmptyState'
import { WatchStatusTabs, WatchStatusCard, WatchStatusTabKey } from '@/components/favorites'
import { ErrorState, ExpandToggleButton, getErrorStateType } from '@/components/ui'
import { useWatchStatus, WatchStatusAnime } from '@/hooks/useWatchStatus'
import { useExpandableList } from '@/hooks/useExpandableList'
import { useResponsiveInitialCount } from '@/hooks/useResponsiveInitialCount'
import { WatchStatusStatus } from '@/lib/api/generated/myanimelogs'
import { get, ApiClientError } from '@/lib/api/client'
import { readMyPageTabFromUrl, myPageHref, myPageSettingsHref } from './myPageTab'

// マイページの各タブで最初に表示する視聴ステータスの件数（超えた分は展開ボタンで表示する）。
// グリッドはSP3列・PC6列のため、SPは詰まりすぎないよう少なめにする
const WATCH_STATUS_INITIAL_COUNT_SP = 18
const WATCH_STATUS_INITIAL_COUNT_PC = 24

// タブごとの空状態。アイコンはステータスボタンと同じものを使い、
// 次の行動を案内する文言にする（空状態UI刷新）
const EMPTY_STATE_CONFIG: Record<
  WatchStatusTabKey,
  { icon: React.ReactNode; title: string; description: string; ctaLabel: string; ctaHref: string }
> = {
  planned: {
    icon: <Bookmark className="w-7 h-7 md:w-8 md:h-8" strokeWidth={1.75} aria-hidden="true" />,
    title: '視聴予定のアニメはありません',
    description: '気になる作品を視聴予定に追加しましょう',
    ctaLabel: 'アニメを探す',
    ctaHref: '/search',
  },
  watching: {
    icon: <CirclePlay className="w-7 h-7 md:w-8 md:h-8" strokeWidth={1.75} aria-hidden="true" />,
    title: '視聴中のアニメはありません',
    description: '見ている作品を視聴中に追加しましょう',
    // 他タブ（視聴予定・視聴済）とCTAの文言・遷移先を統一する
    ctaLabel: 'アニメを探す',
    ctaHref: '/search',
  },
  completed: {
    icon: <CircleCheck className="w-7 h-7 md:w-8 md:h-8" strokeWidth={1.75} aria-hidden="true" />,
    title: '視聴済のアニメはありません',
    description: '見終えた作品を視聴済に追加しましょう',
    ctaLabel: 'アニメを探す',
    ctaHref: '/search',
  },
}

// id が number の作品のみを抽出（WatchStatusCard は id 必須）
const filterValidItems = (items: WatchStatusAnime[]): (WatchStatusAnime & { id: number })[] => {
  return items.filter(
    (anime): anime is WatchStatusAnime & { id: number } => typeof anime.id === 'number'
  )
}

export const MyPageContent: React.FC = () => {
  const { user: authUser, token, isLoading: authLoading, isAuthenticated } = useAuth()
  const router = useRouter()
  const [user, setUser] = useState(authUser)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<unknown>(null)
  const [activeTab, setActiveTab] = useState<WatchStatusTabKey>(readMyPageTabFromUrl)
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

      // その他のエラーはユーザーに通知（画面には生メッセージを出さず、種類だけで表示を決める）
      setError(err ?? new Error('データの取得に失敗しました'))
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

  const changeTab = (tab: WatchStatusTabKey) => {
    setActiveTab(tab)
    // 履歴を積まずにURLだけ更新する。タブ切り替えのたびに戻る操作が増えないように
    router.replace(myPageHref(tab) as Route, { scroll: false })
    // タブを切り替えたら、前のタブで展開していた状態を引き継がない
    resetExpanded()
  }

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

  const tabItems = useMemo(() => {
    return allItems.filter((anime) => anime.user_watch_status?.status === activeTab)
  }, [allItems, activeTab])

  // 件数が多いタブは初期表示を絞り、ボタンで残りを表示する（関連作品と同じパターン）
  const { initialCount, oneRowAtWidest } = useResponsiveInitialCount(
    WATCH_STATUS_INITIAL_COUNT_SP,
    WATCH_STATUS_INITIAL_COUNT_PC
  )
  const {
    visible: filteredItems,
    hiddenCount,
    isExpanded,
    toggle: toggleExpanded,
    reset: resetExpanded,
  } = useExpandableList(tabItems, initialCount, oneRowAtWidest)

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
        <h1 className="text-xl md:text-2xl font-bold text-foreground">マイページ</h1>
        <ErrorState
          type={getErrorStateType(error || watchStatusesError)}
          onRetry={error ? fetchData : () => window.location.reload()}
        />
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

  const avatar = (
    <div className="w-16 h-16 md:w-[88px] md:h-[88px] bg-muted rounded-full overflow-hidden flex-shrink-0">
      {user?.avatar_url && !imageError ? (
        // eslint-disable-next-line @next/next/no-img-element
        <img
          src={user.avatar_url}
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
      // 設定画面のパンくずから、いま開いているタブへ戻れるようにタブを渡す
      onClick={() => router.push(myPageSettingsHref(activeTab) as Route)}
      className="flex items-center justify-center gap-1.5 h-9 w-9 md:h-[38px] md:w-auto md:px-3.5 flex-shrink-0 border border-border rounded-md bg-card text-sm font-semibold text-foreground hover:bg-muted transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
      aria-label="設定"
    >
      <Settings className="w-[15px] h-[15px] flex-shrink-0" aria-hidden="true" />
      <span className="hidden md:inline">設定</span>
    </button>
  )

  return (
    <Container size="xl" className="py-8">
      {/* ヘッダー。アバター・名前・登録日・設定を1つの行にまとめる。
          下線はタブ側にのみ引く（マイページUI刷新v2）。
          表示名とアイコンはGoogleアカウント由来のため編集導線は持たせない（#789） */}
      <div className="flex items-center gap-4 md:gap-5">
        {avatar}
        <div className="flex-1 min-w-0">
          <h1 className="text-lg md:text-2xl font-bold text-foreground truncate leading-tight">
            {user?.name}
          </h1>
          <p className="text-[11.5px] md:text-[12.5px] text-muted-foreground mt-1">
            {registrationDate} 登録
          </p>
        </div>
        {settingsButton}
      </div>

      {/* 視聴ステータス別タブナビゲーション。件数はタブ内のピルに統合し、
          記録数カードとの二重表示をなくす（マイページUI刷新v2） */}
      <div className="mt-[22px] mb-[18px] md:mt-[30px] md:mb-7">
        <WatchStatusTabs activeTab={activeTab} onTabChange={changeTab} counts={counts} />
      </div>

      {/* タブコンテンツ */}
      {filteredItems.length > 0 ? (
        <>
          <div className="grid grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-x-2.5 gap-y-3.5 md:gap-x-[18px] md:gap-y-5">
            {filteredItems.map((anime) => (
              <WatchStatusCard key={anime.id} anime={anime} from={myPageHref(activeTab)} />
            ))}
          </div>
          {hiddenCount > 0 && (
            <ExpandToggleButton
              isExpanded={isExpanded}
              hiddenCount={hiddenCount}
              onClick={toggleExpanded}
            />
          )}
        </>
      ) : (
        <div className="md:min-h-[420px] md:flex md:items-center md:justify-center">
          <EmptyState
            icon={EMPTY_STATE_CONFIG[activeTab].icon}
            title={EMPTY_STATE_CONFIG[activeTab].title}
            description={EMPTY_STATE_CONFIG[activeTab].description}
            action={{
              label: EMPTY_STATE_CONFIG[activeTab].ctaLabel,
              onClick: () => router.push(EMPTY_STATE_CONFIG[activeTab].ctaHref as Route),
            }}
          />
        </div>
      )}
    </Container>
  )
}
