'use client'

import React from 'react'
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
    const lower = provider.provider_name.toLowerCase()
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
    <div className="mb-6 bg-white rounded-lg border p-4 md:p-6">
      <h2 className="text-lg font-semibold text-gray-900 mb-3">配信サービス</h2>
      <div className="flex flex-wrap gap-3">
        {visibleProviders.map((provider) => (
          // ロゴ単体で表示（サービス名はチップ上には表示せず、title/alt で保持）
          <div
            key={provider.provider_id}
            className="flex items-center justify-center p-2 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
            title={provider.provider_name}
          >
            {provider.logo_url ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img
                src={provider.logo_url}
                alt={provider.provider_name}
                className="w-8 h-8 rounded object-contain"
              />
            ) : (
              <div
                className="w-8 h-8 bg-gray-200 rounded flex items-center justify-center"
                role="img"
                aria-label={provider.provider_name}
              >
                <span className="text-xs text-gray-500">📺</span>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}

export default WatchProviders
