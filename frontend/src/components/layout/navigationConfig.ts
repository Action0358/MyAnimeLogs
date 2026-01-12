export interface NavItem {
  label: string
  href: string
}

export const NAV_ITEMS: NavItem[] = [
  {
    label: '今期アニメ',
    href: '/anime/current',
  },
  {
    label: '人気アニメ',
    href: '/anime/popular',
  },
  {
    label: '検索',
    href: '/search',
  },
  {
    label: 'サイト紹介',
    href: '/about',
  },
]

// ナビゲーション共通のテキストスタイル
export const NAV_TEXT_STYLES = 'text-base font-bold'
