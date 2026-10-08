import React from 'react'
import Link from 'next/link'
import Image from 'next/image'
import type { Route } from 'next'
import Container from './Container'
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
          {/* サイト情報 */}
          <div className="mb-8">
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
                  href={"/contact" as Route}
                  className="text-muted-foreground hover:text-foreground text-sm transition-colors"
                >
                  問い合わせ
                </Link>
                <Link
                  href={"/privacy" as Route}
                  className="text-muted-foreground hover:text-foreground text-sm transition-colors"
                >
                  プライバシーポリシー
                </Link>
                <Link
                  href={"/terms" as Route}
                  className="text-muted-foreground hover:text-foreground text-sm transition-colors"
                >
                  利用規約
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
