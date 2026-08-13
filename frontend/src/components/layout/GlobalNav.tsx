'use client'

import React from 'react'
import { usePathname } from 'next/navigation'
import { NavLink } from './NavLink'
import { NAV_ITEMS } from './navigationConfig'

export const GlobalNav: React.FC = () => {
  const pathname = usePathname()
  const isHomePage = pathname === '/'

  return (
    <nav className="hidden md:flex items-center gap-1.5" aria-label="グローバルナビゲーション">
      {NAV_ITEMS.map((item) => (
        <NavLink
          key={item.href}
          href={item.href}
          isHomePage={isHomePage}
        >
          {item.label}
        </NavLink>
      ))}
    </nav>
  )
}
