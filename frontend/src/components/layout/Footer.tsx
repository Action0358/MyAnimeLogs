import React from 'react'
import Link from 'next/link'
import Image from 'next/image'
import type { Route } from 'next'
import { ChevronRight } from 'lucide-react'
import Container from './Container'
import { SUB_LOGO_PATH } from '@/lib/constants'

// 規約・ポリシーは Google OAuth 同意画面からも参照されるため、全ページから到達できるフッターに常設する
const FOOTER_LINKS: { href: Route; label: string }[] = [
  { href: '/contact' as Route, label: 'お問い合わせ' },
  { href: '/privacy' as Route, label: 'プライバシーポリシー' },
  { href: '/terms' as Route, label: '利用規約' },
]

const Footer: React.FC = () => {
  const currentYear = new Date().getFullYear()

  return (
    // 直前のお知らせセクションが既に下余白を持つため、ここでは追加のmt-16を入れない
    <footer className="border-t border-border bg-background">
      {/* 本文と同じ共通コンテナ（最大1180px）に揃える。幅が違うと
          ロゴ・リンクの左右端が本文とズレて見える */}
      <Container size="xl" className="pt-10 pb-8 sm:pb-10">
        <Link href={'/' as Route} className="inline-block">
          <Image
            src={SUB_LOGO_PATH}
            alt="MyAnimeLogs"
            width={1792}
            height={576}
            className="h-8 w-auto sm:h-9"
          />
        </Link>
        <p className="mt-2 text-[13px] text-muted-foreground">あなたの視聴記録がここに</p>

        {/*
          flex-col-reverse でSPのみ見た目の順序を反転させる。DOM順は
          コピーライト→リンクのままなので、PC（md:flex-row）ではコピーライトが
          左・リンクが右に並び、SPでは見た目上リンク→コピーライトの順になる
          （操作要素のリンクを上に、終端情報のコピーライトを最後に残す）。
        */}
        <div className="mt-7 flex flex-col-reverse gap-5 border-t border-border pt-5 sm:gap-6 md:flex-row md:items-center md:justify-between">
          <p className="text-xs text-muted-foreground">
            &copy; {currentYear} MyAnimeLogs. All rights reserved.
          </p>

          {/* PC: 横並びリンク。間隔は24px（詰まって1語に見えないように） */}
          <nav className="hidden items-center gap-6 md:flex">
            {FOOTER_LINKS.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                className="text-xs font-medium text-muted-foreground transition-colors hover:text-foreground"
              >
                {link.label}
              </Link>
            ))}
          </nav>

          {/* SP: 1行1リンク・高さ44pxでタップ領域を確保する（隣接リンクの誤タップ防止） */}
          <nav className="flex flex-col md:hidden">
            {FOOTER_LINKS.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                className="flex h-11 items-center justify-between border-b border-border text-[13px] font-medium text-muted-foreground last:border-b-0"
              >
                <span>{link.label}</span>
                <ChevronRight className="h-[15px] w-[15px] text-muted-foreground" aria-hidden="true" />
              </Link>
            ))}
          </nav>
        </div>
      </Container>
    </footer>
  )
}

export default Footer
