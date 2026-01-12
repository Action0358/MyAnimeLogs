'use client'

import React, { useEffect, useRef } from 'react'
import { NavLink } from './NavLink'
import { X } from 'lucide-react'
import { NAV_ITEMS } from './navigationConfig'

interface MobileNavProps {
  isOpen: boolean
  onClose: () => void
}

export const MobileNav: React.FC<MobileNavProps> = ({ isOpen, onClose }) => {
  const menuRef = useRef<HTMLDivElement>(null)

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

      {/* サイドメニュー（左側から表示） */}
      <div
        ref={menuRef}
        className="fixed inset-y-0 left-0 w-64 bg-white shadow-xl transform transition-transform duration-300 ease-in-out z-50 md:hidden"
        role="dialog"
        aria-modal="true"
        aria-label="モバイルナビゲーション"
      >
        {/* ヘッダー */}
        <div className="flex items-center px-4 py-4 border-b border-gray-200">
          <button
            onClick={onClose}
            className="-ml-2 p-2 rounded-md text-gray-400 hover:text-gray-500 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
            aria-label="メニューを閉じる"
          >
            <X className="h-6 w-6" />
          </button>
        </div>

        {/* ナビゲーションリンク */}
        <nav className="py-4" aria-label="モバイルナビゲーション">
          <ul className="space-y-1">
            {/* ホームリンク（モバイルのみ） */}
            <li>
              <NavLink
                href="/"
                onClick={onClose}
                className="block pl-5 pr-4 py-3 hover:bg-gray-50 rounded-none"
              >
                ホーム
              </NavLink>
            </li>
            {NAV_ITEMS.map((item) => (
              <li key={item.href}>
                <NavLink
                  href={item.href}
                  onClick={onClose}
                  className="block pl-5 pr-4 py-3 hover:bg-gray-50 rounded-none"
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
