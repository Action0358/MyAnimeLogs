'use client'

import React from 'react'

/**
 * アニメカードのスケルトンローダー
 * ローディング中のプレースホルダーとして使用
 */
export const SkeletonCard: React.FC = () => (
  // 読み込み後のカードは枠を持たないため、スケルトンも同じ形にする
  <div className="animate-pulse">
    <div className="aspect-[2/3] bg-muted rounded-lg" />
    <div className="mt-2 sm:mt-2.5 min-h-[32px] sm:min-h-[38px]">
      <div className="h-[11px] sm:h-[13px] bg-muted rounded w-full" />
      <div className="h-[11px] sm:h-[13px] bg-muted rounded w-3/4 mt-1.5" />
    </div>
    <div className="h-[10px] sm:h-3 bg-muted rounded w-1/2 mt-1 sm:mt-1.5" />
  </div>
)

export default SkeletonCard
