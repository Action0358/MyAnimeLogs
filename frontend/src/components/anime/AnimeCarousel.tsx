'use client'

import React, { useRef, useState, useEffect, useCallback } from 'react'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'
import { AnimeCard } from './AnimeCard'
import { SkeletonCard } from './SkeletonCard'
import type { Anime } from '@/lib/api/generated/myanimelogs'

interface AnimeCarouselProps {
  animes: (Anime & { id: number })[]
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
    setCanScrollLeft(scrollLeft > 4)
    setCanScrollRight(scrollLeft < scrollWidth - clientWidth - 4)
    const maxScroll = scrollWidth - clientWidth
    setScrollProgress(maxScroll > 0 ? (scrollLeft / maxScroll) * 100 : 0)
  }, [])

  useEffect(() => {
    checkScroll()
    const el = scrollRef.current
    if (!el) return
    el.addEventListener('scroll', checkScroll, { passive: true })
    const ro = new ResizeObserver(checkScroll)
    ro.observe(el)
    return () => {
      el.removeEventListener('scroll', checkScroll)
      ro.disconnect()
    }
  }, [checkScroll, animes])

  const scroll = (dir: 'left' | 'right') => {
    const el = scrollRef.current
    if (!el) return
    const amount = el.clientWidth * 0.75
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

      {/* スクロールコンテナ */}
      <div
        ref={scrollRef}
        role="list"
        aria-label="アニメ一覧"
        className="flex gap-3 md:gap-4 overflow-x-auto scrollbar-hide scroll-smooth pb-2"
      >
        {animes.map((anime) => (
          <div
            key={anime.id}
            role="listitem"
            className="flex-none w-[calc((100%-0.75rem)/2)] sm:w-[calc((100%-1.5rem)/3)] md:w-[calc((100%-2.25rem)/4)] lg:w-[calc((100%-3rem)/5)] xl:w-[calc((100%-3.75rem)/6)]"
          >
            <AnimeCard anime={anime} size="medium" />
          </div>
        ))}
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
              style={{ width: `${Math.max(scrollProgress, 8)}%` }}
            />
          </div>
        </div>
      )}
    </div>
  )
}

/** カルーセル用スケルトン */
export const CarouselSkeleton: React.FC = () => (
  <div className="flex gap-3 md:gap-4 overflow-hidden">
    {Array.from({ length: 6 }).map((_, i) => (
      <div
        key={i}
        className="flex-none w-[calc((100%-0.75rem)/2)] sm:w-[calc((100%-1.5rem)/3)] md:w-[calc((100%-2.25rem)/4)] lg:w-[calc((100%-3rem)/5)] xl:w-[calc((100%-3.75rem)/6)]"
      >
        <SkeletonCard />
      </div>
    ))}
  </div>
)

export default AnimeCarousel
