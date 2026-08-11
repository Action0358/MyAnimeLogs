import './globals.css'
import type { Metadata } from 'next'
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
  icons: {
    icon: '/favicon.png',
    apple: '/favicon.png',
  },
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
