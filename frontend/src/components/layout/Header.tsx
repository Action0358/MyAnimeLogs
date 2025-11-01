'use client'

import React, { useState } from 'react'
import Link from 'next/link'
import Image from 'next/image'
import { Button } from '../ui/Button'
import { useAuth } from '@/contexts/AuthContext'
import { UserMenu } from './UserMenu'

const Header: React.FC = () => {
  const { user, isLoading } = useAuth()
  const [isMenuOpen, setIsMenuOpen] = useState(false)
  const [imageError, setImageError] = useState(false)

  const authButtonClassName =
    'text-xs sm:text-sm px-2 py-1 sm:px-4 sm:py-2 whitespace-nowrap'

  const toggleMenu = () => {
    setIsMenuOpen(!isMenuOpen)
  }

  return (
    <header className="bg-white shadow-sm border-b border-gray-200">
      <div className="flex items-center justify-between h-22 sm:h-28 px-4 sm:px-6 lg:px-8">
        {/* ロゴ・ブランド */}
        <div className="flex items-center flex-shrink-0">
          <Link href="/" className="flex items-center group">
            <div className="relative h-18 sm:h-24 w-auto transition-transform duration-300 group-hover:scale-105">
              <Image
                src="/logo.png"
                alt="MyAnimeLogs"
                width={400}
                height={80}
                className="h-full w-auto object-contain"
                priority
              />
            </div>
          </Link>
        </div>

        {/* ナビゲーション */}
        <nav className="hidden md:flex items-center space-x-8">
          {/* 将来的にナビゲーションリンクを追加する場合はここに配置 */}
        </nav>

        {/* ユーザーメニュー */}
        <div className="flex items-center gap-2 sm:gap-3">
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
                onClick={toggleMenu}
                className="w-10 h-10 rounded-full bg-blue-100 flex items-center justify-center hover:bg-blue-200 transition-colors cursor-pointer focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                aria-label="ユーザーメニュー"
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
                isOpen={isMenuOpen}
                onClose={() => setIsMenuOpen(false)}
              />
            </div>
          ) : (
            <div className="flex items-center gap-1.5 sm:gap-3">
              <Link href="/register">
                <Button
                  variant="primary"
                  size="sm"
                  className={authButtonClassName}
                >
                  新規登録
                </Button>
              </Link>
              <Link href="/login">
                <Button
                  variant="ghost"
                  size="sm"
                  className={authButtonClassName}
                >
                  ログイン
                </Button>
              </Link>
            </div>
          )}
        </div>
      </div>
    </header>
  )
}

export default Header
