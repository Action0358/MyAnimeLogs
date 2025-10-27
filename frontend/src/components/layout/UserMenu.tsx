'use client'

import React, { useRef, useEffect } from 'react'
import Link from 'next/link'
import { useAuth } from '@/contexts/AuthContext'

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
      className="absolute right-0 top-full mt-2 w-64 bg-white rounded-lg shadow-lg border border-gray-200 py-2 z-50"
    >
      {/* ユーザー情報セクション */}
      <div className="px-4 py-3 border-b border-gray-200">
        <p className="text-base font-bold text-blue-900">{user?.name}</p>
      </div>

      {/* メインメニュー */}
      <nav className="py-1">
        <Link
          href="/mypage"
          onClick={onClose}
          className="block px-4 py-2 text-gray-700 hover:bg-gray-50 transition-colors"
        >
          <span className="text-base font-semibold">マイページ</span>
        </Link>
        <Link
          href="/stats"
          onClick={onClose}
          className="block px-4 py-2 text-gray-700 hover:bg-gray-50 transition-colors"
        >
          <span className="text-base font-semibold">統計</span>
        </Link>
        <Link
          href="/mypage"
          onClick={onClose}
          className="block px-4 py-2 text-gray-700 hover:bg-gray-50 transition-colors"
        >
          <span className="text-base font-semibold">視聴履歴</span>
        </Link>
      </nav>

      {/* ログアウト */}
      <div className="border-t border-gray-200">
        <button
          onClick={handleLogout}
          className="block w-full text-left px-4 py-2 text-gray-700 hover:bg-gray-50 transition-colors mt-2"
        >
          <span className="text-base font-semibold">ログアウト</span>
        </button>
      </div>
    </div>
  )
}
