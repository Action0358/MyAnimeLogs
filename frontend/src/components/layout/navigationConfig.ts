export interface NavItem {
  label: string
  href: string
  requiresScroll?: boolean
}

export const NAV_ITEMS: NavItem[] = [
  {
    label: '今期アニメ',
    href: '/#current-season',
    requiresScroll: true,
  },
  {
    label: 'アニメ一覧',
    href: '/anime',
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

export const scrollToCurrentSeason = () => {
  const element = document.querySelector('#current-season')
  if (element) {
    element.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
}
