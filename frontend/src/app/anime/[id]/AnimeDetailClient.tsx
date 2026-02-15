'use client'

import React, { useState, useEffect } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { useQuery } from '@tanstack/react-query'
import { decode as heDecodeEntities } from 'he'
import { Container } from '@/components/layout'
import { Button } from '@/components/ui'
import { EpisodeList, WatchProviders } from '@/components/anime'
import { getAnimesId, getAnimesIdEpisodes } from '@/lib/api/generated/myanimelogs'
import { cn } from '@/lib/utils'

const AnimeDetailClient: React.FC = () => {
  const params = useParams()
  const [imageError, setImageError] = useState(false)
  const [animeId, setAnimeId] = useState<number>(NaN)
  
  // URLから実際のIDを取得（Cloudflare Pages Functionsによるリライト対応）
  useEffect(() => {
    if (params.id === '_') {
      const pathParts = window.location.pathname.split('/')
      const idFromUrl = pathParts[pathParts.indexOf('anime') + 1]
      if (idFromUrl && idFromUrl !== '_') {
        setAnimeId(parseInt(idFromUrl, 10))
      }
    } else if (typeof params.id === 'string') {
      setAnimeId(parseInt(params.id, 10))
    }
  }, [params.id])
  
  const isValidAnimeId = !isNaN(animeId) && animeId > 0

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

  // ローディング表示
  if (isNaN(animeId) || isLoadingAnime) {
    return (
      <Container className="py-12">
        <div className="animate-pulse max-w-4xl mx-auto">
          <div className="flex flex-col md:flex-row gap-8">
            <div className="w-full md:w-64 aspect-[3/4] bg-gray-200 rounded-lg" />
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

  if (!isValidAnimeId) {
    return (
      <Container className="py-12">
        <div className="text-center">
          <span className="text-6xl mb-4 block">😢</span>
          <h1 className="text-2xl font-bold text-gray-900 mb-2">無効なアニメIDです</h1>
          <p className="text-gray-600 mb-6">正しいURLでアクセスしてください。</p>
          <Link href="/search">
            <Button>アニメを検索する</Button>
          </Link>
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
          <p className="text-gray-600 mb-6">お探しのアニメは存在しないか、削除された可能性があります。</p>
          <Link href="/search">
            <Button>アニメを検索する</Button>
          </Link>
        </div>
      </Container>
    )
  }

  const anime = animeData
  const episodes = episodesData?.episodes || []
  const imageUrl = anime.cover_image_extra_large || anime.cover_image_large || anime.cover_image_medium || anime.image_url

  // メディアタイプの日本語変換
  const getMediaLabel = (media: string | undefined | null) => {
    if (!media) return null
    const mediaMap: Record<string, string> = {
      'tv': 'TV',
      'movie': '映画',
      'ova': 'OVA',
      'ona': 'ONA',
      'special': 'スペシャル',
      'music': 'ミュージック',
    }
    return mediaMap[media.toLowerCase()] || media
  }

  // 放送開始日のフォーマット
  const formatReleasedOn = (releasedOn: string | undefined | null) => {
    if (!releasedOn) return null
    try {
      const date = new Date(releasedOn)
      return `${date.getFullYear()}年${date.getMonth() + 1}月${date.getDate()}日`
    } catch {
      return releasedOn
    }
  }

  // HTMLタグを除去するユーティリティ関数
  // DOMParserを使用してより堅牢にHTMLエンティティをデコード
  const stripHtmlTags = (html: string | undefined | null): string => {
    if (!html) return ''

    // ブラウザ環境ではDOMParserを使用
    if (typeof window !== 'undefined') {
      // <br>タグを改行に変換してから処理
      const htmlWithLineBreaks = html.replace(/<br\s*\/?>/gi, '\n')
      const doc = new DOMParser().parseFromString(htmlWithLineBreaks, 'text/html')
      return doc.body.textContent?.trim() || ''
    }

    // SSR環境では'he'ライブラリを使用してすべてのHTMLエンティティを正しくデコード
    // これにより&copy;, &trade;, &rsquo;などすべてのエンティティに対応
    let text = html.replace(/<br\s*\/?>/gi, '\n').replace(/<[^>]*>/g, '')
    return heDecodeEntities(text).trim()
  }

  // 型アサーション（APIレスポンスの型に新しいフィールドが含まれていない場合の対応）
  const animeWithDetails = anime as typeof anime & {
    genres?: string[]
    media?: string
    episodes_count?: number
    released_on?: string
    official_site_url?: string
    title_kana?: string
    title_en?: string
    watchers_count?: number
    mal_anime_id?: string
    annict_id?: number
    description?: string
    watch_providers?: Array<{
      provider_id: number
      provider_name: string
      logo_url: string
    }>
  }

  // 映画かどうかを判定（エピソード一覧・話数の表示制御に使用）
  const isMovie = animeWithDetails.media?.toLowerCase() === 'movie'

  return (
    <Container className="py-8">
      <div className="max-w-4xl mx-auto">
        {/* パンくずリスト */}
        <nav className="mb-6 text-sm">
          <ol className="flex items-center gap-2 text-gray-500">
            <li>
              <Link href="/" className="hover:text-blue-600">ホーム</Link>
            </li>
            <li>/</li>
            <li>
              <Link href="/search" className="hover:text-blue-600">検索</Link>
            </li>
            <li>/</li>
            <li className="text-gray-900 font-medium truncate max-w-[200px]">{anime.title}</li>
          </ol>
        </nav>

        {/* ヘッダーセクション: カバー画像 + 基本情報 */}
        <div className="flex flex-col md:flex-row gap-6 mb-8 bg-white rounded-lg border p-4 md:p-6">
          {/* カバー画像 */}
          <div className="w-full md:w-48 flex-shrink-0">
            <div className="relative aspect-[3/4] bg-gray-100 rounded-lg overflow-hidden shadow-md">
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
                  <span className="text-5xl">📺</span>
                </div>
              )}
            </div>
          </div>

          {/* 基本情報 */}
          <div className="flex-1 min-w-0">
            {/* タイトル */}
            <h1 className="text-2xl md:text-3xl font-bold text-gray-900 mb-1">{anime.title}</h1>

            {/* メタ情報（放送時期・状況） */}
            <div className="flex flex-wrap items-center gap-2 mb-3">
              {anime.season && (
                <span className="inline-flex items-center px-2.5 py-1 bg-blue-100 text-blue-800 rounded-full text-sm font-medium">
                  {anime.season}
                </span>
              )}
              {anime.status && (
                <span
                  className={cn(
                    'inline-flex items-center px-2.5 py-1 rounded-full text-sm font-medium',
                    anime.status === 'airing'
                      ? 'bg-green-100 text-green-800'
                      : 'bg-gray-100 text-gray-800'
                  )}
                >
                  {anime.status === 'airing' ? '放送中' : '完結'}
                </span>
              )}
            </div>

            {/* ジャンルタグ */}
            {animeWithDetails.genres && animeWithDetails.genres.length > 0 && (
              <div className="flex flex-wrap gap-1.5 mb-4">
                {animeWithDetails.genres.map((genre) => (
                  <Link
                    key={genre}
                    href={`/search?genre=${encodeURIComponent(genre)}`}
                    className="inline-flex items-center px-2 py-0.5 bg-purple-50 text-purple-700 rounded text-xs hover:bg-purple-100 transition-colors"
                  >
                    {genre}
                  </Link>
                ))}
              </div>
            )}

            {/* アクションボタン */}
            <div className="flex flex-wrap gap-2">
              <Button size="sm">お気に入りに追加</Button>
              <Button variant="outline" size="sm">視聴リストに追加</Button>
            </div>
          </div>
        </div>

        {/* あらすじセクション */}
        {animeWithDetails.description && (
          <div className="mb-6 bg-white rounded-lg border p-4 md:p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-3">あらすじ</h2>
            <p className="text-gray-700 leading-relaxed whitespace-pre-wrap text-sm md:text-base">
              {stripHtmlTags(animeWithDetails.description)}
            </p>
          </div>
        )}

        {/* 作品情報セクション */}
        {(animeWithDetails.released_on || animeWithDetails.media || animeWithDetails.episodes_count || episodes.length > 0 || animeWithDetails.official_site_url) && (
          <div className="mb-6 bg-white rounded-lg border p-4 md:p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-3">作品情報</h2>
            <dl className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-sm">
              {animeWithDetails.released_on && (
                <div className="flex">
                  <dt className="text-gray-500 w-24 flex-shrink-0">放送開始</dt>
                  <dd className="text-gray-900">{formatReleasedOn(animeWithDetails.released_on)}</dd>
                </div>
              )}
              {animeWithDetails.media && (
                <div className="flex">
                  <dt className="text-gray-500 w-24 flex-shrink-0">メディア</dt>
                  <dd className="text-gray-900">{getMediaLabel(animeWithDetails.media)}</dd>
                </div>
              )}
              {/* 話数 - 映画の場合は非表示（映画にエピソードの概念がないため） */}
              {!isMovie && (animeWithDetails.episodes_count || episodes.length > 0) && (
                <div className="flex">
                  <dt className="text-gray-500 w-24 flex-shrink-0">話数</dt>
                  <dd className="text-gray-900">{animeWithDetails.episodes_count || episodes.length}話</dd>
                </div>
              )}
              {animeWithDetails.official_site_url && (
                <div className="flex sm:col-span-2">
                  <dt className="text-gray-500 w-24 flex-shrink-0">公式サイト</dt>
                  <dd>
                    <a
                      href={animeWithDetails.official_site_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-blue-600 hover:text-blue-800 hover:underline truncate block max-w-xs"
                    >
                      {(() => {
                        try {
                          return new URL(animeWithDetails.official_site_url).hostname
                        } catch {
                          return animeWithDetails.official_site_url
                        }
                      })()}
                    </a>
                  </dd>
                </div>
              )}
            </dl>
          </div>
        )}

        {/* 配信サービスセクション */}
        {animeWithDetails.watch_providers && animeWithDetails.watch_providers.length > 0 && (
          <WatchProviders providers={animeWithDetails.watch_providers} />
        )}

        {/* エピソード一覧セクション - 映画の場合は非表示（映画にエピソードの概念がないため） */}
        {!isMovie && (
          <div className="bg-white rounded-lg border overflow-hidden">
            <h2 className="text-lg font-semibold text-gray-900 px-4 md:px-6 py-4">
              エピソード一覧
            </h2>
            
            <div className="px-4 md:px-6 pb-4 md:pb-6 border-t">
              <div className="pt-4">
                <EpisodeList
                  episodes={episodes}
                  isLoading={isLoadingEpisodes}
                />
              </div>
            </div>
          </div>
        )}
      </div>
    </Container>
  )
}

export default AnimeDetailClient
