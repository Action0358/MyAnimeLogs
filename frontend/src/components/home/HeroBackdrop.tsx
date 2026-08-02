'use client';

import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { currentSeasonAnimeQuery } from '@/lib/queries/currentSeasonAnime';

/** 背景に敷くタイル数（レスポンシブで表示数を絞る） */
const BACKDROP_TILE_COUNT = 12;

/** タイルの表示ブレークポイント（index順に対応）
 * SP: 4枚 / sm: 8枚 / lg: 12枚 で横幅を埋める
 */
const tileVisibilityClass = (index: number): string => {
  if (index < 4) return '';
  if (index < 8) return 'hidden sm:block';
  return 'hidden lg:block';
};

/**
 * ヒーロー背景。今期アニメのカバー画像をタイル状に敷き、
 * ぼかし＋primaryトーンのオーバーレイで前景テキストの可読性を確保する。
 *
 * クエリ定義は CurrentSeasonAnime と共有しているため、追加のAPIリクエストは発生しない。
 */
const HeroBackdrop: React.FC = () => {
  const { data } = useQuery(currentSeasonAnimeQuery);

  const tiles = (data?.animes ?? [])
    .map((anime) => ({
      id: anime.id,
      url: anime.cover_image_large || anime.cover_image_medium || anime.image_url,
    }))
    .filter((tile): tile is { id: number; url: string } => Boolean(tile.url))
    .slice(0, BACKDROP_TILE_COUNT);

  return (
    <div className="absolute inset-0 overflow-hidden pointer-events-none" aria-hidden="true">
      {/* 画像取得前・取得失敗時のベース（CLS回避のため常に敷く） */}
      <div className="absolute inset-0 bg-gradient-to-br from-primary to-primary/70" />

      {tiles.length > 0 && (
        <div className="absolute inset-0 flex scale-110 blur-[3px]">
          {tiles.map((tile, index) => (
            <div key={tile.id} className={`flex-1 min-w-0 h-full ${tileVisibilityClass(index)}`}>
              {/* AnimeCardと同様、外部CDNドメインが多岐にわたるためimgタグを使用 */}
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={tile.url}
                alt=""
                className="w-full h-full object-cover"
                loading="lazy"
                decoding="async"
              />
            </div>
          ))}
        </div>
      )}

      {/* ブランドトーン＋可読性オーバーレイ */}
      <div className="absolute inset-0 bg-gradient-to-b from-primary/85 via-primary/75 to-primary/90" />
      {/* 下端をページ背景へ馴染ませる */}
      <div className="absolute inset-x-0 bottom-0 h-24 bg-gradient-to-t from-black/25 to-transparent" />
    </div>
  );
};

export default HeroBackdrop;
