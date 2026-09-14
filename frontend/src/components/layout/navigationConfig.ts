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

// フッターのサイトマップに並べるリンク。
// お気に入りはヘッダーに項目がなく、ここが唯一の導線になる。
//
// 作品を見つける導線（今期アニメ・人気アニメ・お気に入り）を隣接させたいため、
// NAV_ITEMS の順序をそのまま流用せず独立して定義している。
// ページを追加・改称したときは NAV_ITEMS と併せて更新すること。
export const FOOTER_LINKS: NavItem[] = [
  { label: 'ホーム', href: '/' },
  { label: '今期アニメ', href: '/anime/current' },
  { label: '人気アニメ', href: '/anime/popular' },
  { label: 'お気に入り', href: '/favorites' },
  { label: '検索', href: '/search' },
  { label: 'サイト紹介', href: '/about' },
  { label: 'お知らせ', href: '/news' },
  { label: 'マイページ', href: '/mypage' },
  { label: 'お問い合わせ', href: '/contact' },
]

// ナビゲーション共通のテキストスタイル
export const NAV_TEXT_STYLES = 'text-sm font-medium'
