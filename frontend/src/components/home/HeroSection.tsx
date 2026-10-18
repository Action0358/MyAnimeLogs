'use client';

import React from 'react';
import Image from 'next/image';
import SearchContainer from './SearchContainer';
import HeroBackdrop from './HeroBackdrop';
import { HERO_LOGO_PATH } from '@/lib/constants';

const HeroSection: React.FC = () => {
  // z-[1]を使用する理由:
  // モバイルメニュー（z-40, z-50）やヘッダー（z-10）より低い階層に配置するため、新しいスタッキングコンテキストを作成。
  // これにより、子要素のz-index（タブのz-10など）がモバイルメニューに影響を与えなくなります。
  // 一方、検索欄の「最近の検索」はヒーローの下端からはみ出して後続の一覧に重なるため、
  // 後続の一覧（page.tsx で z-0 のスタッキングコンテキスト）より1段上にする。

  return (
    <section className="relative flex flex-col items-center text-primary-foreground px-4 pt-36 pb-20 md:pt-44 md:pb-28 z-[1]">
      <HeroBackdrop />

      {/* 前景コンテンツ（背景の上に重ねる） */}
      <div className="relative flex flex-col items-center w-full">
        {/* ロゴ */}
        <div className="mb-5 md:mb-6">
          <Image
            src={HERO_LOGO_PATH}
            alt="MyAnimeLogs"
            width={930}
            height={470}
            priority
            className="h-24 md:h-36 w-auto brightness-0 invert drop-shadow-lg"
          />
        </div>

        {/* キャッチコピー（ページ唯一のh1。ロゴは画像のため見出しにならない） */}
        <h1 className="text-sm md:text-lg text-center mb-10 md:mb-12 font-bold text-primary-foreground/90 drop-shadow">
          あなたのアニメ視聴体験を記録・管理しよう
        </h1>

        {/* 検索コンテナ */}
        <SearchContainer />
      </div>
    </section>
  );
};

export default HeroSection;
