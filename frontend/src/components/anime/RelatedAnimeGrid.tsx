'use client'

import React, { useState } from 'react'
import { ChevronDown, ChevronUp } from 'lucide-react'
import { AnimeGrid, type AnimeGridProps } from './AnimeGrid'
import type { Anime, RelatedAnime } from '@/lib/api/generated/myanimelogs'

/** 関連作品の初期表示件数。APIは最大48件返すが、全部並べると詳細ページが
 *  縦に伸びすぎるため、既定ではここまでに抑えて残りは展開で見せる */
export const RELATED_INITIAL_COUNT = 12

/** 最も列数が多いPC表示（6列）の1行ぶん。残りがこの件数以内なら、
 *  クリックさせるほどの量ではないため折りたたまずそのまま全件表示する */
const ONE_ROW_AT_WIDEST = 6

/** 関連作品は一覧・検索と同じ AnimeCard 規格で表示するため、
 *  APIが返す縮小フィールド（RelatedAnime）を AnimeGrid が要求する Anime 形に合わせる。
 *  status等は含まれないため未設定になるが、AnimeCard 側は任意項目として扱う。 */
const toAnimeCardData = (related: RelatedAnime[]): (Anime & { id: number })[] =>
  related.map((r) => ({
    id: r.id,
    title: r.title,
    season: r.season ?? undefined,
    image_url: r.image_url ?? undefined,
    cover_image_extra_large: r.cover_image_extra_large,
    cover_image_large: r.cover_image_large,
    cover_image_medium: r.cover_image_medium,
  }))

export interface RelatedAnimeGridProps {
  related: RelatedAnime[]
  columns?: AnimeGridProps['columns']
  /** 関連作品の詳細から戻る先（表示中の作品の詳細ページ） */
  from?: string
}

/**
 * 関連作品のグリッド。RELATED_INITIAL_COUNT を超える分は折りたたむ。
 *
 * 収まる作品では展開ボタンを出さないので、大半の作品（関連作品数の平均は5件）の
 * 見た目は変わらない。名探偵コナン36作品・ルパン三世35作品のようなシリーズだけが
 * 折りたたまれる。
 */
export const RelatedAnimeGrid: React.FC<RelatedAnimeGridProps> = ({ related, columns, from }) => {
  const [isExpanded, setIsExpanded] = useState(false)

  const rawHiddenCount = related.length - RELATED_INITIAL_COUNT
  // 残りが1行以内なら、ボタンを押させず最初から全件表示する
  const fitsInOneMoreRow = rawHiddenCount > 0 && rawHiddenCount <= ONE_ROW_AT_WIDEST
  const effectiveInitialCount = fitsInOneMoreRow ? related.length : RELATED_INITIAL_COUNT
  const hiddenCount = related.length - effectiveInitialCount
  const visible = isExpanded ? related : related.slice(0, effectiveInitialCount)

  return (
    <>
      <AnimeGrid animes={toAnimeCardData(visible)} columns={columns} cardSize="small" from={from} />
      {hiddenCount > 0 && (
        <button
          type="button"
          onClick={() => setIsExpanded((prev) => !prev)}
          aria-expanded={isExpanded}
          className="mt-4 inline-flex items-center gap-1 rounded text-sm font-medium text-primary hover:underline focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
        >
          {isExpanded ? (
            <>
              折りたたむ
              <ChevronUp className="w-4 h-4" />
            </>
          ) : (
            <>
              他{hiddenCount}件を表示
              <ChevronDown className="w-4 h-4" />
            </>
          )}
        </button>
      )}
    </>
  )
}

export default RelatedAnimeGrid
