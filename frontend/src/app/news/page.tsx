import type React from 'react'
import type { Metadata } from 'next'
import { Container } from '@/components/layout'
import { AnnouncementBrowser } from '@/components/announcements'
import { getAnnouncements } from '@/constants/announcements'

export const metadata: Metadata = {
  title: 'お知らせ | MyAnimeLogs',
  description: 'MyAnimeLogs の運営からのお知らせ一覧です。',
}

/**
 * お知らせ一覧。1件ごとの詳細ページは持たず、ここに本文まで出す（#848）。
 *
 * データはリポジトリの定数なのでビルド時に確定し、そのまま静的生成される。
 * 絞り込み・ページ送りの操作だけを AnnouncementBrowser（クライアント）に任せる。
 */
const NewsPage: React.FC = () => (
  <Container size="xl" className="max-w-[1180px] py-8">
    {/* 本文はコンテナ中央の読み幅列（760px）に置く。左寄せだと右に大きな空白が
        でき、1件だけのときに特に偏って見える（モックアップ「News Page Redesign」
        修正版）。一覧ページには横に並ぶカードグリッドが無いため、全幅にはせず
        読みやすさを優先して幅を絞る */}
    <div className="mx-auto max-w-[760px]">
      <AnnouncementBrowser announcements={getAnnouncements()} />
    </div>
  </Container>
)

export default NewsPage
