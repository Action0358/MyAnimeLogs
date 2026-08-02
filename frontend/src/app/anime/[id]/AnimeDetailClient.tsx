'use client'

import React, { useState, useEffect } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { useQuery } from '@tanstack/react-query'
import { decode as heDecodeEntities } from 'he'
import { Container } from '@/components/layout'
import { Badge, Button, Card, CardContent, SectionHeading } from '@/components/ui'
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
      <Container size="xl" className="py-8">
        <div className="animate-pulse">
          <div className="h-4 bg-muted rounded w-48 mb-6" />
          <div className="grid grid-cols-1 lg:grid-cols-[300px_1fr] gap-6 lg:gap-10 items-start">
            <div className="space-y-5">
              <div className="w-40 sm:w-48 lg:w-full aspect-[3/4] bg-muted rounded-lg mx-auto sm:mx-0" />
              <div className="h-9 bg-muted rounded" />
            </div>
            <div className="space-y-6">
              <div className="space-y-3">
                <div className="h-8 bg-muted rounded w-3/4" />
                <div className="h-5 bg-muted rounded w-1/3" />
              </div>
              <div className="h-32 bg-muted rounded-lg" />
              <div className="h-24 bg-muted rounded-lg" />
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
          <h1 className="text-2xl font-bold text-foreground mb-2">無効なアニメIDです</h1>
          <p className="text-muted-foreground mb-6">正しいURLでアクセスしてください。</p>
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
          <h1 className="text-2xl font-bold text-foreground mb-2">アニメが見つかりません</h1>
          <p className="text-muted-foreground mb-6">お探しのアニメは存在しないか、削除された可能性があります。</p>
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
    <Container size="xl" className="py-8">
      {/* パンくずリスト */}
      <nav aria-label="パンくずリスト" className="mb-6 text-sm">
        <ol className="flex items-center gap-2 text-muted-foreground">
          <li>
            <Link href="/" className="hover:text-primary transition-colors">ホーム</Link>
          </li>
          <li aria-hidden="true">/</li>
          <li>
            <Link href="/search" className="hover:text-primary transition-colors">検索</Link>
          </li>
          <li aria-hidden="true">/</li>
          <li className="text-foreground font-medium truncate max-w-[200px]" aria-current="page">
            {anime.title}
          </li>
        </ol>
      </nav>

      {/* 2カラム: 左=カバー画像と操作 / 右=テキスト情報 */}
      <div className="grid grid-cols-1 lg:grid-cols-[300px_1fr] gap-6 lg:gap-10 items-start">
        {/* 左カラム（PCでは追従表示） */}
        <div className="lg:sticky lg:top-6 space-y-5">
          <div className="flex flex-col sm:flex-row lg:flex-col gap-5">
            {/* カバー画像 */}
            <div className="w-40 sm:w-48 lg:w-full flex-shrink-0 mx-auto sm:mx-0">
              <div className="relative aspect-[3/4] bg-muted rounded-lg overflow-hidden shadow-lg ring-1 ring-border">
                {imageUrl && !imageError ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img
                    src={imageUrl}
                    alt={anime.title || 'アニメ画像'}
                    className="w-full h-full object-cover"
                    onError={() => setImageError(true)}
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center bg-gradient-to-br from-primary/15 to-primary/5">
                    <span className="text-5xl">📺</span>
                  </div>
                )}
              </div>
            </div>

            {/* お気に入り・視聴ステータス操作 */}
            <div className="flex-1 min-w-0 flex flex-col justify-center lg:justify-start gap-2">
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
                  <Button size="sm" className="w-full">ログインしてお気に入り・視聴管理</Button>
                </Link>
              )}
            </div>
          </div>

          {/* 配信サービス */}
          {anime.watch_providers && anime.watch_providers.length > 0 && (
            <WatchProviders providers={anime.watch_providers} />
          )}
        </div>

        {/* 右カラム */}
        <div className="min-w-0 space-y-6">
          {/* タイトルとメタ情報 */}
          <div>
            <h1 className="text-2xl md:text-3xl font-bold text-foreground tracking-tight mb-3">
              {anime.title}
            </h1>

            <div className="flex flex-wrap items-center gap-2">
              {anime.season && <Badge variant="secondary">{anime.season}</Badge>}
              {statusInfo.label && (
                <span
                  className={cn(
                    'inline-flex items-center px-2.5 py-0.5 rounded-full border border-transparent text-xs font-semibold',
                    statusInfo.className
                  )}
                >
                  {statusInfo.label}
                </span>
              )}
              {anime.media && <Badge variant="outline">{getMediaLabel(anime.media)}</Badge>}
            </div>

            {/* ジャンルタグ */}
            {anime.genres && anime.genres.length > 0 && (
              <div className="flex flex-wrap gap-1.5 mt-3">
                {anime.genres.map((genre) => (
                  <Link
                    key={genre}
                    href={`/search?genre=${encodeURIComponent(genre)}`}
                    className="inline-flex items-center px-2.5 py-0.5 rounded-full border border-border text-xs font-medium text-muted-foreground hover:border-primary hover:text-primary transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
                  >
                    {genre}
                  </Link>
                ))}
              </div>
            )}
          </div>

          {/* あらすじ */}
          {anime.description && (
            <Card>
              <CardContent className="p-4 md:p-6">
                <SectionHeading as="h2" size="sm">あらすじ</SectionHeading>
                <p className="text-foreground/80 leading-relaxed whitespace-pre-wrap text-sm md:text-base">
                  {stripHtmlTags(anime.description)}
                </p>
              </CardContent>
            </Card>
          )}

          {/* 作品情報 */}
          {(anime.released_on || anime.media || anime.official_site_url) && (
            <Card>
              <CardContent className="p-4 md:p-6">
                <SectionHeading as="h2" size="sm">作品情報</SectionHeading>
                <dl className="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-3 text-sm">
                  {anime.released_on && (
                    <div className="flex">
                      <dt className="text-muted-foreground w-24 flex-shrink-0">放送開始</dt>
                      <dd className="text-foreground">{formatReleasedOn(anime.released_on)}</dd>
                    </div>
                  )}
                  {anime.broadcast_day && (
                    <div className="flex">
                      <dt className="text-muted-foreground w-24 flex-shrink-0">放送日</dt>
                      <dd className="text-foreground">{translateBroadcastDay(anime.broadcast_day)}</dd>
                    </div>
                  )}
                  {anime.broadcast_time && (
                    <div className="flex">
                      <dt className="text-muted-foreground w-24 flex-shrink-0">放送時間</dt>
                      <dd className="text-foreground">{anime.broadcast_time}</dd>
                    </div>
                  )}
                  {/* 映画の上映時間 */}
                  {anime.media === 'movie' && anime.duration_minutes && (
                    <div className="flex">
                      <dt className="text-muted-foreground w-24 flex-shrink-0">上映時間</dt>
                      <dd className="text-foreground">{formatDuration(anime.duration_minutes)}</dd>
                    </div>
                  )}
                  {anime.media && (
                    <div className="flex">
                      <dt className="text-muted-foreground w-24 flex-shrink-0">メディア</dt>
                      <dd className="text-foreground">{getMediaLabel(anime.media)}</dd>
                    </div>
                  )}
                  {anime.official_site_url && (
                    <div className="flex sm:col-span-2">
                      <dt className="text-muted-foreground w-24 flex-shrink-0">公式サイト</dt>
                      <dd className="min-w-0">
                        <a
                          href={anime.official_site_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-primary hover:underline truncate block"
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
              </CardContent>
            </Card>
          )}
        </div>
      </div>
    </Container>
  )
}

export default AnimeDetailClient
