import type React from 'react'
import type { Metadata } from 'next'
import { Container } from '@/components/layout'
import { AnnouncementList } from '@/components/announcements'
import { getAnnouncements } from '@/constants/announcements'

export const metadata: Metadata = {
  title: 'お知らせ | MyAnimeLogs',
  description: 'MyAnimeLogs の運営からのお知らせ一覧です。',
}

/**
 * お知らせ一覧。1件ごとの詳細ページは持たず、ここに本文まで出す（#848）。
 *
 * データはリポジトリの定数なのでビルド時に確定する。API を叩かないため
 * クライアントコンポーネントにする必要がなく、そのまま静的生成される。
 */
const NewsPage: React.FC = () => {
  const announcements = getAnnouncements()

  return (
    <Container size="xl" className="max-w-[1180px] py-8">
      {/* 見出し・件数の体裁は今期アニメ／人気アニメ／検索と揃える */}
      <div className="mb-5 flex items-baseline gap-2 md:gap-3">
        <h1 className="text-xl md:text-[26px] font-bold text-foreground tracking-tight">
          お知らせ
        </h1>
        <span className="flex flex-shrink-0 items-baseline gap-1 text-sm text-muted-foreground">
          <span className="text-base font-bold text-foreground">{announcements.length}</span>件
        </span>
      </div>

      {announcements.length > 0 ? (
        <AnnouncementList announcements={announcements} />
      ) : (
        <p className="border-t border-border py-12 text-center text-sm text-muted-foreground">
          現在お知らせはありません。
        </p>
      )}
    </Container>
  )
}

export default NewsPage
