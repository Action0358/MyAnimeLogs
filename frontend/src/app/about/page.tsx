'use client'

import React from 'react'
import Link from 'next/link'
import { useQuery } from '@tanstack/react-query'
import {
  CalendarDays,
  Heart,
  ListChecks,
  Search,
  Tv,
  TrendingUp,
  type LucideIcon,
} from 'lucide-react'
import { Container } from '@/components/layout'
import { useAuth } from '@/contexts/AuthContext'
import { currentSeasonAnimeQuery } from '@/lib/queries/currentSeasonAnime'

interface Feature {
  icon: LucideIcon
  title: string
  desc: string
}

const FEATURES: Feature[] = [
  {
    icon: ListChecks,
    title: '視聴ステータス管理',
    desc: '視聴予定・視聴中・視聴完了の3段階で作品を整理できます。',
  },
  {
    icon: Heart,
    title: 'お気に入り登録',
    desc: '気になる作品を保存して、いつでも見返せます。',
  },
  {
    icon: CalendarDays,
    title: '今期アニメ',
    desc: '放送中の作品を放送状態つきでまとめてチェックできます。',
  },
  {
    icon: TrendingUp,
    title: '人気アニメ',
    desc: 'いま多く記録されている作品をランキングで確認できます。',
  },
  {
    icon: Search,
    title: '条件で絞り込める検索',
    desc: 'キーワード・放送時期・ジャンルから作品を探せます。',
  },
  {
    icon: Tv,
    title: '配信サービス表示',
    desc: '作品ごとに、どのサービスで観られるかがわかります。',
  },
]

interface Step {
  n: number
  title: string
  desc: string
}

const STEPS: Step[] = [
  { n: 1, title: 'ログイン', desc: 'Googleアカウントですぐに始められます。' },
  { n: 2, title: '作品を探す', desc: '今期アニメや検索から気になる作品を見つけます。' },
  { n: 3, title: '記録する', desc: 'お気に入り登録と視聴ステータスを設定します。' },
  { n: 4, title: 'ライブラリを見る', desc: 'マイページで自分の記録をいつでも確認できます。' },
]

interface DataSource {
  kind: string
  name: string
  href: string
  note?: string
}

const DATA_SOURCES: DataSource[] = [
  { kind: 'アニメ基本情報', name: 'Annict', href: 'https://annict.com' },
  { kind: 'カバー画像・ジャンル', name: 'AniList', href: 'https://anilist.co' },
  { kind: '放送情報', name: 'MyAnimeList', href: 'https://myanimelist.net' },
  { kind: '配信情報', name: 'TMDB / JustWatch', href: 'https://www.themoviedb.org/' },
]

/** ヒーローの背景に敷くポスター枚数（レスポンシブでグリッド列数を変える） */
const HERO_POSTER_COUNT = 8

/**
 * ヒーロー背景。今期アニメのカバー画像をタイル状に敷く。
 * クエリ定義はホームの HeroBackdrop と共有しているため、追加のAPIリクエストは発生しない。
 */
const AboutHero: React.FC = () => {
  const { data } = useQuery(currentSeasonAnimeQuery)

  const posters = (data?.animes ?? [])
    .map((anime) => anime.cover_image_large || anime.cover_image_medium || anime.image_url)
    .filter((url): url is string => Boolean(url))
    .slice(0, HERO_POSTER_COUNT)

  const tiles = posters.length > 0 ? posters : Array.from<string | null>({ length: HERO_POSTER_COUNT }).fill(null)

  return (
    <div className="relative h-[220px] overflow-hidden bg-foreground sm:h-[300px] lg:h-[340px]">
      <div className="absolute inset-0 grid grid-cols-4 gap-2 p-3 sm:grid-cols-6 sm:gap-2.5 sm:p-5 lg:grid-cols-8">
        {tiles.map((url, i) => (
          <div key={i} className="aspect-[2/3] overflow-hidden rounded-md bg-background/10 sm:rounded-lg">
            {url && (
              // eslint-disable-next-line @next/next/no-img-element
              <img src={url} alt="" className="h-full w-full object-cover" loading="lazy" decoding="async" />
            )}
          </div>
        ))}
      </div>
      <div className="absolute inset-0 bg-gradient-to-b from-foreground/55 via-foreground/90 to-foreground" />
      <div className="relative flex h-full flex-col justify-center px-5 sm:px-8 lg:px-14">
        <span className="text-[10px] font-bold tracking-[0.14em] text-primary-foreground/60 sm:text-xs">
          ABOUT
        </span>
        <h1 className="mt-2 max-w-xl text-xl font-bold leading-snug tracking-tight text-background sm:mt-3.5 sm:text-3xl lg:text-[38px]">
          観た作品を、そのまま記録に。
        </h1>
        <p className="mt-3 hidden max-w-lg text-sm leading-loose text-background/80 sm:mt-4 sm:block lg:text-[15px]">
          MyAnimeLogsは、アニメの視聴状況を記録して自分だけのライブラリをつくるサービスです。今期アニメから旧作まで、観たい・観ている・観終わったを1か所で管理できます。
        </p>
      </div>
    </div>
  )
}

