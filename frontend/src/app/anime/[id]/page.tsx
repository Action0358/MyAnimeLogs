import AnimeDetailClient from './AnimeDetailClient'

/**
 * 静的エクスポート（output: 'export'）用のフォールバック設定
 *
 * Next.jsの静的エクスポートでは、動的ルート[id]に対して
 * generateStaticParamsでビルド時に生成するパスを指定する必要がある。
 *
 * しかし、アニメIDは動的に増減するため、すべてのIDを事前に列挙することは現実的ではない。
 * そのため、ダミーのID '_' を1つだけ返し、以下の仕組みで動作させている：
 *
 * 1. ビルド時: /anime/_/index.html のみ生成
 * 2. 実行時: /anime/123 へのアクセス時、Cloudflare PagesのSPAフォールバックが発動
 * 3. クライアント側: useParams()でURLからIDを取得し、APIを呼び出してデータを表示
 *
 * 注意: この実装はSEO/OGPに弱点がある。詳細は以下のissueを参照：
 * .kiro/tickets/frontend/open/ISSUE_DYNAMIC_ROUTE_SEO.md
 */
export function generateStaticParams() {
  return [{ id: '_' }]
}

export default function AnimeDetailPage() {
  return <AnimeDetailClient />
}
