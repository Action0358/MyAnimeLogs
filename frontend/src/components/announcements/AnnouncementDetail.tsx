import React from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import { ChevronLeft } from 'lucide-react'
import { CategoryBadge } from './CategoryBadge'
import { formatAnnouncementDate } from './AnnouncementList'
import type { Announcement, AnnouncementBlock, AnnouncementInline } from '@/constants/announcements'

/** 外部リンクかどうか。スキーム付きなら別タブで開く */
const isExternal = (href: string) => /^https?:\/\//.test(href)

const externalProps = (href: string) =>
  isExternal(href) ? { target: '_blank', rel: 'noopener noreferrer' } : {}

const Inline: React.FC<{ nodes: AnnouncementInline[] }> = ({ nodes }) => (
  <>
    {nodes.map((node, i) =>
      typeof node === 'string' ? (
        <React.Fragment key={i}>{node}</React.Fragment>
      ) : (
        <Link
          key={i}
          href={node.href as Route}
          className="text-primary underline underline-offset-[3px] hover:opacity-80"
          {...externalProps(node.href)}
        >
          {node.text}
        </Link>
      )
    )}
  </>
)

const Block: React.FC<{ block: AnnouncementBlock }> = ({ block }) => {
  switch (block.type) {
    case 'p':
      return (
        <p className="[text-wrap:pretty]">
          <Inline nodes={block.children} />
        </p>
      )
    case 'heading':
      // 親の gap に加えて見出しの前を広げ、直後は詰めて本文と結びつける
      return (
        <h2 className="-mb-1 mt-2 text-[16.5px] font-bold text-foreground md:mt-3 md:text-lg">
          {block.text}
        </h2>
      )
    case 'ul':
      return (
        <ul className="list-disc space-y-1 pl-5 md:space-y-1.5">
          {block.items.map((item, i) => (
            <li key={i} className="pl-0.5 [text-wrap:pretty] md:pl-1">
              <Inline nodes={item} />
            </li>
          ))}
        </ul>
      )
  }
}

/**
 * お知らせの詳細。一覧と同じ760px読み幅列・同じメタ行で、遷移しても位置がずれない。
 * 一覧は行の集合なのでカード、詳細は読み物なので白地に直接組む（#1111）。
 *
 * 戻りリンクは上部の1つのみ。本文で終え、末尾に導線（前後の記事など）は置かない。
 * ただし `link` は運営者が意図して置く導線なので、本文の後に出す。
 */
export const AnnouncementDetail: React.FC<{ announcement: Announcement }> = ({ announcement }) => {
  const blocks: AnnouncementBlock[] = announcement.content ?? [
    { type: 'p', children: [announcement.body] },
  ]

  return (
    <article className="mx-auto max-w-[760px]">
      {/* SPはタップ領域を44pxにするため、見た目の余白は負のマージンで相殺する */}
      <Link
        href={'/news' as Route}
        className="-my-2.5 mb-2.5 inline-flex min-h-[44px] items-center gap-1 text-[12.5px] font-semibold text-muted-foreground transition-colors hover:text-primary md:mb-7 md:min-h-0 md:text-[13px]"
      >
        <ChevronLeft aria-hidden="true" className="h-3.5 w-3.5 md:h-4 md:w-4" />
        お知らせ一覧
      </Link>

      <div className="flex items-center gap-2 md:gap-2.5">
        <time
          dateTime={announcement.date}
          className="text-xs tabular-nums text-muted-foreground md:text-[13px]"
        >
          {formatAnnouncementDate(announcement.date)}
        </time>
        <CategoryBadge category={announcement.category} size="sm" />
      </div>

      {/* balance + auto-phrase: 文節単位で行長を均等にし、カタカナ語を分断しない。
          auto-phrase は <html lang="ja"> が前提（layout.tsx で指定済み） */}
      <h1 className="mt-2.5 text-[21px] font-bold leading-[1.5] text-foreground [text-wrap:balance] [word-break:auto-phrase] md:mt-3 md:text-[28px] md:leading-[1.45]">
        {announcement.title}
      </h1>

      <div className="my-5 h-px bg-border md:mb-8 md:mt-7" />

      <div className="flex flex-col gap-4 whitespace-pre-line text-[15px] leading-[1.85] text-foreground/90 md:gap-5 md:text-base md:leading-[1.9]">
        {blocks.map((block, i) => (
          <Block key={i} block={block} />
        ))}
      </div>

      {announcement.link && (
        <Link
          href={announcement.link.href as Route}
          className="mt-6 inline-block text-sm font-medium text-primary hover:underline"
          {...externalProps(announcement.link.href)}
        >
          {announcement.link.label}
        </Link>
      )}
    </article>
  )
}

export default AnnouncementDetail
