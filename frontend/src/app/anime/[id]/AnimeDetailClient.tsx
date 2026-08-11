'use client'

import React, { useState, useEffect, useRef } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { useQuery } from '@tanstack/react-query'
import { decode as heDecodeEntities } from 'he'
import { Container } from '@/components/layout'
import { ChevronDown, ChevronUp } from 'lucide-react'
import { Button, Card, CardContent, SectionHeading } from '@/components/ui'
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

/** あらすじの既定表示行数。作品ごとの文字数差でカード高さが大きくぶれるため折りたたむ。
 *  Tailwind のJITは動的なクラス名を検出できないためリテラルで指定する */
const SYNOPSIS_CLAMP_CLASS = 'line-clamp-5'

/**
 * 既定では指定行数で折りたたみ、収まらない場合のみ展開ボタンを出す本文。
 * 収まる作品ではボタン自体が出ないので、短いあらすじの見た目は変わらない。
 */
const ExpandableSynopsis: React.FC<{ text: string }> = ({ text }) => {
  const textRef = useRef<HTMLParagraphElement>(null)
  const [isExpanded, setIsExpanded] = useState(false)
  const [isOverflowing, setIsOverflowing] = useState(false)

  useEffect(() => {
    // 展開中は clientHeight と scrollHeight が一致するため判定しない
    // （判定するとボタンが消えて折りたたみに戻せなくなる）
    if (isExpanded) return
    const el = textRef.current
    if (!el) return

    const check = () => setIsOverflowing(el.scrollHeight > el.clientHeight + 1)
    check()

    // フォント読み込みや幅変更で行数が変わるため追従する
    const observer = new ResizeObserver(check)
    observer.observe(el)
    return () => observer.disconnect()
  }, [text, isExpanded])

  return (
    <div>
      <p
        ref={textRef}
        className={cn(
          'text-foreground/80 leading-[1.75] whitespace-pre-wrap text-sm',
          // 折りたたみ時は既定行数ぶんの高さを確保し、作品ごとの高さのぶれを抑える
          // 14px(text-sm) × 1.75(leading) × 5行 = 122.5px
          !isExpanded && `${SYNOPSIS_CLAMP_CLASS} min-h-[122.5px]`,
        )}
      >
        {text}
      </p>
      {isOverflowing && (
        <button
          type="button"
          onClick={() => setIsExpanded((prev) => !prev)}
          aria-expanded={isExpanded}
          className="mt-3 inline-flex items-center gap-1 text-sm font-medium text-primary hover:underline focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 rounded"
        >
          {isExpanded ? (
            <>折りたたむ<ChevronUp className="w-4 h-4" /></>
          ) : (
            <>続きを読む<ChevronDown className="w-4 h-4" /></>
          )}
        </button>
      )}
    </div>
  )
}

/** 情報カードの内側余白（デザイン仕様: 24px / 28px） */
const DETAIL_CARD_PADDING = 'p-5 md:px-6 md:py-5'


