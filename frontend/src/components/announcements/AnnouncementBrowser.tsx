'use client'

import React, { useEffect, useMemo, useRef, useState } from 'react'
import type { Route } from 'next'
import { useRouter } from 'next/navigation'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { AnnouncementList } from './AnnouncementList'
import { cn } from '@/lib/utils'
import {
  announcementHref,
  ANNOUNCEMENT_CATEGORY_LABELS,
  type Announcement,
  type AnnouncementCategory,
} from '@/constants/announcements'

/** 1ページに出す件数 */
export const NEWS_PAGE_SIZE = 10

/**
 * カテゴリの絞り込みタブと年の区切りを出し始める件数。
 * これ以下では絞り込むほどの量がなく、タブがあるとかえって過剰に見える
 * （モックアップ「News Page Redesign」修正版の指定）。
 */
export const NEWS_FILTER_MIN_COUNT = 6

type CategoryFilter = 'all' | AnnouncementCategory

/** 表示中の項目を年ごとにまとめる。入力は新しい順に並んでいる前提 */
const groupByYear = (items: Announcement[]) =>
  items.reduce<{ year: string; items: Announcement[] }[]>((groups, item) => {
    const year = item.date.slice(0, 4)
    const last = groups[groups.length - 1]
    if (last && last.year === year) {
      last.items.push(item)
    } else {
      groups.push({ year, items: [item] })
    }
    return groups
  }, [])

/** URL のハッシュ（#<id>）からお知らせの id を取り出す */
const readHashId = (): string | null => {
  const hash = window.location.hash.slice(1)
  if (!hash) return null
  try {
    return decodeURIComponent(hash)
  } catch {
    return hash
  }
}

/**
 * お知らせ一覧の本体。件数が増えたときの受け皿として、カテゴリの絞り込み・
 * 年の区切り・ページ送りを持つ。
 *
 * データはビルド時に確定しているが、絞り込みとページ送りは閲覧者の操作で
 * 変わるためクライアントコンポーネントにしている。初回描画は「すべて・
 * 1ページ目」で静的生成と一致するため、ハイドレーション不整合は起きない。
 *
 * 詳細ページ（/news/[id]）を追加する前は /news#<id> が個別の着地点だった（#848）。
 * すでに共有されたリンクが空振りしないよう、該当するIDなら詳細ページへ転送する。
 */