interface SectionHeadingProps {
  number: string
  title: string
}

const SectionHeading: React.FC<SectionHeadingProps> = ({ number, title }) => (
  <div className="mb-3.5 flex items-baseline gap-2.5 sm:mb-6">
    <span className="text-[11px] font-bold tabular-nums text-primary sm:text-xs">{number}</span>
    <h2 className="text-lg font-bold tracking-tight sm:text-[22px]">{title}</h2>
  </div>
)

const AboutPage: React.FC = () => {
  const { isAuthenticated } = useAuth()

  return (
    <div className="bg-muted/40">
      <AboutHero />

      <Container size="xl" className="py-9 sm:py-14 lg:py-[72px]">
        <p className="text-[13px] leading-relaxed text-muted-foreground sm:hidden">
          MyAnimeLogsは、アニメの視聴状況を記録して自分だけのライブラリをつくるサービスです。今期アニメから旧作まで、観たい・観ている・観終わったを1か所で管理できます。
        </p>

        <section className="mt-9 sm:mt-0">
          <SectionHeading number="01" title="主な機能" />
          <div className="grid grid-cols-1 gap-2.5 sm:grid-cols-2 lg:grid-cols-3 lg:gap-4.5">
            {FEATURES.map((feature) => (
              <div
                key={feature.title}
                className="flex gap-3.5 rounded-lg border border-border bg-card p-4 sm:block sm:p-5"
              >
                <div className="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary sm:mb-3.5">
                  <feature.icon className="h-[19px] w-[19px]" />
                </div>
                <div className="min-w-0">
                  <div className="mb-1 text-sm font-bold sm:mb-1.5 sm:text-[14.5px]">{feature.title}</div>
                  <p className="text-xs leading-relaxed text-muted-foreground sm:text-[13px]">{feature.desc}</p>
                </div>
              </div>
            ))}
          </div>
        </section>

        <section className="mt-9 sm:mt-[72px]">
          <SectionHeading number="02" title="使い方" />
          <div className="grid grid-cols-1 gap-2.5 sm:grid-cols-2 lg:grid-cols-4 lg:gap-4.5">
            {STEPS.map((step) => (
              <div
                key={step.n}
                className="flex gap-3.5 rounded-lg border border-border bg-muted/30 p-4 sm:block sm:p-5"
              >
                <div className="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full bg-primary text-[11.5px] font-bold text-primary-foreground sm:mb-3.5 sm:h-[26px] sm:w-[26px] sm:text-xs">
                  {step.n}
                </div>
                <div className="min-w-0">
                  <div className="mb-1 text-sm font-bold leading-snug sm:mb-1.5 sm:text-sm">{step.title}</div>
                  <p className="text-xs leading-relaxed text-muted-foreground">{step.desc}</p>
                </div>
              </div>
            ))}
          </div>
        </section>

        <section className="mt-9 sm:mt-[72px]">
          <SectionHeading number="03" title="データ提供元" />
          <div className="grid grid-cols-2 gap-2.5 lg:grid-cols-4 lg:gap-3.5">
            {DATA_SOURCES.map((source) => (
              <a
                key={source.name}
                href={source.href}
                target="_blank"
                rel="noopener noreferrer"
                className="rounded-lg border border-border bg-card p-3.5 transition-colors hover:border-primary/40 sm:p-4"
              >
                <div className="text-[10.5px] text-muted-foreground sm:text-[11.5px]">{source.kind}</div>
                <div className="mt-1.5 text-[13px] font-bold leading-snug text-primary sm:text-sm">
                  {source.name}
                </div>
                {source.note && (
                  <div className="mt-1 text-[10.5px] text-muted-foreground sm:text-[11.5px]">{source.note}</div>
                )}
              </a>
            ))}
          </div>
          <p className="mt-3.5 text-[10.5px] leading-relaxed text-muted-foreground sm:text-[11.5px]">
            This product uses the TMDB API but is not endorsed or certified by TMDB.
          </p>
        </section>

        {!isAuthenticated && (
          <section className="mt-9 rounded-xl bg-foreground px-6 py-7 sm:mt-[72px] sm:flex sm:items-center sm:justify-between sm:gap-8 sm:px-12 sm:py-11">
            <div>
              <div className="text-lg font-bold leading-snug text-background sm:text-[22px]">
                今日から、記録をはじめる。
              </div>
              <p className="mt-2 text-[12.5px] leading-relaxed text-background/75 sm:mt-2.5 sm:text-[13.5px]">
                Googleアカウントですぐに使いはじめられます。
              </p>
            </div>
            <Link
              href="/signin"
              className="mt-4 flex h-12 w-full items-center justify-center rounded-md bg-primary px-7 text-sm font-semibold text-primary-foreground transition-colors hover:opacity-90 sm:mt-0 sm:h-[52px] sm:w-auto sm:flex-shrink-0 sm:text-[15px]"
            >
              はじめる
            </Link>
          </section>
        )}
      </Container>
    </div>
  )
}

export default AboutPage
