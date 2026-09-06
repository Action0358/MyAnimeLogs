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
      {/* 本文の幅を720pxに制限する。一覧ページには（トップページのお知らせ
          セクションと違い）横に並ぶカードグリッドが無いため、読みやすさを
          優先して幅を絞る（モックアップ「News Page Redesign」の指摘）。 */}
      <div className="max-w-[720px]">
        {/* 見出し・件数の体裁は今期アニメ／人気アニメ／検索と揃える */}
        <div className="mb-6 flex items-baseline gap-2 md:gap-3">
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
          <p className="rounded-lg border border-dashed border-border bg-card px-5 py-10 text-center text-sm text-muted-foreground">
            現在お知らせはありません
          </p>
        )}
      </div>
    </Container>
  )
}

export default NewsPage
