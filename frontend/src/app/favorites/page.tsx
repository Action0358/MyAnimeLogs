'use client'

import React from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import { Container } from '@/components/layout'
import { Button } from '@/components/ui/Button'

const FavoritesPage: React.FC = () => {
  return (
    <Container size="lg" className="py-12">
      <h1 className="text-3xl font-bold text-gray-900 mb-8">
        お気に入りアニメ
      </h1>

      {/* 空状態の表示 */}
      <div className="text-center py-16">
        <div className="text-6xl mb-4">⭐</div>
        <h2 className="text-xl font-semibold text-gray-700 mb-2">
          お気に入りがまだありません
        </h2>
        <p className="text-gray-500 mb-6">
          気になるアニメを見つけたら、お気に入りに追加してみましょう
        </p>
        <Link href={"/" as Route}>
          <Button variant="primary" size="md">
            アニメを探す
          </Button>
        </Link>
      </div>

      {/* TODO: お気に入りアニメ一覧の実装 */}
      {/* 
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {favorites.map((anime) => (
          <AnimeCard key={anime.id} anime={anime} />
        ))}
      </div>
      */}
    </Container>
  )
}

export default FavoritesPage
