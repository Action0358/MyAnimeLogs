'use client';

import type React from 'react';
import { Container } from '@/components/layout';
import { CurrentSeasonAnime, PopularAnime } from '@/components/anime';
import { HeroSection } from '@/components/home';
import { AnnouncementSection } from '@/components/announcements';

const HomePage: React.FC = () => {
  return (
    <>
      {/* Heroセクション */}
      <HeroSection />

      {/* 既存のコンテンツ */}
      <Container size="xl" className="py-10 md:py-14">
        {/* 今期アニメセクション */}
        <div id="current-season" className="mb-12 md:mb-16 scroll-mt-24">
          <CurrentSeasonAnime />
        </div>

        {/* 人気アニメセクション */}
        <div>
          <PopularAnime />
        </div>
      </Container>

      {/* お知らせセクション。運営からの情報で作品カードとは性質が違うため、
          背景をわずかに沈めた帯にして切り分ける。帯を全幅にするため
          Container の外側に置く（セクション側で内側に Container を持つ） */}
      <AnnouncementSection />
    </>
  );
};

export default HomePage;