'use client'

import React from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'

interface NavLinkProps {
  href: string
  children: React.ReactNode
  onClick?: (e: React.MouseEvent) => void
  isHomePage?: boolean
  className?: string
}

export const NavLink: React.FC<NavLinkProps> = ({
  href,
  children,
  onClick,
  isHomePage = false,
  className = '',
}) => {
  const pathname = usePathname()
  
  // アクティブ状態の判定
  const isActive = pathname === href || (href !== '/' && pathname.startsWith(href))
  
  // ホームページでの白文字スタイル
  const baseStyle = isHomePage
    ? 'text-white hover:text-white/80'
    : 'text-gray-700 hover:text-gray-900'
  
  // アクティブ状態のスタイル
  const activeStyle = isHomePage
    ? 'border-b-2 border-white'
    : 'text-blue-600 border-b-2 border-blue-600'
  
  const linkClassName = `
    px-2 py-2 text-base font-bold transition-colors
    ${baseStyle}
    ${isActive ? activeStyle : ''}
    ${className}
  `.trim()

  return (
    <Link
      href={href as any}
      onClick={onClick}
      className={linkClassName}
      aria-current={isActive ? 'page' : undefined}
    >
      {children}
    </Link>
  )
}
