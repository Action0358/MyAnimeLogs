'use client'

import React from 'react'

/**
 * アニメカードのスケルトンローダー
 * ローディング中のプレースホルダーとして使用
 */
export const SkeletonCard: React.FC = () => (
  <div className="bg-card rounded-lg shadow-sm border border-border overflow-hidden animate-pulse">
    <div className="aspect-[2/3] bg-muted" />
    <div className="p-2 sm:p-3 h-[72px] sm:h-[100px] flex flex-col">
      <div className="h-4 bg-muted rounded w-full mb-1" />
      <div className="h-4 bg-muted rounded w-3/4 mb-2" />
      <div className="h-3 bg-muted rounded w-1/2 mt-auto" />
    </div>
  </div>
)

export default SkeletonCard
