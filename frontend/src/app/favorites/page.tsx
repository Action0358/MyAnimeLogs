'use client'

import React, { useMemo } from 'react'
import { useRouter } from 'next/navigation'
import type { Route } from 'next'
import { Heart } from 'lucide-react'
import { Container } from '@/components/layout'
import { ProtectedRoute } from '@/components/auth/ProtectedRoute'
import { Button } from '@/components/ui/Button'
import { EmptyState } from '@/components/ui/EmptyState'
import { ExpandToggleButton } from '@/components/ui/ExpandToggleButton'
import { FavoriteCard } from '@/components/favorites'
import { SkeletonCard } from '@/components/anime'
import { useFavorites, FavoriteAnime } from '@/hooks/useFavorites'
import { useExpandableList } from '@/hooks/useExpandableList'
import { useResponsiveInitialCount } from '@/hooks/useResponsiveInitialCount'

// お気に入りの初期表示件数（超えた分は展開ボタンで表示する）。
// グリッドはSP3列・PC6列のため、SPは詰まりすぎないよう少なめにする
const FAVORITES_INITIAL_COUNT_SP = 18
const FAVORITES_INITIAL_COUNT_PC = 24

// id が number のお気に入りのみを抽出
const filterValidFavorites = (favorites: FavoriteAnime[]): (FavoriteAnime & { id: number })[] => {
  return favorites.filter((anime): anime is FavoriteAnime & { id: number } =>
    typeof anime.id === 'number'
  )
}

const FavoritesContent: React.FC = () => {
  const router = useRouter()
  // 解除は視聴ステータスと同様に詳細ページで行うため、一覧では取得のみ
  const { favorites, isLoading, error } = useFavorites()

  const validFavorites = useMemo(() => filterValidFavorites(favorites), [favorites])
  const { initialCount, oneRowAtWidest } = useResponsiveInitialCount(
    FAVORITES_INITIAL_COUNT_SP,
    FAVORITES_INITIAL_COUNT_PC
  )
  const {
    visible: visibleFavorites,
    hiddenCount,
    isExpanded,
    toggle: toggleExpanded,
  } = useExpandableList(validFavorites, initialCount, oneRowAtWidest)

  if (error) {
    return (
      <Container size="xl" className="py-8">
        <div className="text-center py-16">
          <div className="text-6xl mb-4" aria-hidden="true">⚠️</div>
          <h2 className="text-xl font-semibold text-foreground mb-2">
            エラーが発生しました
          </h2>
          <p className="text-muted-foreground mb-6">{error.message}</p>
          <Button variant="primary" onClick={() => window.location.reload()}>
            再読み込み
          </Button>
        </div>
      </Container>
    )
  }

  return (
    <Container size="xl" className="py-8">
      {/* ヘッダーの構成は今期／人気アニメの AnimeListPage と揃える */}
      <div className="pb-4 md:pb-[18px] border-b border-border mb-5 md:mb-6">
        <div className="flex items-baseline gap-2 md:gap-3 min-w-0">
          <h1 className="text-xl md:text-[26px] font-bold text-foreground tracking-tight truncate">
            お気に入りアニメ
          </h1>
          <span className="text-sm text-muted-foreground flex-shrink-0 flex items-baseline gap-1">
            {isLoading ? (
              '読み込み中...'
            ) : (
              <>
                <span className="text-base font-bold text-foreground">
                  {validFavorites.length.toLocaleString()}
                </span>
                件
              </>
            )}
          </span>
        </div>
      </div>

      {isLoading ? (
        // 読み込み中の見え方も一覧ページに合わせ、スケルトンをグリッドに敷く
        <div className="grid grid-cols-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-x-2.5 gap-y-3.5 md:gap-x-[18px] md:gap-y-5">
          {Array.from({ length: 12 }).map((_, i) => (
            <SkeletonCard key={i} />
          ))}
        </div>
      ) : validFavorites.length === 0 ? (
        <div className="md:min-h-[420px] md:flex md:items-center md:justify-center">
          <EmptyState
            icon={<Heart className="w-7 h-7 md:w-8 md:h-8" strokeWidth={1.75} aria-hidden="true" />}
            title="お気に入りはまだありません"
            description="好きな作品をお気に入りに追加しましょう"
            action={{
              label: 'アニメを探す',
              onClick: () => router.push('/search' as Route),
            }}
          />
        </div>
      ) : (
        <>
          <div className="grid grid-cols-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-x-2.5 gap-y-3.5 md:gap-x-[18px] md:gap-y-5">
            {visibleFavorites.map((anime) => (
              <FavoriteCard key={anime.id} anime={anime} />
            ))}
          </div>
          {hiddenCount > 0 && (
            <ExpandToggleButton isExpanded={isExpanded} hiddenCount={hiddenCount} onClick={toggleExpanded} />
          )}
        </>
      )}
    </Container>
  )
}

// お気に入りは認証が必須。未ログインでの直接アクセスは
// API がエラーになるため、マイページと同様にログイン画面へ送る。
const FavoritesPage: React.FC = () => {
  return (
    <ProtectedRoute>
      <FavoritesContent />
    </ProtectedRoute>
  )
}

export default FavoritesPage
