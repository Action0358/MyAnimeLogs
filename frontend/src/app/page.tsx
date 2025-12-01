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
      <Container className="py-12">
        {/* 今期アニメセクション */}
        <div className="mb-16">
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