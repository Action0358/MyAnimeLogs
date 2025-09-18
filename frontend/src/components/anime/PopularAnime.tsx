'use client'

import type React from 'react'
import { useState, useEffect } from 'react'
import Link from 'next/link'
import { Button } from '@/components/ui'
import { annictApi, type AnnictWork } from '@/lib/api/annict'

interface PopularAnimeProps {
  className?: string
}

const PopularAnime: React.FC<PopularAnimeProps> = ({ className = '' }) => {
  const [animes, setAnimes] = useState<AnnictWork[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchPopularAnime = async () => {
      try {
        setLoading(true)
        const data = await annictApi.getPopularAnime()
        setAnimes(data.works || [])
      } catch (err) {
        setError(
          err instanceof Error ? err.message : '不明なエラーが発生しました'
        )
      } finally {
        setLoading(false)
      }
    }

    fetchPopularAnime()
  }, [])

  if (loading) {
    return (
      <div className={`${className}`}>
        <div className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <span className="ml-2 text-gray-600">読み込み中...</span>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className={`${className}`}>
        <div className="text-center py-12">
          <p className="text-red-600 mb-4">{error}</p>
          <Button variant="outline" onClick={() => window.location.reload()}>
            再試行
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className={`${className}`}>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold text-gray-900">人気アニメ</h2>
        <Link
          href={'/anime?sort=popular' as any}
          className="inline-flex items-center justify-center h-8 px-3 text-sm font-medium rounded-md border border-gray-300 bg-transparent text-gray-700 hover:bg-gray-50 active:bg-gray-100 transition-colors"
        >
          すべて見る
        </Link>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
        {animes.slice(0, 6).map((anime) => (
          <div key={anime.id} className="group cursor-pointer">
            <div className="aspect-[3/4] bg-gradient-to-br from-green-100 to-green-300 rounded-lg overflow-hidden mb-2 relative flex items-center justify-center">
              <span className="text-4xl">🎬</span>
              <div className="absolute inset-0 bg-black bg-opacity-0 group-hover:bg-opacity-20 transition-all duration-200"></div>
            </div>

            <div className="space-y-1">
              <h3
                className="text-sm font-medium text-gray-900 group-hover:text-blue-600 transition-colors overflow-hidden"
                style={{
                  display: '-webkit-box',
                  WebkitLineClamp: 2,
                  WebkitBoxOrient: 'vertical',
                }}
              >
                {anime.title}
              </h3>
              <div className="flex items-center justify-between text-xs text-gray-500">
                <span>{anime.media_text}</span>
                {anime.watchers_count > 0 && (
                  <span className="flex items-center">
                    <span className="mr-1">👥</span>
                    {anime.watchers_count}
                  </span>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>

      {animes.length === 0 && (
        <div className="text-center py-12">
          <p className="text-gray-500">人気アニメが見つかりませんでした</p>
        </div>
      )}
    </div>
  )
}

export default PopularAnime
