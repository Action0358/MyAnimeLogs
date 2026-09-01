'use client'

import React, { useEffect, useRef, useState } from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import { Container } from '@/components/layout'
import { OPERATOR_NAME, SUPPORT_EMAIL } from '@/lib/constants'
import { cn } from '@/lib/utils'

interface Article {
  num: string
  title: string
  paras: string[]
  items?: string[]
  tail?: string
}

const ARTICLES: Article[] = [
  {
    num: '第1条',
    title: '適用',
    paras: [
      `本規約は、${OPERATOR_NAME}（以下「運営者」）が提供する MyAnimeLogs（以下「本サービス」）の提供条件、および利用者と運営者との権利義務関係を定めるものです。利用者は、本サービスを利用することにより本規約に同意したものとみなされます。`,
      '本サービスは個人により運営されている無償のサービスです。',
    ],
  },
  {
    num: '第2条',
    title: 'アカウント登録',
    paras: [],
    items: [
      '本サービスの利用にはGoogleアカウントによる認証が必要です。',
      '運営者は、Googleアカウントから提供されるメールアドレスおよびプロフィール名を、アカウントの識別と表示のために利用します。',
      '利用者は、自己のアカウントを第三者に利用させてはなりません。',
      'アカウントの管理不十分により生じた損害の責任は利用者が負うものとします。',
    ],
  },
  {
    num: '第3条',
    title: '禁止事項',
    paras: ['利用者は、本サービスの利用にあたり、以下の行為を行ってはなりません。'],
    items: [
      '法令または公序良俗に違反する行為',
      '犯罪行為に関連する行為',
      '運営者、他の利用者、または第三者の知的財産権、名誉、プライバシーその他の権利を侵害する行為',
      '本サービスのサーバーまたはネットワークの機能を破壊・妨害する行為',
      '本サービスの運営を妨害するおそれのある行為',
      '不正アクセスを試みる行為',
      '本サービスの情報を、自動化された手段により過度に取得する行為',
      '他の利用者になりすます行為',
      'その他、運営者が不適切と判断する行為',
    ],
  },
  {
    num: '第4条',
    title: '本サービスの提供の停止等',
    paras: [
      '運営者は、以下のいずれかの事由があると判断した場合、利用者に事前に通知することなく本サービスの全部または一部の提供を停止または中断することができます。',
    ],
    items: [
      '本サービスにかかるシステムの保守点検または更新を行う場合',
      '地震、落雷、火災、停電、天災などの不可抗力により本サービスの提供が困難となった場合',
      'コンピュータまたは通信回線等が事故により停止した場合',
      '本サービスが利用する外部APIまたは外部サービスの提供が停止または変更された場合',
      'その他、運営者が本サービスの提供が困難と判断した場合',
    ],
    tail: '運営者は、本サービスの提供の停止または中断により利用者または第三者が被った不利益・損害について、一切の責任を負わないものとします。',
  },
  {
    num: '第5条',
    title: '作品情報の取り扱い',
    paras: [
      '本サービスが表示するアニメ作品の情報および画像は、Annict、AniList、Jikan（MyAnimeList）、TMDBなどの外部サービスから取得したものです。これらの情報に関する権利は、各提供元および原著作権者に帰属します。',
      '運営者は、これらの情報の正確性、完全性、最新性について保証しません。',
    ],
  },
  {
    num: '第6条',
    title: '利用者の記録データ',
    paras: [],
    items: [
      '利用者が本サービスに登録した視聴状況・お気に入り等の記録データは、当該利用者に帰属します。',
      '運営者は、本サービスの提供・改善のために必要な範囲でこれらのデータを利用します。',
      '運営者は、個人を特定できない統計情報として、これらのデータを集計・利用する場合があります。',
    ],
  },
  {
    num: '第7条',
    title: '保証の否認および免責事項',
    paras: [
      '本サービスは現状有姿で提供されます。運営者は、本サービスに事実上または法律上の瑕疵がないことを明示的にも黙示的にも保証しません。',
      '運営者は、本サービスに起因して利用者に生じたあらゆる損害について、運営者の故意または重過失による場合を除き、一切の責任を負いません。',
    ],
  },
  {
    num: '第8条',
    title: 'サービス内容の変更・終了',
    paras: [
      '運営者は、利用者への事前の通知をもって、本サービスの内容を変更し、または本サービスの提供を終了することができます。本サービスを終了する場合、運営者は合理的な期間をもって事前に告知するよう努めます。',
    ],
  },
  {
    num: '第9条',
    title: 'アカウントの削除',
    paras: [
      '利用者は、いつでも自己のアカウントを削除することができます。アカウント削除により、利用者の記録データは削除されます。削除されたデータの復旧はできません。',
    ],
  },
  {
    num: '第10条',
    title: '利用規約の変更',
    paras: [
      '運営者は、必要と判断した場合には、利用者に通知することなく本規約を変更することができます。変更後の規約は、本ページに掲載した時点から効力を生じるものとします。',
    ],
  },
  {
    num: '第11条',
    title: '準拠法・裁判管轄',
    paras: [
      '本規約の解釈にあたっては日本法を準拠法とします。本サービスに関して紛争が生じた場合には、運営者の所在地を管轄する裁判所を専属的合意管轄とします。',
    ],
  },
]