const AnimeDetailClient: React.FC = () => {
  const params = useParams()
  const { isAuthenticated } = useAuth()
  const { isFavorite, addFavorite, removeFavorite } = useFavorites({ enabled: isAuthenticated })
  const { getStatus, updateStatus, removeStatus } = useWatchStatus({ enabled: isAuthenticated })
  const [imageError, setImageError] = useState(false)
  const [animeId, setAnimeId] = useState<number>(NaN)
  // ハイドレーションミスマッチを回避するため、クライアントサイドでのみ放送状況を計算
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


  // クライアントサイドでのみ放送状況を計算（ハイドレーションミスマッチ回避）
  // 注意: フックは条件付きreturnの前に配置する必要がある（Reactのルール）
  useEffect(() => {
    if (animeData) {
      setStatusInfo(getAnimeStatusLabel(animeData.status, animeData.released_on))
    }
  }, [animeData])

  // ローディング表示
  if (isNaN(animeId) || isLoadingAnime) {
    return (
      <Container size="xl" className="max-w-[1180px] py-8">
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
    <Container size="xl" className="max-w-[1180px] py-8">
      {/* パンくずリスト */}
      <nav aria-label="パンくずリスト" className="mb-6 text-sm">
        <ol className="flex items-center gap-2 text-muted-foreground whitespace-nowrap">
          <li className="flex-shrink-0">
            <Link href="/" className="hover:text-primary transition-colors">ホーム</Link>
          </li>
          <li aria-hidden="true" className="flex-shrink-0">/</li>
          <li className="flex-shrink-0">
            <Link href="/search" className="hover:text-primary transition-colors">検索</Link>
          </li>
          <li aria-hidden="true" className="flex-shrink-0">/</li>
          <li className="text-foreground font-medium truncate min-w-0" aria-current="page">
            {anime.title}
          </li>
        </ol>
      </nav>

      {/* 2カラム: 左=ポスターと操作 / 右=情報カード
          コンテンツ列に上限を設けないと情報が横に間延びするため minmax で 720px に制限する */}
      <div className="grid grid-cols-1 lg:grid-cols-[405px_minmax(0,720px)] gap-8 lg:gap-10 items-start">
        {/* 左カラム */}
        <div>
          {/* ポスター画像。余白の中で主役になるようドロップシャドウを添える */}
          <div className="relative aspect-[2/3] w-full bg-muted rounded-lg overflow-hidden shadow-[0_12px_32px_-12px_rgba(15,23,42,0.25)]">
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

          {/* お気に入り・視聴ステータス操作。
              2つを横並びにし、flex-1 で等分してポスターの横幅にちょうど収める。
              高さは h-11 で揃える（StatusSelector は内部の button が h-full で追従） */}
          <div className="mt-4 lg:mt-5 flex items-stretch gap-2">
            {isAuthenticated ? (
              <>
                {/* flex-1 は padding/border を基準サイズに加算するため、
                    button を直接 flex アイテムにすると StatusSelector 側と幅がずれる。
                    余白を持たない div で包んで両者を対称にする */}
                <div className="flex-1 min-w-0 h-11">
                  <FavoriteButton
                    isFavorite={isFavorite(animeId)}
                    onToggle={() =>
                      isFavorite(animeId)
                        ? removeFavorite(animeId)
                        : addFavorite(animeId)
                    }
                    size="sm"
                    className="w-full h-full justify-center"
                  />
                </div>
                <StatusSelector
                  currentStatus={getStatus(animeId)}
                  onStatusChange={(status) => updateStatus(animeId, status)}
                  onRemove={() => removeStatus(animeId)}
                  size="sm"
                  className="flex-1 min-w-0 h-11"
                />
              </>
            ) : (
              <Link href="/login" className="block flex-1">
                <Button className="w-full h-12">ログインしてお気に入り・視聴管理</Button>
              </Link>
            )}
          </div>
        </div>

        {/* 右カラム */}
        <div className="min-w-0">
          {/* タイトルと属性 */}
          <h1 className="text-xl md:text-[28px] font-bold text-foreground tracking-tight leading-snug mb-7">
            {anime.title}
          </h1>




          {/* 情報カード群。カード間の余白は 24px で統一する */}
          <div className="flex flex-col gap-6">
            {/* あらすじ */}
            {anime.description && (
              <Card>
                <CardContent className={DETAIL_CARD_PADDING}>
                  <SectionHeading as="h2" size="sm">あらすじ</SectionHeading>
                  <ExpandableSynopsis text={stripHtmlTags(anime.description)} />
                </CardContent>
              </Card>
            )}

            {/* 作品情報。ラベルと値を4列グリッドに並べ、右側の余白を作らない */}
            {(statusInfo.label || anime.released_on || anime.media || anime.official_site_url || (anime.genres && anime.genres.length > 0)) && (
              <Card>
                <CardContent className={DETAIL_CARD_PADDING}>
                  <SectionHeading as="h2" size="sm">作品情報</SectionHeading>
                  <dl className="grid grid-cols-[76px_1fr] md:grid-cols-[88px_1fr_88px_1fr] gap-x-4 gap-y-3.5 text-sm">
                    {statusInfo.label && (
                      <>
                        <dt className="text-muted-foreground">放送状況</dt>
                        <dd className="text-foreground font-medium">{statusInfo.label}</dd>
                      </>
                    )}
                    {anime.released_on && (
                      <>
                        <dt className="text-muted-foreground">放送開始</dt>
                        <dd className="text-foreground font-medium">{formatReleasedOn(anime.released_on)}</dd>
                      </>
                    )}
                    {anime.broadcast_day && (
                      <>
                        <dt className="text-muted-foreground">放送日</dt>
                        <dd className="text-foreground font-medium">{translateBroadcastDay(anime.broadcast_day)}</dd>
                      </>
                    )}
                    {anime.broadcast_time && (
                      <>
                        <dt className="text-muted-foreground">放送時間</dt>
                        <dd className="text-foreground font-medium">{anime.broadcast_time}</dd>
                      </>
                    )}
                    {/* 映画の上映時間 */}
                    {anime.media === 'movie' && anime.duration_minutes && (
                      <>
                        <dt className="text-muted-foreground">上映時間</dt>
                        <dd className="text-foreground font-medium">{formatDuration(anime.duration_minutes)}</dd>
                      </>
                    )}
                    {anime.media && (
                      <>
                        <dt className="text-muted-foreground">メディア</dt>
                        <dd className="text-foreground font-medium">{getMediaLabel(anime.media)}</dd>
                      </>
                    )}
                    {anime.official_site_url && (
                      <>
                        <dt className="text-muted-foreground">公式サイト</dt>
                        <dd className="min-w-0">
                          <a
                            href={anime.official_site_url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-primary font-medium hover:underline truncate block"
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
                      </>
                    )}
                    {anime.genres && anime.genres.length > 0 && (
                      <>
                        <dt className="text-muted-foreground">ジャンル</dt>
                        <dd className="md:col-span-3 flex flex-wrap gap-1.5">
                          {anime.genres.map((genre) => (
                            <Link
                              key={genre}
                              href={`/search?genre=${encodeURIComponent(genre)}`}
                              className="inline-flex items-center px-2.5 py-0.5 rounded-full border border-border text-xs font-medium text-foreground/70 hover:border-primary hover:text-primary hover:bg-primary/5 transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
                            >
                              {genre}
                            </Link>
                          ))}
                        </dd>
                      </>
                    )}
                  </dl>
                </CardContent>
              </Card>
            )}

            {/* 配信サービス */}
            {anime.watch_providers && anime.watch_providers.length > 0 && (
              <WatchProviders providers={anime.watch_providers} />
            )}
          </div>
        </div>
      </div>
    </Container>
  )
}

export default AnimeDetailClient
