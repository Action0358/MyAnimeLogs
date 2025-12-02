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
    label: '検索',
    href: '/search',
  },
  {
    label: 'サイト紹介',
    href: '/about',
  },
]

export const scrollToCurrentSeason = () => {
  const element = document.querySelector('#current-season')
  if (element) {
    element.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
}
