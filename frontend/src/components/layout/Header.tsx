'use client'

import React, { useState } from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { Menu } from 'lucide-react'
import { Button } from '../ui/Button'
import { useAuth } from '@/contexts/AuthContext'
import { UserMenu } from './UserMenu'
import { GlobalNav } from './GlobalNav'
import { MobileNav } from './MobileNav'

const Header: React.FC = () => {
  const { user, isLoading } = useAuth()
  const [isUserMenuOpen, setIsUserMenuOpen] = useState(false)
  const [isMobileNavOpen, setIsMobileNavOpen] = useState(false)
  const [imageError, setImageError] = useState(false)
  const pathname = usePathname()

  const isHomePage = pathname === '/'

  const authButtonClassName =
    'text-xs sm:text-sm px-2 py-1 sm:px-4 sm:py-2 whitespace-nowrap'

  const toggleUserMenu = () => {
    setIsUserMenuOpen(!isUserMenuOpen)
  }

  const toggleMobileNav = () => {
    setIsMobileNavOpen(!isMobileNavOpen)
  }

  return (
    <header className={isHomePage ? 'bg-transparent absolute top-0 left-0 right-0 z-10' : 'bg-gray-50'}>
      <div className="flex items-center justify-between h-20 sm:h-24 px-4 sm:px-6 lg:px-8">
        {/* 左側: ロゴ + グローバルナビゲーション */}
        <div className="flex items-center gap-6 flex-1">
          {/* ロゴ・ブランド */}
          <Link href="/" className="flex items-center group flex-shrink-0">
            <img
              src="/logo.svg"
              alt="MyAnimeLogs"
              className={`h-10 sm:h-12 w-auto transition-transform duration-300 group-hover:scale-105 ${isHomePage ? 'brightness-0 invert' : ''}`}
            />
          </Link>

          {/* グローバルナビゲーション（デスクトップ） */}
          <GlobalNav />
        </div>

        {/* 右側のメニュー */}
        <div className="flex items-center gap-2 sm:gap-3">
          {/* モバイルメニューボタン */}
          <button
            onClick={toggleMobileNav}
            className={`md:hidden p-2 rounded-md transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 ${
              isHomePage
                ? 'text-white hover:bg-white/10 focus:ring-white'
                : 'text-gray-700 hover:bg-gray-100 focus:ring-blue-500'
            }`}
            aria-label="メニューを開く"
            aria-expanded={isMobileNavOpen}
          >
            <Menu className="h-6 w-6" />
          </button>
          {isLoading ? (
            <div className="flex items-center space-x-2">
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600"></div>
              <span className="text-sm text-gray-500 hidden sm:inline">
                読み込み中...
              </span>
            </div>
          ) : user ? (
            <div className="relative flex items-center gap-2">
              {/* ユーザーアイコン（クリックでメニュー表示） */}
              <button
                onClick={toggleUserMenu}
                className="w-10 h-10 rounded-full bg-blue-100 flex items-center justify-center hover:bg-blue-200 transition-colors cursor-pointer focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                aria-label="ユーザーメニュー"
                aria-expanded={isUserMenuOpen}
              >
                {user.icon && !imageError ? (
                  <img
                    src={`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}${user.icon}`}
                    alt={user.name}
                    className="w-10 h-10 rounded-full object-cover"
                    onError={() => setImageError(true)}
                  />
                ) : (
                  <span className="text-blue-600 font-medium text-base">
                    {user.name.charAt(0).toUpperCase()}
                  </span>
                )}
              </button>

              {/* ドロップダウンメニュー */}
              <UserMenu
                isOpen={isUserMenuOpen}
                onClose={() => setIsUserMenuOpen(false)}
              />
            </div>
          ) : (
            <div className="flex items-center gap-1.5 sm:gap-3">
              <Link href="/login">
                <Button
                  variant="ghost"
                  size="sm"
                  className={`${authButtonClassName} ${isHomePage ? 'text-white border-white hover:bg-white/10' : ''}`}
                >
                  ログイン
                </Button>
              </Link>
              <Link href="/register">
                <Button
                  variant="primary"
                  size="sm"
                  className={`${authButtonClassName} ${isHomePage ? 'bg-white text-blue-600 hover:bg-gray-100' : ''}`}
                >
                  新規登録
                </Button>
              </Link>
            </div>
          )}
        </div>
      </div>

      {/* モバイルナビゲーション */}
      <MobileNav isOpen={isMobileNavOpen} onClose={() => setIsMobileNavOpen(false)} />
    </header>
  )
}

export default Header
