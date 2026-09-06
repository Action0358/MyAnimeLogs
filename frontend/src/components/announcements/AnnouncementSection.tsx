import React from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import { Button, SectionHeading } from '@/components/ui'
import { CategoryBadge } from './CategoryBadge'
import { formatAnnouncementDate } from './AnnouncementList'
import { getAnnouncements, HOME_ANNOUNCEMENT_COUNT } from '@/constants/announcements'

/**
 * トップページのお知らせセクション。最新数件をタイトルだけのリストで出す。
 *
 * アニメカードが主役のページなので、本文は出さず一覧ページへ送る。
 * お知らせが1件もないときはセクションごと描画しない。
 */
export const AnnouncementSection: React.FC<{ className?: string }> = ({ className }) => {
  const announcements = getAnnouncements().slice(0, HOME_ANNOUNCEMENT_COUNT)

  if (announcements.length === 0) return null

  return (
    <section className={className} aria-labelledby="home-announcements">
      <SectionHeading
        id="home-announcements"
        action={
          <Button variant="outline-gray" size="sm" asChild className="font-bold">
            <Link href={'/news' as Route}>もっと見る</Link>
          </Button>
        }
      >
        お知らせ
      </SectionHeading>

      <ul className="divide-y divide-border border-y border-border">
        {announcements.map((announcement) => (
          <li key={announcement.id}>
            <Link
              href={'/news' as Route}
              className="flex flex-col gap-1.5 py-3.5 transition-colors hover:bg-muted/50 sm:flex-row sm:items-center sm:gap-3"
            >
              <span className="flex flex-shrink-0 items-center gap-3">
                <time
                  dateTime={announcement.date}
                  className="text-sm tabular-nums text-muted-foreground"
                >
                  {formatAnnouncementDate(announcement.date)}
                </time>
                <CategoryBadge category={announcement.category} />
              </span>
              {/* 長いタイトルでも行が増えないよう1行に省略する */}
              <span className="truncate text-sm font-medium text-foreground">
                {announcement.title}
              </span>
            </Link>
          </li>
        ))}
      </ul>
    </section>
  )
}

export default AnnouncementSection
