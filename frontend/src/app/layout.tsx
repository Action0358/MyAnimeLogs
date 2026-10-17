import './globals.css'
import type { Metadata, Viewport } from 'next'
import { Noto_Sans_JP } from 'next/font/google'
import { Header, Footer } from '@/components/layout'
import { QueryProvider } from '@/components/providers'
import { AuthProvider } from '@/contexts/AuthContext'
import { TokenDebug } from '@/components/debug/TokenDebug'

const notoSansJP = Noto_Sans_JP({
  subsets: ['latin'],
  weight: ['400', '500', '700'],
  variable: '--font-noto-sans-jp',
  display: 'swap',
})

export const metadata: Metadata = {
  title: 'MyAnimeLogs',
  description: 'アニメ視聴管理アプリ',
  manifest: '/manifest.json',
  icons: {
    icon: [
      { url: '/favicon-16x16.png', sizes: '16x16', type: 'image/png' },
      { url: '/favicon-32x32.png', sizes: '32x32', type: 'image/png' },
    ],
    apple: '/apple-touch-icon.png',
  },
  appleWebApp: {
    capable: true,
    // ホーム画面追加時、ステータスバー領域を透過してページ背景を見せる
    // (黒文字/白文字どちらにも対応する唯一の設定値)。トップのヒーロー背景が
    // ステータスバー裏まで続くようにするのが目的で、Header側にsafe-area分の
    // paddingを足して時計・バッテリー表示とUIの衝突を防いでいる。
    statusBarStyle: 'black-translucent',
    title: 'MyAnimeLogs',
  },
}

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  viewportFit: 'cover',
  themeColor: '#6366f1',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="ja" className={notoSansJP.variable}>
      <body className="min-h-screen bg-background flex flex-col">
        <AuthProvider>
          <QueryProvider>
            <Header />
            <main className="flex-1">
              {children}
            </main>
            <Footer />
            {process.env.NODE_ENV === 'development' && <TokenDebug />}
          </QueryProvider>
        </AuthProvider>
      </body>
    </html>
  )
}
