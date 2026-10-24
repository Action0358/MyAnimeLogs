import React from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import { ChevronRight } from 'lucide-react'
import { CategoryBadge } from './CategoryBadge'
import { announcementHref, type Announcement } from '@/constants/announcements'

/** YYYY-MM-DD を「2026年9月6日」に整形する。日付は文字列で持っているためパースしない */
export const formatAnnouncementDate = (date: string): string => {
  const [year, month, day] = date.split('-')
  if (!year || !month || !day) return date
  return `${year}年${Number(month)}月${Number(day)}日`
}

/**
 * お知らせ一覧。各行が詳細ページ（/news/[id]）へのリンクになる。
 *
 * 本文は導入として2行で省略する。全文は詳細で読ませるため、一覧は
 * 「日付・カテゴリ → タイトル → 導入」の3段で見渡せる密度に保つ（#1111）。
 * 以前は詳細ページを持たず本文まで一覧に出していたが（#848）、見出しや
 * リストを含む長い告知が増えると一覧が読みにくくなるため見直した。
 */
export const AnnouncementList: React.FC<{ announcements: Announcement[] }> = ({
  announcements,
}) => (
  <ul className="divide-y divide-border overflow-hidden rounded-lg border border-border bg-card">
    {announcements.map((announcement) => (
      <li key={announcement.id}>
        {/* 行全体をリンクにする。リンクの中にリンクは置けないため、本文末尾の
            任意リンク（link）は詳細ページ側に出す */}
        <Link
          href={announcementHref(announcement.id) as Route}
          className="flex items-start gap-3 px-4 py-4 transition-colors hover:bg-muted/60 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring md:gap-4 md:px-6 md:py-[22px]"
        >
          <span className="min-w-0 flex-1">
            <span className="mb-1.5 flex items-center gap-2 md:mb-2 md:gap-2.5">
              <time
                dateTime={announcement.date}
                className="text-[11px] tabular-nums text-muted-foreground md:text-xs"
              >
                {formatAnnouncementDate(announcement.date)}
              </time>
              <CategoryBadge category={announcement.category} size="sm" />
            </span>

            <h3 className="mb-1.5 text-sm font-bold leading-[1.55] text-foreground md:mb-2 md:text-base md:leading-[1.5]">
              {announcement.title}
            </h3>

            {/* text-wrap: pretty は付けない。Safari は段落全体の行長を揃えようとして
                各行を早めに折り返すため、日本語では右側に余白が残り左に寄って見える。
                代わりに auto-phrase で文節単位に折り返し、「お伝えしま／す。」のような
                1文字だけの行を防ぐ（<html lang="ja"> が前提） */}
            <p className="line-clamp-2 whitespace-pre-line [word-break:auto-phrase] text-[12.5px] leading-[1.8] text-foreground/80 md:text-[13.5px] md:leading-[1.85]">
              {announcement.body}
            </p>
          </span>
          <ChevronRight
            aria-hidden="true"
            className="mt-0.5 h-[15px] w-[15px] flex-shrink-0 text-muted-foreground md:mt-[3px] md:h-[17px] md:w-[17px]"
          />
        </Link>
      </li>
    ))}
  </ul>
)

export default AnnouncementList
