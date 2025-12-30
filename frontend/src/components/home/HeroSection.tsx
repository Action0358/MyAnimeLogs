'use client';

import React from 'react';
import Image from 'next/image';
import SearchContainer from './SearchContainer';
import { HERO_LOGO_PATH } from '@/lib/constants';

const HeroSection: React.FC = () => {
  return (
    <section className="relative flex flex-col items-center bg-gradient-to-br from-blue-600 to-blue-800 text-white px-4 pt-44 pb-24 md:pt-52 md:pb-32">
      {/* キャッチコピー */}
      <h2 className="text-base md:text-xl text-center mb-4 md:mb-6 font-medium">
        あなたのアニメ視聴体験を記録・管理しよう
      </h2>

      {/* ロゴ */}
      <div className="mb-16 md:mb-20">
        <Image
          src={HERO_LOGO_PATH}
          alt="MyAnimeLogs"
          width={930}
          height={470}
          priority
          className="h-28 md:h-40 w-auto brightness-0 invert"
        />
      </div>

      {/* 検索コンテナ */}
      <SearchContainer />
    </section>
  );
};

export default HeroSection;
