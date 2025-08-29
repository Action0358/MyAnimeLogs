import './globals.css'
import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'MyAnimeLog',
  description: 'Anime tracking app',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="ja">
      <body>{children}</body>
    </html>
  )
}
