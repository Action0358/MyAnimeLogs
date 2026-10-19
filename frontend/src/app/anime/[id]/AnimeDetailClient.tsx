'use client'

import React, { useState, useEffect, useRef } from 'react'
import { useParams, usePathname } from 'next/navigation'
import Link from 'next/link'
import type { Route } from 'next'
import { useQuery } from '@tanstack/react-query'
import { decode as heDecodeEntities } from 'he'
import { Container } from '@/components/layout'
import { ChevronDown, ChevronUp } from 'lucide-react'
import {
  Button,
  Card,
  CardContent,
  ErrorState,
  SectionHeading,
  getErrorStateType,
  isNotFoundError,
} from '@/components/ui'
import { WatchProviders, RelatedAnimeGrid, PosterFallback } from '@/components/anime'
import type { Anime } from '@/lib/api/generated/myanimelogs'
import { FavoriteButton, StatusSelector } from '@/components/favorites'
import { getAnimesId } from '@/lib/api/generated/myanimelogs'
import { useFavorites } from '@/hooks/useFavorites'
import { useWatchStatus } from '@/hooks/useWatchStatus'
import { useAuth } from '@/contexts/AuthContext'
import { SignInLink } from '@/components/auth/SignInLink'
import { cn } from '@/lib/utils'
import { getAnimeStatusLabel, type BroadcastStatusLabel } from '@/utils/animeStatus'
import { formatSeason } from '@/utils/formatSeason'
import { resolveBackTarget, type BackTarget } from '@/lib/animeListOrigin'

/** あらすじの既定表示行数。作品ごとの文字数差でカード高さが大きくぶれるため折りたたむ。
 *  Tailwind のJITは動的なクラス名を検出できないためリテラルで指定する */
