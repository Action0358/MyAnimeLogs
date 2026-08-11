'use client';

import type React from 'react';
import { Container } from '@/components/layout';
import { CurrentSeasonAnime, PopularAnime } from '@/components/anime';
import { HeroSection } from '@/components/home';

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
    </>
  );
};

export default HomePage;