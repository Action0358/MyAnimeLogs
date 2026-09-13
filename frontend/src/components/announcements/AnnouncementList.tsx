import React from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import { CategoryBadge } from './CategoryBadge'
import { cn } from '@/lib/utils'
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
 * 代わりに各項目へ id を振り、/news#<id> で個別に指せるようにしている。
 *
 * モックアップ（News Page Redesign）は行全体をクリック可能にしホバー＋
 * シェブロンで反応させているが、これは各項目に遷移先（詳細ページ）が
 * ある一般的なお知らせページの前提であり、詳細ページを持たないこの
 * プロジェクトでは行を押しても何も起きない「空振り」になってしまう。
 * 枠・階層構造（メタ行→タイトル→本文）などの視覚デザインは採用しつつ、
 * 行全体のリンク化・ホバー・シェブロンは見送った。`link` を指定した項目は
 * 従来どおり本文の下に明示的なテキストリンクとして出す。
 */
export const AnnouncementList: React.FC<{ announcements: Announcement[] }> = ({
  announcements,
}) => (
  <ul className="divide-y divide-border overflow-hidden rounded-lg border border-border bg-card">
    {announcements.map((announcement) => (
      <li
        key={announcement.id}
        // /news#<id> で個別のお知らせを直接指せるようにする。
        // 詳細ページを持たない代わりの導線で、トップのお知らせ行や
        // 外部からの共有リンクがここへ着地する。
        id={announcement.id}
        className={cn(
          'px-4 py-4 md:px-6 md:py-[22px]',
          // ヘッダーは固定ではないが、上端ぴったりに着地すると窮屈に見えるため少し余白を残す
          'scroll-mt-6',
          // アンカーで来たときに「どれが指された項目か」が分かるようにする。
          // 枠線ではなく背景にするのは、行の高さが変わって他が動かないようにするため
          'target:bg-primary/5',
        )}
      >
        <div className="mb-2 flex items-center gap-2.5">
          <time
            dateTime={announcement.date}
            className="text-xs tabular-nums text-muted-foreground"
          >
            {formatAnnouncementDate(announcement.date)}
          </time>
          <CategoryBadge category={announcement.category} size="sm" />
        </div>

        <h3 className="mb-2 text-sm font-bold leading-[1.5] text-foreground md:text-base">
          {announcement.title}
        </h3>

        {/* 本文の改行をそのまま反映する。Markdownは導入していない */}
        <p className="whitespace-pre-wrap text-xs leading-[1.8] text-foreground/80 md:text-sm">
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