const SYNOPSIS_CLAMP_CLASS = 'line-clamp-4'

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
          // 14px(text-sm) × 1.75(leading) × 4行 = 98px
          !isExpanded && `${SYNOPSIS_CLAMP_CLASS} min-h-[98px]`,
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
  // useSearchParams は静的エクスポート時に Suspense 境界を要求し、
  // SPAフォールバックで描画されるこのページとは噛み合わないため、
  // animeId と同様にマウント後に window から解決する。
  const [backTarget, setBackTarget] = useState<BackTarget>(() => resolveBackTarget(null))
  // 関連作品へ渡す戻り先（この作品の詳細ページ。この作品自身の戻り先を含む）
  const [relatedFrom, setRelatedFrom] = useState<string | undefined>(undefined)
  const pathname = usePathname()
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
      const parsed = parseInt(params.id, 10)
      // 数値でないIDは NaN のままだとスケルトンから進まないため、不正なID（0）として扱う
      setAnimeId(Number.isNaN(parsed) ? 0 : parsed)
    }
  }, [params.id])

  // 戻り先も同様にURLから解決する。関連作品から別の作品へ移ったときに
  // 解決し直せるよう、パスの変化に追従する
  useEffect(() => {
    const from = new URLSearchParams(window.location.search).get('from')
    setBackTarget(resolveBackTarget(from))
    setRelatedFrom(window.location.pathname + window.location.search)
  }, [pathname])

  const isValidAnimeId = !isNaN(animeId) && animeId > 0

  const {
    data: animeData,
    isLoading: isLoadingAnime,
    error: animeError,
    refetch: refetchAnime,
  } = useQuery({
    queryKey: ['anime', animeId],
    queryFn: async () => {
      const response = await getAnimesId(animeId)
      return response.data
    },
    enabled: isValidAnimeId,
    // 存在しない作品（404）は再試行しても変わらないため、すぐにエラー画面を出す
    retry: (failureCount, error) => !isNotFoundError(error) && failureCount < 3,
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

  // 不正なIDと、APIが404を返した場合は「作品が見つかりません」。再読み込みでは直らない
  if (!isValidAnimeId || isNotFoundError(animeError) || (!animeError && !animeData)) {
    return (
      <Container size="xl" className="py-8">
        <ErrorState type="notFound" />
      </Container>
    )
  }

  // 404以外の取得失敗（通信エラー等）。存在しない作品と取り違えないよう、種類で文言を分ける
  if (animeError || !animeData) {
    return (
      <Container size="xl" className="py-8">
        <ErrorState type={getErrorStateType(animeError)} onRetry={() => refetchAnime()} />
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

  // 放送状況・メディア種別・放送時期はタイトル直下のバッジにまとめ、
  // 下の作品情報表からは外す（アニメ詳細UI刷新v2）
  const badges = [statusInfo.label, getMediaLabel(anime.media), formatSeason(anime.season)].filter(
    (b): b is string => Boolean(b)
  )

  const badgesEl = badges.length > 0 && (
    <div className="flex flex-wrap gap-1.5">
      {badges.map((b) => (
        <span
          key={b}
          className="inline-flex items-center rounded-md bg-muted px-2.5 py-1 text-xs font-semibold text-foreground/70"
        >
          {b}
        </span>
      ))}
    </div>
  )

  const posterEl = (
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
        <PosterFallback />
      )}
    </div>
  )

  // ログイン後は「視聴ステータス」を主ボタン（塗り）、お気に入りを48px角の
  // アイコンボタンにする。ログイン前のCTAと高さ48pxを揃え、状態が変わっても
  // 操作の重さが揃って見えるようにする（アニメ詳細UI刷新v2）
  const actionsEl = isAuthenticated ? (
    <div className="flex items-stretch gap-2">
      <div className="flex-1 min-w-0 h-12">
        <StatusSelector
          currentStatus={getStatus(animeId)}
          onStatusChange={(status) => updateStatus(animeId, status)}
          onRemove={() => removeStatus(animeId)}
          size="lg"
          variant="primary"
          className="w-full h-full"
        />
      </div>
      <FavoriteButton
        isFavorite={isFavorite(animeId)}
        onToggle={() =>
          isFavorite(animeId) ? removeFavorite(animeId) : addFavorite(animeId)
        }
        variant="icon"
        size="lg"
        className="flex-shrink-0"
      />
    </div>
  ) : (
    <div>
      {/* ログイン後にこの作品へ戻ってきて、そのまま記録を始められるようにする */}
      <SignInLink className="block">
        <Button className="w-full h-12">ログインして記録する</Button>
      </SignInLink>
      <p className="mt-2.5 text-center text-xs text-muted-foreground">
        視聴状況とお気に入りを保存できます
      </p>
    </div>
  )

  const synopsisCard = anime.description && (
    <Card>
      <CardContent className={DETAIL_CARD_PADDING}>
        <SectionHeading as="h2" size="sm">あらすじ</SectionHeading>
        <ExpandableSynopsis text={stripHtmlTags(anime.description)} />
      </CardContent>
    </Card>
  )

  // 作品情報は放送開始・公式サイト・ジャンルに絞る（放送状況・メディア・放送時期は
  // 上のバッジへ移動済み。アニメ詳細UI刷新v2）
  const infoCard = (anime.released_on || anime.official_site_url || (anime.genres && anime.genres.length > 0)) && (
    <Card>
      <CardContent className={DETAIL_CARD_PADDING}>
        <SectionHeading as="h2" size="sm">作品情報</SectionHeading>
        <dl className="grid grid-cols-[76px_1fr] md:grid-cols-[84px_1fr] gap-x-4 gap-y-3 text-sm">
          {anime.released_on && (
            <>
              <dt className="text-muted-foreground">放送開始</dt>
              <dd className="text-foreground font-medium">{formatReleasedOn(anime.released_on)}</dd>
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
              <dt className="text-muted-foreground self-start pt-0.5">ジャンル</dt>
              <dd className="flex flex-wrap gap-1.5">
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
  )

  const providersEl = anime.watch_providers && anime.watch_providers.length > 0 && (
    <WatchProviders providers={anime.watch_providers} />
  )

  const relatedSection = anime.related_seasons && anime.related_seasons.length > 0 && (
    <div className="mt-7 pt-[22px] lg:mt-14 lg:pt-9 border-t border-border">
      <div className="mb-3.5 lg:mb-5 flex items-baseline gap-2.5 lg:gap-3">
        <h2 className="text-base lg:text-xl font-bold tracking-tight text-foreground">関連作品</h2>
        <span className="text-xs lg:text-sm text-muted-foreground">{anime.related_seasons.length}作品</span>
      </div>
      <RelatedAnimeGrid related={anime.related_seasons} from={relatedFrom} />
    </div>
  )

  return (
    <Container size="xl" className="py-8">
      {/* 戻る導線。遷移元は from クエリで受け取る（関連作品からの遷移は「前の作品」）。
          既知の画面でなければホームへ */}
      <div className="mb-6">
        <Link
          href={backTarget.href as Route}
          className="inline-flex items-center gap-1.5 -ml-1 px-1 py-1 text-sm text-muted-foreground hover:text-primary transition-colors rounded focus:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          <svg
            className="w-4 h-4 flex-shrink-0"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
          <span>{backTarget.label}に戻る</span>
        </Link>
      </div>

      {/* SP: ポスター（左・40%）＋タイトル・バッジ（右）の横並び。
          最初の画面でタイトルと操作が見えるようにする（アニメ詳細UI刷新v2） */}
      <div className="lg:hidden">
        <div className="flex gap-3.5 items-start">
          <div className="w-[40%] flex-shrink-0">{posterEl}</div>
          <div className="flex-1 min-w-0 pt-0.5">
            <h1 className="text-lg font-bold text-foreground tracking-tight leading-snug">
              {anime.title}
            </h1>
            {badgesEl && <div className="mt-2.5">{badgesEl}</div>}
          </div>
        </div>

        <div className="mt-4">{actionsEl}</div>

        <div className="flex flex-col gap-4 mt-5">
          {synopsisCard}
          {infoCard}
          {providersEl}
        </div>
      </div>

      {/* PC: 2カラム（左=ポスターと操作 / 右=タイトル・情報カード）。
          コンテンツ列に上限を設けないと情報が横に間延びするため minmax で 720px に制限する */}
      <div className="hidden lg:grid lg:grid-cols-[405px_minmax(0,720px)] gap-10 items-start">
        <div>
          {posterEl}
          <div className="mt-5">{actionsEl}</div>
        </div>

        <div className="min-w-0">
          <h1 className="text-[28px] font-bold text-foreground tracking-tight leading-snug">
            {anime.title}
          </h1>
          {badgesEl && <div className="mt-3 mb-7">{badgesEl}</div>}

          <div className="flex flex-col gap-6">
            {synopsisCard}
            {infoCard}
            {providersEl}
          </div>
        </div>
      </div>

      {relatedSection}
    </Container>
  )
}

export default AnimeDetailClient