const CONTACT_ID = 'contact'

// 目次のジャンプ先。IntersectionObserver で現在地をハイライトするために各条の id と対応させる
const TOC_ITEMS = [
  ...ARTICLES.map((a, i) => ({ id: `article-${i}`, num: String(i + 1), label: a.title })),
  { id: CONTACT_ID, num: '', label: 'お問い合わせ' },
]

/**
 * 利用規約ページ本文。デスクトップは追従する目次サイドバー、
 * モバイルは単一カラムで表示する（#790）。
 *
 * /privacy 刷新（別ブランチで進行中）とはコンポーネントを共有せず、
 * /terms 専用に閉じている。共有化は両方の刷新が揃ってから検討する。
 */
export const TermsContent: React.FC = () => {
  const [activeId, setActiveId] = useState(TOC_ITEMS[0].id)
  const sectionRefs = useRef<Record<string, HTMLElement | null>>({})

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        const visible = entries.filter((e) => e.isIntersecting)
        if (visible.length === 0) return
        const topMost = visible.reduce((a, b) => (a.boundingClientRect.top < b.boundingClientRect.top ? a : b))
        const id = topMost.target.getAttribute('id')
        if (id) setActiveId(id)
      },
      { rootMargin: '-96px 0px -70% 0px', threshold: 0 }
    )

    TOC_ITEMS.forEach(({ id }) => {
      const el = sectionRefs.current[id]
      if (el) observer.observe(el)
    })

    return () => observer.disconnect()
  }, [])

  return (
    <div className="bg-muted/30 min-h-screen">
      <Container size="xl" className="max-w-[1180px] py-8 md:py-10">
        {/* パンくず */}
        <nav aria-label="パンくずリスト" className="mb-4 text-[12px] md:text-[12.5px]">
          <Link href={'/' as Route} className="text-muted-foreground hover:text-primary transition-colors">
            ホーム
          </Link>
          <span className="text-muted-foreground mx-1.5" aria-hidden="true">
            /
          </span>
          <span className="text-foreground">利用規約</span>
        </nav>

        <div className="grid grid-cols-1 md:grid-cols-[236px_minmax(0,1fr)] gap-9 items-start">
          {/* 目次（デスクトップのみ）。現在地をIntersectionObserverでハイライトする */}
          <nav
            aria-label="目次"
            className="hidden md:block sticky top-6 rounded-lg border border-border bg-card py-[18px] pl-[18px] pr-2"
          >
            <div className="mb-3 text-[11px] font-bold tracking-[0.08em] text-muted-foreground">目次</div>
            <div className="flex max-h-[520px] flex-col gap-px overflow-y-auto pr-2.5">
              {TOC_ITEMS.map((item) => {
                const isActive = item.id === activeId
                return (
                  <a
                    key={item.id}
                    href={`#${item.id}`}
                    className={cn(
                      'flex gap-2.5 rounded-md px-2.5 py-[7px] text-[12.5px] leading-[1.5] transition-colors',
                      isActive ? 'bg-primary/10 text-primary font-bold' : 'text-foreground/70 hover:bg-muted'
                    )}
                  >
                    {item.num && (
                      <span
                        className={cn(
                          'min-w-[16px] flex-shrink-0 pt-px text-[11px] font-bold tabular-nums',
                          isActive ? 'text-primary' : 'text-muted-foreground'
                        )}
                      >
                        {item.num}
                      </span>
                    )}
                    <span>{item.label}</span>
                  </a>
                )
              })}
            </div>
          </nav>

          {/* 本文 */}
          <div className="min-w-0">
            <div className="max-w-[720px] rounded-lg border border-border bg-card px-6 py-11 md:px-14 md:py-12">
              <h1 className="border-b border-border pb-[22px] text-2xl font-bold tracking-tight text-foreground md:text-[30px]">
                利用規約
              </h1>

              {ARTICLES.map((article, i) => {
                const id = `article-${i}`
                return (
                  <div
                    key={id}
                    id={id}
                    ref={(el) => {
                      sectionRefs.current[id] = el
                    }}
                    className="scroll-mt-24 border-b border-border/70 py-[26px] first:pt-[26px] md:py-[34px]"
                  >
                    <div className="mb-[7px] text-[11.5px] font-bold tracking-[0.06em] text-primary">
                      {article.num}
                    </div>
                    <h2 className="mb-3.5 text-[17px] font-bold tracking-tight text-foreground md:text-[19px]">
                      {article.title}
                    </h2>
                    {article.paras.map((p, pi) => (
                      <p key={pi} className="mb-3 text-[13px] leading-[1.95] text-foreground/80 md:text-sm">
                        {p}
                      </p>
                    ))}
                    {article.items && article.items.length > 0 && (
                      <div className="mt-1.5 flex flex-col gap-2">
                        {article.items.map((li, li_i) => (
                          <div key={li_i} className="flex gap-2.5 text-[12.5px] leading-[1.85] text-foreground/80 md:text-[13.5px]">
                            <span className="mt-2.5 h-[5px] w-[5px] flex-shrink-0 rounded-full bg-primary/40" />
                            <span>{li}</span>
                          </div>
                        ))}
                      </div>
                    )}
                    {article.tail && (
                      <p className="mt-3.5 text-[13px] leading-[1.95] text-foreground/80 md:text-sm">
                        {article.tail}
                      </p>
                    )}
                  </div>
                )
              })}

              {/* お問い合わせ */}
              <div
                id={CONTACT_ID}
                ref={(el) => {
                  sectionRefs.current[CONTACT_ID] = el
                }}
                className="scroll-mt-24 mt-[34px] rounded-lg border border-border bg-muted/30 px-[22px] py-6 md:px-[26px]"
              >
                <h2 className="mb-2.5 text-[15px] font-bold text-foreground md:text-base">お問い合わせ</h2>
                <p className="mb-4 text-[12.5px] leading-[1.9] text-foreground/80 md:text-[13.5px]">
                  本規約に関するお問い合わせは、以下までご連絡ください。
                </p>
                <div className="flex flex-col gap-2.5 text-[12.5px] md:grid md:grid-cols-[64px_1fr] md:gap-x-3.5 md:gap-y-2.5 md:text-[13.5px]">
                  <span className="text-muted-foreground">運営者</span>
                  <span className="font-medium text-foreground">{OPERATOR_NAME}</span>
                  <span className="text-muted-foreground">連絡先</span>
                  <a
                    href={`mailto:${SUPPORT_EMAIL}`}
                    className="break-all font-semibold text-primary hover:underline"
                  >
                    {SUPPORT_EMAIL}
                  </a>
                </div>
              </div>
            </div>
          </div>
        </div>
      </Container>
    </div>
  )
}
