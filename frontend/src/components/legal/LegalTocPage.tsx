import React from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import { Container } from '@/components/layout'

export interface LegalTocItem {
  id: string
  label: string
}

interface LegalTocPageProps {
  title: string
  tocItems: LegalTocItem[]
  children: React.ReactNode
  /** 目次末尾に足す追加リンク（お問い合わせカードへのジャンプなど） */
  extraTocLabel?: string
  extraTocHref?: string
}

/**
 * 目次サイドバー付きの法務ページレイアウト。
 * 既存の LegalPage は /terms と共有しているため変更せず、
 * /privacy 刷新用にこちらを新設して影響範囲を分離している。
 */
export const LegalTocPage: React.FC<LegalTocPageProps> = ({
  title,
  tocItems,
  children,
  extraTocLabel,
  extraTocHref,
}) => (
  <Container size="xl" className="py-8 sm:py-10">
    <nav aria-label="パンくず" className="mb-4 text-xs text-muted-foreground sm:mb-5">
      <Link href="/" className="hover:text-foreground">
        ホーム
      </Link>
      <span className="mx-1.5">/</span>
      <span className="text-foreground/80">{title}</span>
    </nav>

    <div className="grid grid-cols-1 items-start gap-6 lg:grid-cols-[236px_minmax(0,1fr)] lg:gap-9">
      <aside className="hidden rounded-lg border border-border bg-card p-3.5 lg:sticky lg:top-6 lg:block">
        <div className="mb-3 px-1.5 text-[11px] font-bold tracking-wide text-muted-foreground">
          目次
        </div>
        <div className="flex max-h-[520px] flex-col gap-0.5 overflow-y-auto pr-1">
          {tocItems.map((item, index) => (
            <a
              key={item.id}
              href={`#${item.id}`}
              className="rounded-md px-2 py-1.5 text-[12.5px] leading-snug text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
            >
              <span className="mr-2 tabular-nums text-muted-foreground/70">{index + 1}</span>
              {item.label}
            </a>
          ))}
          {extraTocLabel && extraTocHref && (
            <a
              href={extraTocHref}
              className="rounded-md px-2 py-1.5 text-[12.5px] leading-snug text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
            >
              {extraTocLabel}
            </a>
          )}
        </div>
      </aside>

      <div className="min-w-0 rounded-lg border border-border bg-card px-5 py-7 sm:px-8 sm:py-9 lg:px-14 lg:py-11">
        <h1 className="border-b border-border pb-5 text-xl font-bold tracking-tight text-foreground sm:text-[26px] lg:pb-5.5 lg:text-[30px]">
          {title}
        </h1>
        {children}
      </div>
    </div>
  </Container>
)

interface LegalArticleProps {
  id: string
  num: string
  title: string
  children: React.ReactNode
}

/** 目次から#idでジャンプする1節。scroll-margin-topでヘッダー分のオフセットを確保する */
export const LegalArticle: React.FC<LegalArticleProps> = ({ id, num, title, children }) => (
  <section id={id} className="scroll-mt-20 border-b border-border/70 py-6 last:border-b-0 sm:py-8">
    <div className="mb-1.5 text-[11px] font-bold tracking-wide text-primary">{num}</div>
    <h2 className="mb-3.5 text-base font-bold tracking-tight text-foreground sm:text-[19px]">{title}</h2>
    <div className="space-y-3 text-[13.5px] leading-[1.95] text-muted-foreground sm:text-sm">{children}</div>
  </section>
)

interface LegalDefinition {
  term: string
  desc: string
}

/** 「用語＋説明」を枠付きリストで示す */
export const LegalDefinitionList: React.FC<{ items: LegalDefinition[] }> = ({ items }) => (
  <div className="mt-1.5 overflow-hidden rounded-lg border border-border">
    {items.map((item) => (
      <div key={item.term} className="border-b border-border/70 p-3.5 last:border-b-0 sm:p-4">
        <div className="mb-1 text-[13px] font-bold text-foreground">{item.term}</div>
        <div className="text-[12.5px] leading-relaxed text-muted-foreground sm:text-[13px]">{item.desc}</div>
      </div>
    ))}
  </div>
)

/** 中黒ドットの箇条書き。デザイントークンに合わせたアクセントカラーのドットを使う */
export const LegalDotList: React.FC<{ items: string[] }> = ({ items }) => (
  <div className="mt-1.5 flex flex-col gap-2">
    {items.map((item) => (
      <div key={item} className="flex gap-2.5 text-[13px] leading-relaxed text-muted-foreground sm:text-[13.5px]">
        <span className="mt-[9px] h-[5px] w-[5px] flex-shrink-0 rounded-full bg-primary/40" />
        <span>{item}</span>
      </div>
    ))}
  </div>
)

interface LegalContactCardProps {
  id: string
  operator: string
  email: string
}

/**
 * 法的ページ（プライバシーポリシー）の問い合わせ窓口。
 *
 * メールを主に出すのは、個人情報の開示・訂正・削除の請求窓口だから。
 * 法的な請求の受付をフォームだけにしないため、ここはメールのまま残す（#869）。
 * 一般の問い合わせは /contact のフォームへ誘導する。
 */
export const LegalContactCard: React.FC<LegalContactCardProps> = ({ id, operator, email }) => (
  <div id={id} className="mt-7 scroll-mt-20 rounded-lg border border-border bg-muted/40 p-5 sm:p-6">
    <h2 className="mb-2 text-sm font-bold text-foreground sm:text-base">お問い合わせ</h2>
    <p className="mb-4 text-[12.5px] leading-relaxed text-muted-foreground sm:text-[13.5px]">
      保有する情報の開示・訂正・削除のご請求は、以下までご連絡ください。
      不具合のご報告など本ポリシー以外のお問い合わせは、
      <Link href={'/contact' as Route} className="mx-0.5 text-primary hover:underline">
        お問い合わせフォーム
      </Link>
      をご利用ください。
    </p>
    <div className="flex flex-col gap-2.5 text-[12.5px] sm:grid sm:grid-cols-[64px_1fr] sm:gap-x-3.5 sm:gap-y-2 sm:text-[13.5px]">
      <div className="text-muted-foreground">運営者</div>
      <div className="font-medium text-foreground">{operator}</div>
      <div className="text-muted-foreground">連絡先</div>
      <a href={`mailto:${email}`} className="break-all font-semibold text-primary hover:underline">
        {email}
      </a>
    </div>
  </div>
)
