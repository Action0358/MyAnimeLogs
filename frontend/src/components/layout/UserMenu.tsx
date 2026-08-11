'use client'

import React, { useRef, useEffect } from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import { useAuth } from '@/contexts/AuthContext'
import { USER_MENU_ITEMS } from './navigationConfig'

interface UserMenuProps {
  isOpen: boolean
  onClose: () => void
}

export const UserMenu: React.FC<UserMenuProps> = ({ isOpen, onClose }) => {
  const { user, logout } = useAuth()
  const menuRef = useRef<HTMLDivElement>(null)

  // メニュー外をクリックしたら閉じる
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        onClose()
      }
    }

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside)
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [isOpen, onClose])

  if (!isOpen) return null

  const handleLogout = () => {
    logout()
    onClose()
  }

  return (
    <div
      ref={menuRef}
      className="absolute right-0 top-full mt-2 w-64 bg-popover text-popover-foreground rounded-lg shadow-lg border border-border py-2 z-50"
    >
      {/* ユーザー情報セクション */}
      <div className="px-4 py-3 border-b border-border">
        <p className="text-base font-bold text-foreground">{user?.name}</p>
      </div>

      {/* メインメニュー */}
      <nav className="py-1">
        {USER_MENU_ITEMS.map((item) => (
          <Link
            key={item.href}
            href={item.href as Route}
            onClick={onClose}
            className="block px-4 py-2 text-foreground hover:bg-muted transition-colors"
          >
            <span className="text-base font-semibold">{item.label}</span>
          </Link>
        ))}
      </nav>

      {/* ログアウト */}
      <div className="border-t border-border py-1">
        <button
          onClick={handleLogout}
          className="block w-full text-left px-4 py-2 text-foreground hover:bg-muted transition-colors"
        >
          <span className="text-base font-semibold">ログアウト</span>
        </button>
      </div>
    </div>
  )
}
