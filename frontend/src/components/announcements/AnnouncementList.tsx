import React from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import { CategoryBadge } from './CategoryBadge'
import type { Announcement } from '@/constants/announcements'

/** YYYY-MM-DD を「2026年9月6日」に整形する。日付は文字列で持っているためパースしない */
export const formatAnnouncementDate = (date: string): string => {
  const [year, month, day] = date.split('-')
  if (!year || !month || !day) return date
  return `${year}年${Number(month)}月${Number(day)}日`
}

/** 外部リンクかどうか。スキーム付きなら別タブで開く */
const isExternal = (href: string) => /^https?:\/\//.test(href)

/**
 * お知らせ一覧。本文までその場に出す。
 *
 * 1件ごとの詳細ページは持たない。お知らせは短文中心で、
 * 遷移させるより読み切れたほうが速いため（#848）。
 */
export const AnnouncementList: React.FC<{ announcements: Announcement[] }> = ({
  announcements,
}) => (
  <ul className="divide-y divide-border border-t border-border">
    {announcements.map((announcement) => (
      <li key={announcement.id} className="py-6 first:pt-5">
        <div className="mb-2 flex flex-wrap items-center gap-x-3 gap-y-2">
          <time dateTime={announcement.date} className="text-sm text-muted-foreground">
            {formatAnnouncementDate(announcement.date)}
          </time>
          <CategoryBadge category={announcement.category} />
        </div>

        <h3 className="mb-2 text-base font-bold text-foreground md:text-lg">
          {announcement.title}
        </h3>

        {/* 本文の改行をそのまま反映する。Markdownは導入していない */}
        <p className="whitespace-pre-wrap text-sm leading-[1.75] text-foreground/80">
          {announcement.body}
        </p>

        {announcement.link && (
          <Link
            href={announcement.link.href as Route}
            className="mt-3 inline-block text-sm font-medium text-primary hover:underline"
            {...(isExternal(announcement.link.href)
              ? { target: '_blank', rel: 'noopener noreferrer' }
              : {})}
          >
            {announcement.link.label}
          </Link>
        )}
      </li>
    ))}
  </ul>
)

export default AnnouncementList
