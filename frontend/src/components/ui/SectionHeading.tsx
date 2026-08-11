import React from 'react'
import { cn } from '@/lib/utils'

interface SectionHeadingProps extends React.HTMLAttributes<HTMLHeadingElement> {
  /** 見出しレベル。ページ内の階層に合わせて指定する（既定: h2） */
  as?: 'h2' | 'h3'
  /** 見出しの右に置く要素（「もっと見る」ボタンなど） */
  action?: React.ReactNode
  /** 文字サイズ。カード内など密度の高い場所では 'sm' を使う */
  size?: 'sm' | 'md'
}

/**
 * セクション見出し。primaryのアクセントバーを左に添える。
 * Home（AnimeSection）とアニメ詳細ページで共通利用し、見出しの見え方を統一する。
 */
export const SectionHeading: React.FC<SectionHeadingProps> = ({
  as: Tag = 'h2',
  action,
  size = 'md',
  className,
  children,
  ...props
}) => (
  <div className={cn('flex items-end justify-between gap-4', action ? 'mb-5' : 'mb-3')}>
    <Tag
      className={cn(
        'relative pl-3 font-bold text-foreground tracking-tight',
        'before:absolute before:left-0 before:top-1 before:bottom-1 before:w-1 before:rounded-full before:bg-primary',
        size === 'md' ? 'text-xl md:text-2xl' : 'text-base md:text-lg',
        className,
      )}
      {...props}
    >
      {children}
    </Tag>
    {action && <div className="flex-shrink-0">{action}</div>}
  </div>
)

export default SectionHeading
