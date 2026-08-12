import React from 'react'
import { Container } from '@/components/layout'

interface LegalSectionProps {
  title: string
  children: React.ReactNode
}

/** 法務ページの1条分。見出しと本文の余白を統一する */
export const LegalSection: React.FC<LegalSectionProps> = ({ title, children }) => (
  <section className="mb-10">
    <h2 className="mb-3 text-xl font-semibold text-foreground">{title}</h2>
    <div className="space-y-3 text-[15px] leading-relaxed text-muted-foreground">{children}</div>
  </section>
)

interface LegalPageProps {
  title: string
  /** 「最終更新日: 2026-08-12」形式で表示する日付 */
  updatedAt: string
  children: React.ReactNode
}

export const LegalPage: React.FC<LegalPageProps> = ({ title, updatedAt, children }) => (
  <Container className="py-12">
    <div className="mx-auto max-w-3xl">
      <h1 className="mb-2 text-3xl font-bold text-foreground">{title}</h1>
      <p className="mb-10 text-sm text-muted-foreground">最終更新日: {updatedAt}</p>
      {children}
    </div>
  </Container>
)

/** 箇条書き。法務ページ内で余白と記号の扱いを揃える */
export const LegalList: React.FC<{ items: React.ReactNode[] }> = ({ items }) => (
  <ul className="list-disc space-y-2 pl-5 marker:text-primary">
    {items.map((item, index) => (
      <li key={index}>{item}</li>
    ))}
  </ul>
)
