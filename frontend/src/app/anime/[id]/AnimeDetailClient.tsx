'use client'

import React, { useState, useEffect } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { useQuery } from '@tanstack/react-query'
import { decode as heDecodeEntities } from 'he'
import { Container } from '@/components/layout'
import { Button } from '@/components/ui'
import { WatchProviders } from '@/components/anime'
import { FavoriteButton, StatusSelector } from '@/components/favorites'
import { getAnimesId } from '@/lib/api/generated/myanimelogs'
import { useFavorites } from '@/hooks/useFavorites'
import { useWatchStatus } from '@/hooks/useWatchStatus'
import { useAuth } from '@/contexts/AuthContext'
import { cn } from '@/lib/utils'
import { getAnimeStatusLabel, type BroadcastStatusLabel } from '@/utils/animeStatus'
import { translateBroadcastDay } from '@/utils/broadcast'
import { formatDuration } from '@/utils/duration'

const AnimeDetailClient: React.FC = () => {
  const params = useParams()
  const { isAuthenticated } = useAuth()
  const { isFavorite, addFavorite, removeFavorite } = useFavorites({ enabled: isAuthenticated })
  const { getStatus, updateStatus, removeStatus } = useWatchStatus({ enabled: isAuthenticated })
  const [imageError, setImageError] = useState(false)
  const [animeId, setAnimeId] = useState<number>(NaN)
  // ハイドレーションミスマッチを回避するため、クライアントサイドでのみ放送状態を計算
  const [statusInfo, setStatusInfo] = useState<BroadcastStatusLabel>({ label: null, className: '' })
  
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

  // クライアントサイドでのみ放送状態を計算（ハイドレーションミスマッチ回避）
  // 注意: フックは条件付きreturnの前に配置する必要がある（Reactのルール）
  useEffect(() => {
    if (animeData) {
      setStatusInfo(getAnimeStatusLabel(animeData.status, animeData.released_on))
    }
  }, [animeData])

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
              {statusInfo.label && (
                <span
                  className={cn(
                    'inline-flex items-center px-2.5 py-1 rounded-full text-sm font-medium',
                    statusInfo.className
                  )}
                >
                  {statusInfo.label}
                </span>
              )}
            </div>

            {/* ジャンルタグ */}
            {anime.genres && anime.genres.length > 0 && (
              <div className="flex flex-wrap gap-1.5 mb-4">
                {anime.genres.map((genre) => (
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

            {/* アクションボタン：お気に入り（ハート）と視聴ステータスは独立 */}
            <div className="flex flex-wrap items-start gap-2">
              {isAuthenticated ? (
                <>
                  <FavoriteButton
                    isFavorite={isFavorite(animeId)}
                    onToggle={() =>
                      isFavorite(animeId)
                        ? removeFavorite(animeId)
                        : addFavorite(animeId)
                    }
                    size="sm"
                  />
                  <StatusSelector
                    currentStatus={getStatus(animeId)}
                    onStatusChange={(status) => updateStatus(animeId, status)}
                    onRemove={() => removeStatus(animeId)}
                    size="sm"
                  />
                </>
              ) : (
                <Link href="/login">
                  <Button size="sm">ログインしてお気に入り・視聴管理</Button>
                </Link>
              )}
            </div>
          </div>
        </div>

        {/* あらすじセクション */}
        {anime.description && (
          <div className="mb-6 bg-white rounded-lg border p-4 md:p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-3">あらすじ</h2>
            <p className="text-gray-700 leading-relaxed whitespace-pre-wrap text-sm md:text-base">
              {stripHtmlTags(anime.description)}
            </p>
          </div>
        )}

        {/* 作品情報セクション */}
        {(anime.released_on || anime.media || anime.official_site_url) && (
          <div className="mb-6 bg-white rounded-lg border p-4 md:p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-3">作品情報</h2>
            <dl className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-sm">
              {anime.released_on && (
                <div className="flex">
                  <dt className="text-gray-500 w-24 flex-shrink-0">放送開始</dt>
                  <dd className="text-gray-900">{formatReleasedOn(anime.released_on)}</dd>
                </div>
              )}
              {anime.broadcast_day && (
                <div className="flex">
                  <dt className="text-gray-500 w-24 flex-shrink-0">放送日</dt>
                  <dd className="text-gray-900">{translateBroadcastDay(anime.broadcast_day)}</dd>
                </div>
              )}
              {anime.broadcast_time && (
                <div className="flex">
                  <dt className="text-gray-500 w-24 flex-shrink-0">放送時間</dt>
                  <dd className="text-gray-900">{anime.broadcast_time}</dd>
                </div>
              )}
              {/* 映画の上映時間 */}
              {anime.media === 'movie' && anime.duration_minutes && (
                <div className="flex">
                  <dt className="text-gray-500 w-24 flex-shrink-0">上映時間</dt>
                  <dd className="text-gray-900">{formatDuration(anime.duration_minutes)}</dd>
                </div>
              )}
              {anime.media && (
                <div className="flex">
                  <dt className="text-gray-500 w-24 flex-shrink-0">メディア</dt>
                  <dd className="text-gray-900">{getMediaLabel(anime.media)}</dd>
                </div>
              )}
              {anime.official_site_url && (
                <div className="flex sm:col-span-2">
                  <dt className="text-gray-500 w-24 flex-shrink-0">公式サイト</dt>
                  <dd>
                    <a
                      href={anime.official_site_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-blue-600 hover:text-blue-800 hover:underline truncate block max-w-xs"
                    >
                      {(() => {
                        try {
                          return new URL(anime.official_site_url).hostname
                        } catch {
                          return anime.official_site_url
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
        {anime.watch_providers && anime.watch_providers.length > 0 && (
          <WatchProviders providers={anime.watch_providers} />
        )}

      </div>
    </Container>
  )
}

export default AnimeDetailClient
