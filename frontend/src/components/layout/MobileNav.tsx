'use client'

import React, { useEffect, useRef } from 'react'
import { usePathname } from 'next/navigation'
import { NavLink } from './NavLink'
import { X } from 'lucide-react'
import { NAV_ITEMS, scrollToCurrentSeason } from './navigationConfig'

interface MobileNavProps {
  isOpen: boolean
  onClose: () => void
}

export const MobileNav: React.FC<MobileNavProps> = ({ isOpen, onClose }) => {
  const pathname = usePathname()
  const isHomePage = pathname === '/'
  const menuRef = useRef<HTMLDivElement>(null)

  const handleNavClick = (item: typeof NAV_ITEMS[0]) => (e: React.MouseEvent) => {
    if (item.requiresScroll && isHomePage) {
      e.preventDefault()
      scrollToCurrentSeason()
    }
    onClose()
  }

  // メニュー外クリックで閉じる
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        onClose()
      }
    }

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside)
      // スクロールを防止
      document.body.style.overflow = 'hidden'
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
      document.body.style.overflow = 'unset'
    }
  }, [isOpen, onClose])

  // ESCキーで閉じる
  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose()
      }
    }

    if (isOpen) {
      document.addEventListener('keydown', handleEscape)
    }

    return () => {
      document.removeEventListener('keydown', handleEscape)
    }
  }, [isOpen, onClose])

  if (!isOpen) return null

  return (
    <>
      {/* オーバーレイ */}
      <div
        className="fixed inset-0 bg-black bg-opacity-50 z-40 md:hidden"
        aria-hidden="true"
      />

      {/* サイドメニュー */}
      <div
        ref={menuRef}
        className="fixed inset-y-0 right-0 w-64 bg-white shadow-xl transform transition-transform duration-300 ease-in-out z-50 md:hidden"
        role="dialog"
        aria-modal="true"
        aria-label="モバイルナビゲーション"
      >
        {/* ヘッダー */}
        <div className="flex items-center justify-between p-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">メニュー</h2>
          <button
            onClick={onClose}
            className="p-2 rounded-md text-gray-400 hover:text-gray-500 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
            aria-label="メニューを閉じる"
          >
            <X className="h-6 w-6" />
          </button>
        </div>

        {/* ナビゲーションリンク */}
        <nav className="py-4" aria-label="モバイルナビゲーション">
          <ul className="space-y-1">
            {NAV_ITEMS.map((item) => (
              <li key={item.href}>
                <NavLink
                  href={item.href}
                  onClick={handleNavClick(item)}
                  className="block px-4 py-3 hover:bg-gray-50 rounded-none"
                >
                  {item.label}
                </NavLink>
              </li>
            ))}
          </ul>
        </nav>
      </div>
    </>
  )
}
