'use client'

import React, { useState } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { useQuery } from '@tanstack/react-query'
import { Container } from '@/components/layout'
import { Button } from '@/components/ui'
import { EpisodeList } from '@/components/anime'
import { getAnimesId, getAnimesIdEpisodes } from '@/lib/api/generated/myanimelogs'
import { cn } from '@/lib/utils'

const AnimeDetailPage: React.FC = () => {
  const params = useParams()
  const animeId = typeof params.id === 'string' ? parseInt(params.id, 10) : NaN
  const isValidAnimeId = !isNaN(animeId) && animeId > 0
  const [imageError, setImageError] = useState(false)

  const {
    data: animeData,
    isLoading: isLoadingAnime,
    error: animeError,
  } = useQuery({
    queryKey: ['anime', animeId],
    queryFn: async () => {
      const response = await getAnimesId(animeId)
      return response.data
    },
    enabled: isValidAnimeId,
  })

  const {
    data: episodesData,
    isLoading: isLoadingEpisodes,
  } = useQuery({
    queryKey: ['anime-episodes', animeId],
    queryFn: async () => {
      const response = await getAnimesIdEpisodes(animeId)
      return response.data
    },
    enabled: isValidAnimeId,
  })

  if (isLoadingAnime) {
    return (
      <Container className="py-12">
        <div className="animate-pulse">
          <div className="flex flex-col md:flex-row gap-8">
            <div className="w-full md:w-80 aspect-[3/4] bg-gray-200 rounded-lg" />
            <div className="flex-1 space-y-4">
              <div className="h-8 bg-gray-200 rounded w-3/4" />
              <div className="h-4 bg-gray-200 rounded w-1/4" />
              <div className="h-24 bg-gray-200 rounded" />
            </div>
          </div>
        </div>
      </Container>
    )
  }

  if (animeError || !animeData) {
    return (
      <Container className="py-12">
        <div className="text-center">
          <span className="text-6xl mb-4 block">😢</span>
          <h1 className="text-2xl font-bold text-gray-900 mb-2">アニメが見つかりません</h1>
          <p className="text-gray-600 mb-6">
            お探しのアニメは存在しないか、削除された可能性があります。
          </p>
          <Link href="/search">
            <Button>アニメを検索する</Button>
          </Link>
        </div>
      </Container>
    )
  }

  const anime = animeData
  const episodes = episodesData?.episodes || []
  const imageUrl = anime.cover_image_large || anime.cover_image_medium || anime.image_url

  return (
    <Container className="py-8">
      <div className="max-w-6xl mx-auto">
        {/* パンくずリスト */}
        <nav className="mb-6 text-sm">
          <ol className="flex items-center gap-2 text-gray-500">
            <li>
              <Link href="/" className="hover:text-blue-600">
                ホーム
              </Link>
            </li>
            <li>/</li>
            <li>
              <Link href="/search" className="hover:text-blue-600">
                検索
              </Link>
            </li>
            <li>/</li>
            <li className="text-gray-900 font-medium truncate max-w-[200px]">{anime.title}</li>
          </ol>
        </nav>

        {/* メインコンテンツ */}
        <div className="flex flex-col lg:flex-row gap-8">
          {/* 左カラム: 画像 */}
          <div className="lg:w-80 flex-shrink-0">
            <div className="sticky top-8">
              <div className="relative aspect-[3/4] bg-gray-100 rounded-lg overflow-hidden shadow-lg">
                {imageUrl && !imageError ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img
                    src={imageUrl}
                    alt={anime.title || 'アニメ画像'}
                    className="w-full h-full object-cover"
                    onError={() => setImageError(true)}
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center bg-gradient-to-br from-blue-100 to-purple-100">
                    <span className="text-6xl">📺</span>
                  </div>
                )}
              </div>

              {/* アクションボタン */}
              <div className="mt-4 space-y-2">
                <Button className="w-full">お気に入りに追加</Button>
                <Button variant="outline" className="w-full">
                  視聴リストに追加
                </Button>
              </div>
            </div>
          </div>

          {/* 右カラム: 詳細情報 */}
          <div className="flex-1 min-w-0">
            {/* タイトル */}
            <h1 className="text-3xl font-bold text-gray-900 mb-2">{anime.title}</h1>

            {/* メタ情報 */}
            <div className="flex flex-wrap items-center gap-3 mb-6">
              {anime.season && (
                <span className="inline-flex items-center px-3 py-1 bg-blue-100 text-blue-800 rounded-full text-sm font-medium">
                  {anime.season}
                </span>
              )}
              {anime.status && (
                <span
                  className={cn(
                    'inline-flex items-center px-3 py-1 rounded-full text-sm font-medium',
                    anime.status === 'airing'
                      ? 'bg-green-100 text-green-800'
                      : 'bg-gray-100 text-gray-800'
                  )}
                >
                  {anime.status === 'airing' ? '放送中' : '完結'}
                </span>
              )}
            </div>

            {/* 説明 */}
            {anime.description && (
              <div className="mb-8">
                <h2 className="text-lg font-semibold text-gray-900 mb-2">あらすじ</h2>
                <p className="text-gray-700 leading-relaxed whitespace-pre-wrap">
                  {anime.description}
                </p>
              </div>
            )}

            {/* エピソード一覧 */}
            <div>
              <h2 className="text-lg font-semibold text-gray-900 mb-4">
                エピソード一覧
                {episodes.length > 0 && (
                  <span className="ml-2 text-sm font-normal text-gray-500">
                    ({episodes.length}話)
                  </span>
                )}
              </h2>
              <EpisodeList
                episodes={episodes}
                isLoading={isLoadingEpisodes}
                animeId={animeId}
              />
            </div>
          </div>
        </div>
      </div>
    </Container>
  )
}

export default AnimeDetailPage
