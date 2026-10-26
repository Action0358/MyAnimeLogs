'use client'

import React from 'react'
import { Card, CardContent, SectionHeading } from '@/components/ui'
import type { WatchProvider } from '@/lib/api/generated/myanimelogs'

interface WatchProvidersProps {
  providers: WatchProvider[]
}

// 同一サービスの広告付き廉価版やチャンネル再販などを1つに集約するためのグループ定義。
// matches に該当する provider は同じ key にまとめ、preferredName（基本名）を優先表示する。
const CANONICAL_PROVIDER_GROUPS: {
  key: string
  preferredName: string
  matches: (lowerName: string) => boolean
}[] = [
  {
    key: 'netflix',
    preferredName: 'Netflix',
    // Netflix / Netflix Standard with Ads など
    matches: (n) => n.includes('netflix'),
  },
  {
    key: 'amazon-prime-video',
    preferredName: 'Amazon Prime Video',
    // Amazon Prime Video / Prime Video / Amazon Prime Video with Ads / Amazonプライムビデオ /
    // dAnime Amazon Channel など。表記ゆれは拾いつつ、無関係な「○○ Amazon Channel」は
    // 過剰集約しないよう 'amazon' 単独では判定しない。
    matches: (n) =>
      n.includes('prime video') ||
      n.includes('プライムビデオ') ||
      n.includes('danime amazon channel'),
  },
]

// 重複サービスを集約する。グループに該当しない provider はそのまま残す。
// グループ内では基本名（preferredName）を優先し、無ければ最初に出現したものを採用する。
function dedupeProviders(providers: WatchProvider[]): WatchProvider[] {
  const byKey = new Map<string, WatchProvider>()
  for (const provider of providers) {
    // APIが provider_name を null/undefined で返した場合のクラッシュを防ぐ
    const lower = (provider.provider_name ?? '').toLowerCase()
    const group = CANONICAL_PROVIDER_GROUPS.find((g) => g.matches(lower))
    const key = group ? group.key : `id:${provider.provider_id}`
    const existing = byKey.get(key)
    if (!existing) {
      byKey.set(key, provider)
    } else if (group && lower === group.preferredName.toLowerCase()) {
      // 基本名（例: Netflix / Amazon Prime Video）を広告付き・チャンネル版より優先。
      // APIの大文字小文字ゆれ（netflix / NETFLIX 等）に対応するため小文字で比較。
      byKey.set(key, provider)
    }
  }
  return Array.from(byKey.values())
}

export const WatchProviders: React.FC<WatchProvidersProps> = ({ providers }) => {
  if (!providers || providers.length === 0) {
    return null
  }

  const visibleProviders = dedupeProviders(providers)

  return (
    <Card>
      <CardContent className="p-5 md:px-6 md:py-5">
        <SectionHeading as="h2" size="sm">
          配信サービス
        </SectionHeading>
        {/* ロゴは元画像の縦横比・余白がまちまちなため、均一なタイルに収めて揃える。
            48px（SP）はタイル5つが横一列に収まるサイズ */}
        <div className="flex flex-wrap gap-2 md:gap-2.5">
          {visibleProviders.map((provider) => (
            // ロゴ単体で表示（サービス名はタイル上には表示せず、title/alt で保持）
            <div
              key={provider.provider_id}
              className="w-12 h-12 md:w-14 md:h-14 flex items-center justify-center p-1.5 bg-muted border border-border rounded-[10px] hover:bg-accent transition-colors"
              title={provider.provider_name}
            >
              {provider.logo_url ? (
                // eslint-disable-next-line @next/next/no-img-element
                <img
                  src={provider.logo_url}
                  alt={provider.provider_name}
                  className="max-w-full max-h-full rounded object-contain"
                />
              ) : (
                <div
                  className="w-full h-full flex items-center justify-center bg-muted-foreground/10 rounded"
                  role="img"
                  aria-label={provider.provider_name}
                >
                  <span className="text-lg text-muted-foreground">📺</span>
                </div>
              )}
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}

export default WatchProviders
