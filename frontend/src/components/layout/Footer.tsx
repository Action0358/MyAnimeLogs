import React from 'react'
import Link from 'next/link'
import Image from 'next/image'
import type { Route } from 'next'
import Container from './Container'
import { FOOTER_LINKS } from './navigationConfig'
import { SUB_LOGO_PATH } from '@/lib/constants'

// フッターカラムヘッダーコンポーネント
interface FooterColumnHeaderProps {
  children: React.ReactNode
}

const FooterColumnHeader = ({ children }: FooterColumnHeaderProps) => (
  <div className="h-10 flex items-center mb-4">
    {children}
  </div>
)

const Footer: React.FC = () => {
  const currentYear = new Date().getFullYear()

  return (
    <footer className="bg-background border-t border-border mt-16">
      <Container>
        <div className="py-12">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-8">
            {/* サイト情報 */}
            <div>
              <FooterColumnHeader>
                <Link href={"/" as Route}>
                  <Image
                    src={SUB_LOGO_PATH}
                    alt="MyAnimeLogs"
                    width={1792}
                    height={576}
                    className="h-10 w-auto"
                  />
                </Link>
              </FooterColumnHeader>
              <p className="text-muted-foreground">あなたの視聴記録がここに</p>
            </div>

            {/* サイトマップ */}
            <div>
              <FooterColumnHeader>
                <h3 className="text-lg font-semibold text-foreground">
                  サイトマップ
                </h3>
              </FooterColumnHeader>
              {/*
                項目数が多いため2列に流し、フッターが縦に伸びるのを防ぐ。
                grid-flow-col + grid-rows-4 で「横→縦」ではなく「縦→横」に
                並べることで、FOOTER_LINKS の並び順どおりに上から読め、
                今期アニメ・人気アニメ・お気に入りが同じ列で隣接する。
              */}
              <ul className="grid grid-rows-4 grid-flow-col gap-x-6 gap-y-2">
                {FOOTER_LINKS.map((item) => (
                  <li key={item.href}>
                    <Link
                      href={item.href as Route}
                      className="text-muted-foreground hover:text-foreground transition-colors"
                    >
                      {item.label}
                    </Link>
                  </li>
                ))}
              </ul>
            </div>

          </div>

          {/* コピーライト */}
          <div className="border-t border-border mt-8 pt-8">
            <div className="flex flex-col md:flex-row justify-between items-center">
              <p className="text-muted-foreground text-sm">
                &copy; {currentYear} MyAnimeLogs. All rights reserved.
              </p>
              <div className="flex items-center space-x-4 mt-4 md:mt-0">
                {/* 規約・ポリシーは Google OAuth 同意画面からも参照されるため、
                    全ページから到達できるフッターに常設する */}
                <Link
                  href={"/terms" as Route}
                  className="text-muted-foreground hover:text-foreground text-sm transition-colors"
                >
                  利用規約
                </Link>
                <Link
                  href={"/privacy" as Route}
                  className="text-muted-foreground hover:text-foreground text-sm transition-colors"
                >
                  プライバシーポリシー
                </Link>
              </div>
            </div>
          </div>
        </div>
      </Container>
    </footer>
  )
}

export default Footer
