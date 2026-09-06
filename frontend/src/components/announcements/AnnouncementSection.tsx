import React from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import { ChevronRight } from 'lucide-react'
import { Container } from '@/components/layout'
import { Button, SectionHeading } from '@/components/ui'
import { CategoryBadge } from './CategoryBadge'
import { formatAnnouncementDate } from './AnnouncementList'
import { getAnnouncements, HOME_ANNOUNCEMENT_COUNT } from '@/constants/announcements'

/**
 * トップページのお知らせセクション。
 *
 * デザインの意図（モックアップ「News Section Redesign」より）:
 *
 *   - 行を枠付きにして右端にシェブロンを置く。全幅に文字だけを置くと左端に
 *     固まって右側が意図しない空白に見えるが、枠と右端の目印があれば全幅でも
 *     余白が空きに見えない。幅は上のカードグリッドと揃える（絞ると右端が
 *     不揃いになるため）。
 *   - 行を枠付きにし、ホバーで背景を変える。枠も背景もないとクリックできる
 *     ことが伝わらない。
 *   - 「すべて見る」は出す価値があるときだけ出す。表示件数で収まりきる件数
 *     しかないのに置くと、遷移先に続きがある期待を与えてしまう。
 *   - アニメカードのセクションと区切り線で分ける。左右の余白は他ページと
 *     同じにする（Container が担う）。
 *   - 日付とカテゴリをメタ行にまとめ、タイトルを主役にする。3つを均等に
 *     並べると要素間の関係が読めない。
 *
 * 見出し・アクションボタンはカードのセクション（AnimeSection）と同じ
 * SectionHeading・Button（variant="outline-gray"）を使う。背景も含め、
 * ページの他セクションと同じ白地にする。モックアップは見出しを一段小さく・
 * アクションを色付きテキストリンクにし、背景を沈めた帯にして区別していたが、
 * 実装後に「今期アニメ・人気アニメと配色・文字サイズを揃えたい」との
 * フィードバックを受け、区切り線のみで分ける形に変えた。
 *
 * Container の外側に置いているのは、区切り線を全幅に通すため
 * （内側の Container がカードのセクションと同じ size="xl" で左右を揃える）。
 */
export const AnnouncementSection: React.FC<{ className?: string }> = ({ className }) => {
  const all = getAnnouncements()
  const announcements = all.slice(0, HOME_ANNOUNCEMENT_COUNT)
  // 表示しきれない分があるときだけ一覧への導線を出す
  const hasMore = all.length > announcements.length

  return (
    <section className={className} aria-labelledby="home-announcements">
      <div className="border-t border-border">
        <Container size="xl" className="pb-12 pt-9 md:pb-16 md:pt-11">
          <div>
            <SectionHeading
              id="home-announcements"
              action={
                hasMore ? (
                  <Button variant="outline-gray" size="sm" asChild className="font-bold">
                    <Link href={'/news' as Route}>もっと見る</Link>
                  </Button>
                ) : undefined
              }
            >
              お知らせ
            </SectionHeading>

            {announcements.length > 0 ? (
              <ul className="overflow-hidden rounded-lg border border-border bg-card">
                {announcements.map((announcement) => (
                  <li key={announcement.id} className="border-b border-border last:border-b-0">
                    {/* 詳細ページは持たないため、行のリンク先は一覧ページ */}
                    <Link
                      href={'/news' as Route}
                      className="flex items-center gap-3 px-4 py-3.5 transition-colors hover:bg-muted/60 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring md:gap-[18px] md:px-5 md:py-4"
                    >
                      <span className="min-w-0 flex-1">
                        <span className="mb-1 flex items-center gap-2 md:mb-1.5 md:gap-2.5">
                          {/* 日付・カテゴリの文字サイズはカード（AnimeCard）の
                              放送時期表記と揃える */}
                          <time
                            dateTime={announcement.date}
                            className="text-xs tabular-nums text-muted-foreground"
                          >
                            {formatAnnouncementDate(announcement.date)}
                          </time>
                          <CategoryBadge category={announcement.category} size="sm" />
                        </span>
                        {/* タイトルはカード（AnimeCard）のタイトル文字サイズと揃える。
                            全幅を使えるようになったため、省略はせず折り返させる
                            （SPで長いタイトルを隠さない） */}
                        <span className="block text-sm font-semibold leading-[1.4] text-foreground md:text-base">
                          {announcement.title}
                        </span>
                      </span>
                      <ChevronRight
                        aria-hidden="true"
                        className="h-[15px] w-[15px] flex-shrink-0 text-muted-foreground md:h-[17px] md:w-[17px]"
                      />
                    </Link>
                  </li>
                ))}
              </ul>
            ) : (
              /* 0件でもセクションは残す。何も起きていないことが分かるほうがよい */
              <p className="rounded-lg border border-dashed border-border bg-card px-5 py-7 text-center text-[13.5px] text-muted-foreground">
                現在お知らせはありません
              </p>
            )}
          </div>
        </Container>
      </div>
    </section>
  )
}

export default AnnouncementSection
