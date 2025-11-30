'use client';

import React from 'react';
import Image from 'next/image';
import SearchContainer from './SearchContainer';

const HeroSection: React.FC = () => {
  return (
    <section className="relative min-h-[600px] md:min-h-[60vh] flex flex-col items-center justify-center bg-gradient-to-br from-blue-600 to-blue-800 text-white px-4 py-16">
      {/* キャッチコピー */}
      <h2 className="text-sm md:text-lg text-center mb-2 md:mb-3">
        あなたのアニメ視聴体験を記録・管理しよう
      </h2>

      {/* ロゴ */}
      <div className="mb-8 md:mb-12">
        <Image
          src="/logo.png"
          alt="MyAnimeLogs"
          width={400}
          height={80}
          className="w-[250px] md:w-[400px] h-auto brightness-0 invert"
          priority
        />
      </div>

      {/* 検索コンテナ */}
      <SearchContainer />
    </section>
  );
};

export default HeroSection;
