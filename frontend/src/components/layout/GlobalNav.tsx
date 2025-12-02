'use client'

import React from 'react'
import { usePathname } from 'next/navigation'
import { NavLink } from './NavLink'
import { NAV_ITEMS, scrollToCurrentSeason } from './navigationConfig'

export const GlobalNav: React.FC = () => {
  const pathname = usePathname()
  const isHomePage = pathname === '/'

  const handleNavClick = (item: typeof NAV_ITEMS[0]) => (e: React.MouseEvent) => {
    if (item.requiresScroll && isHomePage) {
      e.preventDefault()
      scrollToCurrentSeason()
    }
  }

  return (
    <nav className="hidden md:flex items-center space-x-4" aria-label="グローバルナビゲーション">
      {NAV_ITEMS.map((item) => (
        <NavLink
          key={item.href}
          href={item.href}
          onClick={handleNavClick(item)}
          isHomePage={isHomePage}
        >
          {item.label}
        </NavLink>
      ))}
    </nav>
  )
}
