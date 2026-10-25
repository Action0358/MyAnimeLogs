import './globals.css'
import type { Metadata } from 'next'
import { Noto_Sans_JP } from 'next/font/google'
import { Header, Footer } from '@/components/layout'
import { QueryProvider } from '@/components/providers'
import { MaintenanceGate } from '@/components/maintenance'
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
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    // メンテナンス中の /news では、エッジが data-maintenance を付ける（React の描画には無い属性）
    <html lang="ja" className={notoSansJP.variable} suppressHydrationWarning>
      <body className="min-h-screen bg-background flex flex-col">
        <MaintenanceGate>
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
        </MaintenanceGate>
      </body>
    </html>
  )
}
