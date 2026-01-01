import AnimeDetailClient from './AnimeDetailClient'

// 静的エクスポート用: ビルド時に生成するパスを指定
// プレースホルダーIDを返すことで、クライアントサイドで実際のIDを処理
export function generateStaticParams() {
  // ダミーのIDを返す（実際のルーティングはクライアントサイドで処理）
  return [{ id: '_' }]
}

export default function AnimeDetailPage() {
  return <AnimeDetailClient />
}
