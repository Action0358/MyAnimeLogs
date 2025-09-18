import './globals.css'
import type { Metadata } from 'next'
import { Header, Footer } from '@/components/layout'
import { QueryProvider } from '@/components/providers'

export const metadata: Metadata = {
  title: 'MyAnimeLogs',
  description: 'アニメ視聴管理アプリ',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="ja">
      <body className="min-h-screen bg-gray-50 flex flex-col">
        <QueryProvider>
          <Header />
          <main className="flex-1">
            {children}
          </main>
          <Footer />
        </QueryProvider>
      </body>
    </html>
  )
}
