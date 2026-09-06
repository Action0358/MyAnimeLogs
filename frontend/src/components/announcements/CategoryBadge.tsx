import React from 'react'
import { Badge, type BadgeProps } from '@/components/ui'
import {
  ANNOUNCEMENT_CATEGORY_LABELS,
  type AnnouncementCategory,
} from '@/constants/announcements'

/**
 * カテゴリごとのバッジ色。
 * 「重要」だけ目を引く色にし、他は情報量を抑えて本文の邪魔をしないようにする。
 */
const CATEGORY_VARIANTS: Record<AnnouncementCategory, BadgeProps['variant']> = {
  important: 'destructive',
  maintenance: 'warning',
  release: 'default',
  feature: 'secondary',
}

export const CategoryBadge: React.FC<{ category: AnnouncementCategory; className?: string }> = ({
  category,
  className,
}) => (
  <Badge variant={CATEGORY_VARIANTS[category]} className={className}>
    {ANNOUNCEMENT_CATEGORY_LABELS[category]}
  </Badge>
)

export default CategoryBadge
