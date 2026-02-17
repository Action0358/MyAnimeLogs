'use client'

import React from 'react'
import type { WatchProvider } from '@/lib/api/generated/myanimelogs'

interface WatchProvidersProps {
  providers: WatchProvider[]
}

export const WatchProviders: React.FC<WatchProvidersProps> = ({ providers }) => {
  if (!providers || providers.length === 0) {
    return null
  }

  return (
    <div className="mb-6 bg-white rounded-lg border p-4 md:p-6">
      <h2 className="text-lg font-semibold text-gray-900 mb-3">配信サービス</h2>
      <div className="flex flex-wrap gap-3">
        {providers.map((provider) => (
          <div
            key={provider.provider_id}
            className="flex items-center gap-2 px-3 py-2 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
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
              <div className="w-8 h-8 bg-gray-200 rounded flex items-center justify-center">
                <span className="text-xs text-gray-500">📺</span>
              </div>
            )}
            <span className="text-sm text-gray-700">{provider.provider_name}</span>
          </div>
        ))}
      </div>
      <p className="mt-3 text-xs text-gray-400">
        配信情報は
        <a
          href="https://www.themoviedb.org/"
          target="_blank"
          rel="noopener noreferrer"
          className="text-blue-500 hover:underline ml-1"
        >
          TMDB
        </a>
        より提供
      </p>
    </div>
  )
}

export default WatchProviders
