import React from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import { ChevronRight } from 'lucide-react'
import { Container } from '@/components/layout'
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
 *   - 背景をわずかに沈めた帯にして、アニメカードのセクションと切り分ける。
 *     同じ見出し規格で並べると性質の違う情報が同列に見える。左右の余白は
 *     他ページと同じにする（Container が担う）。
 *   - 日付とカテゴリをメタ行にまとめ、タイトルを主役にする。3つを均等に
 *     並べると要素間の関係が読めない。
 *
 * 帯にするため Container の外側に置く。ページ側で Container に入れないこと。
 * 内側の Container はカードのセクションと同じ size="xl" にして右端を揃える。
 */
export const AnnouncementSection: React.FC<{ className?: string }> = ({ className }) => {
  const all = getAnnouncements()
  const announcements = all.slice(0, HOME_ANNOUNCEMENT_COUNT)
  // 表示しきれない分があるときだけ一覧への導線を出す
  const hasMore = all.length > announcements.length

  return (
    <section className={className} aria-labelledby="home-announcements">
      <div className="border-t border-border bg-muted/40">
        <Container size="xl" className="pb-8 pt-[26px] md:pb-12 md:pt-11">
          <div>
            <div className="mb-3.5 flex items-center justify-between gap-4 md:mb-[18px]">
              <div className="flex items-center gap-2.5">
                <span
                  aria-hidden="true"
                  className="h-[17px] w-1 flex-shrink-0 rounded-sm bg-primary md:h-[19px]"
                />
                <h2
                  id="home-announcements"
                  className="text-[17px] font-bold tracking-tight text-foreground md:text-xl"
                >
                  お知らせ
                </h2>
              </div>
              {hasMore && (
                <Link
                  href={'/news' as Route}
                  className="flex-shrink-0 whitespace-nowrap text-xs font-semibold text-primary hover:underline md:text-[13px]"
                >
                  すべて見る →
                </Link>
              )}
            </div>

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
                        <span className="mb-1 flex items-center gap-2 md:mb-[5px] md:gap-2.5">
                          <time
                            dateTime={announcement.date}
                            className="text-[11px] tabular-nums text-muted-foreground md:text-xs"
                          >
                            {formatAnnouncementDate(announcement.date)}
                          </time>
                          <CategoryBadge category={announcement.category} size="sm" />
                        </span>
                        {/* タイトルを主役にする。全幅を使えるようになったため、
                            省略はせず折り返させる（SPで長いタイトルを隠さない） */}
                        <span className="block text-[13px] font-semibold leading-[1.6] text-foreground md:text-[14.5px] md:leading-[1.55]">
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
