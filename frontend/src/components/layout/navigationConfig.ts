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
  { label: 'マイページ', href: '/mypage' },
]

// ナビゲーション共通のテキストスタイル
export const NAV_TEXT_STYLES = 'text-sm font-medium'
