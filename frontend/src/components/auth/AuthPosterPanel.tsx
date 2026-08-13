'use client'

import React from 'react'
import { useQuery } from '@tanstack/react-query'
import { animeApi, type DBAnime } from '@/lib/api/annict'

/** デスクトップは3×3、モバイルは横1列4枚 */
const DESKTOP_POSTER_COUNT = 9
const MOBILE_POSTER_COUNT = 4

const posterUrl = (anime: DBAnime): string | undefined =>
  anime.cover_image_large ?? anime.cover_image_medium ?? anime.image_url

interface AuthPosterPanelProps {
  /** パネル下部の見出し。ログインと新規登録で出し分ける */
  headline: React.ReactNode
  /** 見出し下の補足。新規登録ではベネフィットのリストを渡す */
  children?: React.ReactNode
}

/**
 * 認証画面の左（モバイルでは上）に置くビジュアルパネル。
 *
 * 人気アニメのカバー画像をグリッド表示し、暗いグラデーションを重ねて
 * その上にコピーを載せる。画像の取得に失敗しても背景色とグラデーションだけで
 * 成立するため、認証フロー自体は画像の可否に依存しない。
 */
export const AuthPosterPanel: React.FC<AuthPosterPanelProps> = ({ headline, children }) => {
  const { data } = useQuery({
    queryKey: ['authPosters'],
    queryFn: () => animeApi.getPopularAnimes(DESKTOP_POSTER_COUNT),
    // 認証画面の装飾のため、失敗してもリトライせず静かに諦める
    retry: false,
    staleTime: 1000 * 60 * 60,
  })

  const posters = (data?.animes ?? []).map(posterUrl).filter((url): url is string => !!url)

  return (
    <div className="relative overflow-hidden bg-[hsl(222_25%_12%)] max-lg:h-[168px] lg:min-h-[620px]">
      {posters.length > 0 && (
        <div
          className="absolute inset-0 grid grid-cols-4 gap-2 p-3.5 lg:grid-cols-3 lg:gap-2.5 lg:p-7"
          aria-hidden="true"
        >
          {posters.slice(0, DESKTOP_POSTER_COUNT).map((url, index) => (
            <div
              key={url}
              className={`aspect-[2/3] overflow-hidden rounded-md bg-white/5 lg:rounded-lg ${
                index >= MOBILE_POSTER_COUNT ? 'max-lg:hidden' : ''
              }`}
            >
              {/* 外部画像URLは様々なドメインから取得されるため、next/imageではなくimgタグを使用 */}
              <img src={url} alt="" loading="lazy" className="h-full w-full object-cover" />
            </div>
          ))}
        </div>
      )}

      <div className="absolute inset-0 bg-gradient-to-b from-[hsl(222_25%_12%/0.55)] to-[hsl(222_25%_12%/0.95)] lg:from-[hsl(222_25%_12%/0.62)] lg:to-[hsl(222_25%_12%/0.94)]" />

      <div className="relative flex h-full flex-col justify-between p-[18px] lg:p-10">
        <span className="text-[15px] font-bold tracking-tight text-white lg:text-lg">
          MyAnimeLogs
        </span>
        <div>
          <h2 className="text-base font-bold leading-relaxed text-white lg:text-[26px]">
            {headline}
          </h2>
          {children}
        </div>
      </div>
    </div>
  )
}
