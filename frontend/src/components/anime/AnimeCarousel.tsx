'use client'

import React, { useRef, useState, useEffect, useCallback } from 'react'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'
import { AnimeCard, type AnimeCardData } from './AnimeCard'
import { SkeletonCard } from './SkeletonCard'

/** カルーセルの定数 */
const CAROUSEL_CONSTANTS = {
  /** スクロール判定の閾値（px） */
  SCROLL_THRESHOLD: 4,
  /** スクロール量の係数（コンテナ幅に対する割合） */
  SCROLL_AMOUNT_RATIO: 0.75,
  /** プログレスバーの最小幅（%） */
  MIN_PROGRESS_WIDTH: 8,
  /** スケルトン表示数 */
  SKELETON_COUNT: 6,
} as const

/** カルーセルアイテムの幅計算クラス（レスポンシブ対応）
 * 幅計算の共通形式: (100% - gap数 × gap幅) / 表示列数
 * SP版: 2.5列表示（2枚+3枚目が約50%チラ見え）
 *   gap-3(0.75rem) × 2(ビューポート内のgap数) = 1.5rem
 * sm以上: gap-4(1rem) で整数列表示
 */
const CAROUSEL_ITEM_WIDTH_CLASSES =
  'flex-none w-[calc((100%-1.5rem)/2.5)] sm:w-[calc((100%-2rem)/3)] md:w-[calc((100%-3rem)/4)] lg:w-[calc((100%-4rem)/5)] xl:w-[calc((100%-5rem)/6)]'

/** グラデーションマスクの共通クラス */
const GRADIENT_MASK_CLASSES =
  'absolute top-0 bottom-2 w-16 md:w-20 from-gray-50 to-transparent pointer-events-none z-[1]'

interface AnimeCarouselProps {
  animes: AnimeCardData[]
  className?: string
}

export const AnimeCarousel: React.FC<AnimeCarouselProps> = ({
  animes,
  className,
}) => {
  const scrollRef = useRef<HTMLDivElement>(null)
  const [canScrollLeft, setCanScrollLeft] = useState(false)
  const [canScrollRight, setCanScrollRight] = useState(false)
  const [scrollProgress, setScrollProgress] = useState(0)

  const checkScroll = useCallback(() => {
    const el = scrollRef.current
    if (!el) return
    const { scrollLeft, scrollWidth, clientWidth } = el
    setCanScrollLeft(scrollLeft > CAROUSEL_CONSTANTS.SCROLL_THRESHOLD)
    setCanScrollRight(scrollLeft < scrollWidth - clientWidth - CAROUSEL_CONSTANTS.SCROLL_THRESHOLD)
    const maxScroll = scrollWidth - clientWidth
    setScrollProgress(maxScroll > 0 ? (scrollLeft / maxScroll) * 100 : 0)
  }, [])

  // イベントリスナーの登録（初回のみ）
  useEffect(() => {
    const el = scrollRef.current
    if (!el) return

    el.addEventListener('scroll', checkScroll, { passive: true })
    const ro = new ResizeObserver(checkScroll)
    ro.observe(el)

    return () => {
      el.removeEventListener('scroll', checkScroll)
      ro.disconnect()
    }
  }, [checkScroll])

  // animes変更時のスクロール状態チェック
  useEffect(() => {
    checkScroll()
  }, [animes, checkScroll])

  const scroll = (dir: 'left' | 'right') => {
    const el = scrollRef.current
    if (!el) return
    const amount = el.clientWidth * CAROUSEL_CONSTANTS.SCROLL_AMOUNT_RATIO
    el.scrollBy({ left: dir === 'left' ? -amount : amount, behavior: 'smooth' })
  }

  return (
    <div className={cn('group/carousel relative', className)}>
      {/* 左ナビゲーション */}
      {canScrollLeft && (
        <button
          onClick={() => scroll('left')}
          className={cn(
            'absolute -left-4 top-[calc(50%-40px)] -translate-y-1/2 z-10',
            'w-10 h-10 md:w-11 md:h-11 rounded-full',
            'bg-white/95 shadow-lg border border-gray-200',
            'flex items-center justify-center',
            'hover:bg-white hover:scale-110 hover:shadow-xl',
            'transition-all duration-200',
            'opacity-0 group-hover/carousel:opacity-100',
            'focus:opacity-100 focus:outline-none focus:ring-2 focus:ring-blue-500'
          )}
          aria-label="前へスクロール"
        >
          <ChevronLeft className="w-5 h-5 text-gray-700" />
        </button>
      )}

      {/* スクロールコンテナのラッパー（左右端透過表示用） */}
      <div className="relative overflow-hidden">
        {/* 左端グラデーションマスク（スクロール可能を示唆） */}
        {canScrollLeft && (
          <div
            className={cn(GRADIENT_MASK_CLASSES, 'left-0 bg-gradient-to-r')}
            aria-hidden="true"
          />
        )}

        {/* スクロールコンテナ */}
        <div
          ref={scrollRef}
          role="list"
          aria-label="アニメ一覧"
          tabIndex={0}
          className="flex gap-3 sm:gap-4 overflow-x-auto scrollbar-hide scroll-smooth pb-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 rounded-lg"
        >
          {animes.map((anime) => (
            <div
              key={anime.id}
              role="listitem"
              className={CAROUSEL_ITEM_WIDTH_CLASSES}
            >
              <AnimeCard anime={anime} size="medium" />
            </div>
          ))}
        </div>

        {/* 右端グラデーションマスク（スクロール可能を示唆） */}
        {canScrollRight && (
          <div
            className={cn(GRADIENT_MASK_CLASSES, 'right-0 bg-gradient-to-l')}
            aria-hidden="true"
          />
        )}
      </div>

      {/* 右ナビゲーション */}
      {canScrollRight && (
        <button
          onClick={() => scroll('right')}
          className={cn(
            'absolute -right-4 top-[calc(50%-40px)] -translate-y-1/2 z-10',
            'w-10 h-10 md:w-11 md:h-11 rounded-full',
            'bg-white/95 shadow-lg border border-gray-200',
            'flex items-center justify-center',
            'hover:bg-white hover:scale-110 hover:shadow-xl',
            'transition-all duration-200',
            'opacity-0 group-hover/carousel:opacity-100',
            'focus:opacity-100 focus:outline-none focus:ring-2 focus:ring-blue-500'
          )}
          aria-label="次へスクロール"
        >
          <ChevronRight className="w-5 h-5 text-gray-700" />
        </button>
      )}

      {/* スクロールプログレスバー */}
      {animes.length > 0 && (
        <div className="mt-4 flex justify-center">
          <div className="w-24 h-1 bg-gray-200 rounded-full overflow-hidden">
            <div
              className="h-full bg-blue-500 rounded-full transition-all duration-150"
              style={{ width: `${Math.max(scrollProgress, CAROUSEL_CONSTANTS.MIN_PROGRESS_WIDTH)}%` }}
            />
          </div>
        </div>
      )}
    </div>
  )
}

/** カルーセル用スケルトン */
export const CarouselSkeleton: React.FC = () => (
  <div className="flex gap-3 sm:gap-4 overflow-hidden">
    {Array.from({ length: CAROUSEL_CONSTANTS.SKELETON_COUNT }).map((_, i) => (
      <div key={i} className={CAROUSEL_ITEM_WIDTH_CLASSES}>
        <SkeletonCard />
      </div>
    ))}
  </div>
)

export default AnimeCarousel
