import { WATCH_STATUS_TAB_LABELS, type WatchStatusTabKey } from '@/components/favorites'

const DEFAULT_TAB: WatchStatusTabKey = 'planned'

const isWatchStatusTabKey = (value: string | null): value is WatchStatusTabKey =>
  value !== null && value in WATCH_STATUS_TAB_LABELS

/**
 * マイページで表示中のタブはURLの ?tab= に持たせる。作品詳細の「マイページに戻る」・
 * 設定画面のパンくず・ブラウザの戻るで、遷移前に開いていたタブへ戻ってこられるようにするため。
 *
 * マイページと設定画面は ProtectedRoute の認証確認後にクライアントでのみ描画されるため、
 * 初期値を window から読んでもハイドレーション不整合は起きない。
 * useSearchParams は静的エクスポート時に Suspense 境界を要求するので使わない。
 */
export const readMyPageTabFromUrl = (): WatchStatusTabKey => {
  if (typeof window === 'undefined') return DEFAULT_TAB
  const tab = new URLSearchParams(window.location.search).get('tab')
  return isWatchStatusTabKey(tab) ? tab : DEFAULT_TAB
}

/** タブ付きのマイページ／設定画面のURL */
export const myPageHref = (tab: WatchStatusTabKey): string => `/mypage?tab=${tab}`
export const myPageSettingsHref = (tab: WatchStatusTabKey): string => `/mypage/edit?tab=${tab}`
