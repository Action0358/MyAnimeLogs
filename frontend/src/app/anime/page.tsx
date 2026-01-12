import { redirect } from 'next/navigation'

// 旧アニメ一覧ページ → 今期アニメページへサーバーサイドリダイレクト
// SEO対応のため、サーバーサイドでリダイレクト
export default function AnimeListPage() {
  redirect('/anime/current')
}
