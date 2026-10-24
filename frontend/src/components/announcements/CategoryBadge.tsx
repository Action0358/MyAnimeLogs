import React from 'react'
import { cn } from '@/lib/utils'
import {
  ANNOUNCEMENT_CATEGORY_LABELS,
  type AnnouncementCategory,
} from '@/constants/announcements'

/**
 * カテゴリのタグ意匠。
 *
 * 淡い背景に濃い文字（塗りつぶしではない）とし、角は4px。タイトルより先に
 * 目に入らないようにするため。モックアップの指定に合わせている。
 *
 * 「重要」だけは他を上回って目を引く必要があるため赤系にする。
 * リリース・機能は「新しくなった」知らせなのでどちらもプライマリ系で揃え、
 * メンテナンスは予定された作業の案内なので、
 * 赤・琥珀のような警告色は使わず、落ち着いたスレートにする。
 */
const CATEGORY_CLASSES: Record<AnnouncementCategory, string> = {
  important: 'bg-destructive/10 text-destructive',
  maintenance: 'bg-slate-100 text-slate-700',
  release: 'bg-primary/10 text-primary',
  feature: 'bg-primary/10 text-primary',
}

export interface CategoryBadgeProps {
  category: AnnouncementCategory
  /** 'sm' はトップページのリスト行用。密度が高いぶん一段小さくする */
  size?: 'sm' | 'md'
  className?: string
}

export const CategoryBadge: React.FC<CategoryBadgeProps> = ({
  category,
  size = 'md',
  className,
}) => (
  <span
    className={cn(
      'inline-flex flex-shrink-0 items-center rounded font-bold',
      size === 'sm' ? 'px-2 py-0.5 text-[10px] md:text-[11px]' : 'px-2.5 py-0.5 text-xs',
      CATEGORY_CLASSES[category],
      className,
    )}
  >
    {ANNOUNCEMENT_CATEGORY_LABELS[category]}
  </span>
)

export default CategoryBadge