export const AnnouncementBrowser: React.FC<{ announcements: Announcement[] }> = ({
  announcements,
}) => {
  const router = useRouter()
  const [category, setCategory] = useState<CategoryFilter>('all')
  const [page, setPage] = useState(1)
  const listTopRef = useRef<HTMLDivElement>(null)

  const showFilter = announcements.length >= NEWS_FILTER_MIN_COUNT

  // データに存在するカテゴリだけをタブにする。0件になるタブは押しても意味がない
  const categories = useMemo(() => {
    const present = new Set(announcements.map((a) => a.category))
    return (Object.keys(ANNOUNCEMENT_CATEGORY_LABELS) as AnnouncementCategory[]).filter((c) =>
      present.has(c),
    )
  }, [announcements])

  const filtered = useMemo(
    () =>
      category === 'all' ? announcements : announcements.filter((a) => a.category === category),
    [announcements, category],
  )

  const totalPages = Math.max(1, Math.ceil(filtered.length / NEWS_PAGE_SIZE))
  const currentPage = Math.min(page, totalPages)
  const pageItems = filtered.slice(
    (currentPage - 1) * NEWS_PAGE_SIZE,
    currentPage * NEWS_PAGE_SIZE,
  )
  const groups = showFilter
    ? groupByYear(pageItems)
    : [{ year: '', items: pageItems }]

  // 旧アンカー /news#<id> を詳細ページへ転送する。履歴に旧URLを残さないよう replace
  useEffect(() => {
    const redirect = () => {
      const id = readHashId()
      if (id && announcements.some((a) => a.id === id)) {
        router.replace(announcementHref(id) as Route)
      }
    }
    redirect()
    window.addEventListener('hashchange', redirect)
    return () => window.removeEventListener('hashchange', redirect)
  }, [announcements, router])

  const changeCategory = (next: CategoryFilter) => {
    setCategory(next)
    setPage(1)
  }

  const changePage = (next: number) => {
    setPage(next)
    listTopRef.current?.scrollIntoView({ block: 'start', behavior: 'smooth' })
  }

  return (
    <>
      <div ref={listTopRef} className="mb-5 flex scroll-mt-6 items-baseline gap-2 md:mb-6 md:gap-3">
        <h1 className="text-xl font-bold tracking-tight text-foreground md:text-[26px]">
          お知らせ
        </h1>
        {/* 件数の体裁は今期アニメ／人気アニメ／検索と揃える */}
        <span className="flex flex-shrink-0 items-baseline gap-1 text-sm text-muted-foreground">
          <span className="text-base font-bold text-foreground">{filtered.length}</span>件
        </span>
      </div>

      {showFilter && (
        // SPでは横スクロールで収める。折り返すと件数が増えるほどタブの段が増える
        <div
          role="group"
          aria-label="カテゴリで絞り込む"
          className="-mx-4 mb-5 flex gap-1.5 overflow-x-auto px-4 md:mx-0 md:mb-7 md:px-0"
        >
          {(['all', ...categories] as CategoryFilter[]).map((c) => (
            <button
              key={c}
              type="button"
              onClick={() => changeCategory(c)}
              aria-pressed={category === c}
              className={cn(
                'flex-shrink-0 whitespace-nowrap rounded-md px-3 py-1.5 text-[12.5px] font-semibold transition-colors md:px-3.5 md:py-[7px] md:text-[13px]',
                'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
                category === c
                  ? 'bg-primary/10 text-primary'
                  : 'text-muted-foreground hover:bg-muted hover:text-foreground',
              )}
            >
              {c === 'all' ? 'すべて' : ANNOUNCEMENT_CATEGORY_LABELS[c]}
            </button>
          ))}
        </div>
      )}

      {filtered.length > 0 ? (
        <div className="space-y-7">
          {groups.map((group) => (
            <section key={group.year || 'all'} aria-label={group.year ? `${group.year}年` : undefined}>
              {group.year && (
                <h2 className="mb-2 text-[11px] font-bold tracking-[0.08em] text-muted-foreground md:mb-2.5 md:text-xs">
                  {group.year}年
                </h2>
              )}
              <AnnouncementList announcements={group.items} />
            </section>
          ))}
        </div>
      ) : (
        <p className="rounded-lg border border-dashed border-border bg-card px-5 py-10 text-center text-[13.5px] text-muted-foreground md:py-11">
          現在お知らせはありません
        </p>
      )}

      {totalPages > 1 && (
        <NewsPagination currentPage={currentPage} totalPages={totalPages} onChange={changePage} />
      )}
    </>
  )
}

/**
 * お知らせ用の簡素なページ送り。お知らせは月に数件の想定でページ数が
 * 少ないため、検索ページの Pagination（最初へ・最後へ・番号の省略つき）は
 * 持たず、前後と全ページ番号だけを中央に並べる。
 */
const NewsPagination: React.FC<{
  currentPage: number
  totalPages: number
  onChange: (page: number) => void
}> = ({ currentPage, totalPages, onChange }) => {
  const base =
    'flex h-8 w-8 items-center justify-center rounded-md border text-[12.5px] font-semibold transition-colors md:h-[34px] md:w-[34px] md:text-[13px] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring'
  const idle = 'border-border bg-background text-foreground hover:bg-muted'
  const disabled = 'border-border bg-background text-muted-foreground/60 cursor-not-allowed'

  return (
    <nav aria-label="ページネーション" className="mt-6 flex justify-center gap-[5px] md:mt-8 md:gap-1.5">
      <button
        type="button"
        onClick={() => onChange(currentPage - 1)}
        disabled={currentPage === 1}
        aria-label="前のページ"
        className={cn(base, currentPage === 1 ? disabled : idle)}
      >
        <ChevronLeft className="h-4 w-4" aria-hidden="true" />
      </button>
      {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
        <button
          key={p}
          type="button"
          onClick={() => onChange(p)}
          aria-current={p === currentPage ? 'page' : undefined}
          className={cn(
            base,
            p === currentPage ? 'border-primary bg-primary text-primary-foreground' : idle,
          )}
        >
          {p}
        </button>
      ))}
      <button
        type="button"
        onClick={() => onChange(currentPage + 1)}
        disabled={currentPage === totalPages}
        aria-label="次のページ"
        className={cn(base, currentPage === totalPages ? disabled : idle)}
      >
        <ChevronRight className="h-4 w-4" aria-hidden="true" />
      </button>
    </nav>
  )
}

export default AnnouncementBrowser
