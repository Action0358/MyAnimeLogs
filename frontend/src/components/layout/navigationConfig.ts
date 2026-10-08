import type { LucideIcon } from 'lucide-react'
import { CalendarDays, Heart, Home, Info, Search, Star, UserRound } from 'lucide-react'

export interface NavItem {
  label: string
  href: string
  /**
   * モバイルドロワーの行頭に出すアイコン。
   * テキストだけの羅列だと走査しづらいため、意味を一目で掴めるようにする。
   * デスクトップのピルタブとフッターでは使わない。
   */
  icon?: LucideIcon
}

// ドロワー先頭のホーム。デスクトップはロゴがその役目を果たすため NAV_ITEMS には入れない
export const HOME_NAV_ITEM: NavItem = { label: 'ホーム', href: '/', icon: Home }

export const NAV_ITEMS: NavItem[] = [
  {
    label: '今期アニメ',
    href: '/anime/current',
    icon: CalendarDays,
  },
  {
    label: '人気アニメ',
    href: '/anime/popular',
    icon: Star,
  },
  {
    label: '検索',
    href: '/search',
    icon: Search,
  },
  {
    label: 'サイト紹介',
    href: '/about',
    icon: Info,
  },
]

// ログイン中のユーザー向けリンク。
// デスクトップはヘッダーのユーザーメニュー、狭い画面はハンバーガーメニューで使う。
export const USER_MENU_ITEMS: NavItem[] = [
  { label: 'マイページ', href: '/mypage', icon: UserRound },
  { label: 'お気に入り', href: '/favorites', icon: Heart },
]

// ナビゲーション共通のテキストスタイル
export const NAV_TEXT_STYLES = 'text-sm font-medium'
