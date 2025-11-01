import './globals.css'
import type { Metadata } from 'next'
import { Header, Footer } from '@/components/layout'
import { QueryProvider } from '@/components/providers'
import { AuthProvider } from '@/contexts/AuthContext'

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
    <html lang="ja">
      <body className="min-h-screen bg-gray-50 flex flex-col">
        <AuthProvider>
          <QueryProvider>
            <Header />
            <main className="flex-1">
              {children}
            </main>
            <Footer />
          </QueryProvider>
        </AuthProvider>
      </body>
    </html>
  )
}
