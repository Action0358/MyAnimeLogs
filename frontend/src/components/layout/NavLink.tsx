'use client'

import React from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import type { Route } from 'next'
import { cn } from '@/lib/utils'
import { NAV_TEXT_STYLES } from './navigationConfig'

interface NavLinkProps {
  href: string
  children: React.ReactNode
  onClick?: (e: React.MouseEvent) => void
  isHomePage?: boolean
  /** 'pill': デスクトップ横並びナビ（ピルタブ）/ 'drawer': モバイルドロワーの縦並び行 */
  variant?: 'pill' | 'drawer'
  className?: string
}

export const NavLink: React.FC<NavLinkProps> = ({
  href,
  children,
  onClick,
  isHomePage = false,
  variant = 'pill',
  className = '',
}) => {
  const pathname = usePathname()

  // アクティブ状態の判定
  const isActive = pathname === href || (href !== '/' && pathname.startsWith(href))

  const pillStyle = isHomePage
    ? isActive
      ? 'rounded-full px-3 py-1.5 bg-white/20 text-white font-semibold'
      : 'rounded-full px-3 py-1.5 text-white/85 hover:text-white'
    : isActive
      ? 'rounded-full px-3 py-1.5 bg-primary/10 text-primary font-semibold'
      : 'rounded-full px-3 py-1.5 text-muted-foreground hover:text-foreground'

  const drawerStyle = isActive
    ? 'rounded-lg px-3 py-2.5 bg-primary/10 text-primary font-semibold'
    : 'rounded-lg px-3 py-2.5 text-foreground hover:bg-muted'

  return (
    <Link
      href={href as Route}
      onClick={onClick}
      className={cn(
        NAV_TEXT_STYLES,
        'transition-colors',
        variant === 'drawer' ? drawerStyle : pillStyle,
        className
      )}
      aria-current={isActive ? 'page' : undefined}
    >
      {children}
    </Link>
  )
}
