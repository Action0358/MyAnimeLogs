'use client'

import type React from 'react'
import { Container } from '@/components/layout'
import { CurrentSeasonAnime, PopularAnime } from '@/components/anime'
import { HeroSection } from '@/components/home'
import { AnnouncementSection } from '@/components/announcements'

const HomePage: React.FC = () => {
  return (
    <>
      {/* Heroセクション */}
      <HeroSection />

      {/* 既存のコンテンツ。カルーセルの矢印など（z-10）がヒーローの検索履歴に重ならないよう、
          z-0 のスタッキングコンテキストに閉じ込める */}
      <Container size="xl" className="relative z-0 py-10 md:py-14">
        {/* 今期アニメセクション */}
        <div id="current-season" className="mb-12 md:mb-16 scroll-mt-24">
          <CurrentSeasonAnime />
        </div>

        {/* 人気アニメセクション */}
        <div>
          <PopularAnime />
        </div>
      </Container>

      {/* お知らせセクション。区切り線を全幅に通すため Container の外側に置く
          （セクション側で内側に Container を持つ） */}
      <AnnouncementSection />
    </>
  )
}

export default HomePage
