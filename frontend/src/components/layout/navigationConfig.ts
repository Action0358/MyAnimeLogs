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

// フッターのサイトマップに並べるリンク。
// ヘッダーナビ（NAV_ITEMS）に加えて、ホームと認証後の導線を含む。
// お気に入りはヘッダーに項目がなく、ここが唯一の導線になる。
export const FOOTER_LINKS: NavItem[] = [
  { label: 'ホーム', href: '/' },
  ...NAV_ITEMS,
  { label: 'お気に入り', href: '/favorites' },
  { label: 'マイページ', href: '/mypage' },
]

// ナビゲーション共通のテキストスタイル
export const NAV_TEXT_STYLES = 'text-sm font-medium'
