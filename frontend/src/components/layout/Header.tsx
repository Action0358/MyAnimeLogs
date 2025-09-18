'use client'

import React from 'react'
import Link from 'next/link'
import Container from './Container'
import Button from '../ui/Button'

interface HeaderProps {
  user?: {
    id: number
    name: string
    email: string
  } | null
}

const Header: React.FC<HeaderProps> = ({ user }) => {
  return (
    <header className="bg-white shadow-sm border-b border-gray-200">
      <Container>
        <div className="flex items-center justify-between h-16">
          {/* ロゴ・ブランド */}
          <div className="flex items-center">
            <Link href="/" className="flex items-center space-x-2">
              <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
                <span className="text-white font-bold text-sm">MA</span>
              </div>
              <span className="text-xl font-bold text-gray-900">
                MyAnimeLogs
              </span>
            </Link>
          </div>

          {/* ナビゲーション */}
          <nav className="hidden md:flex items-center space-x-8">
            <Link
              href={'/animes' as any}
              className="text-gray-600 hover:text-gray-900 transition-colors"
            >
              アニメ
            </Link>
            <Link
              href={'/search' as any}
              className="text-gray-600 hover:text-gray-900 transition-colors"
            >
              検索
            </Link>
            {user && (
              <>
                <Link
                  href={'/mypage' as any}
                  className="text-gray-600 hover:text-gray-900 transition-colors"
                >
                  マイページ
                </Link>
                <Link
                  href={'/stats' as any}
                  className="text-gray-600 hover:text-gray-900 transition-colors"
                >
                  統計
                </Link>
              </>
            )}
          </nav>

          {/* ユーザーメニュー */}
          <div className="flex items-center space-x-4">
            {user ? (
              <div className="flex items-center space-x-3">
                <span className="text-sm text-gray-700">
                  こんにちは、{user.name}さん
                </span>
                <Button variant="outline" size="sm">
                  ログアウト
                </Button>
              </div>
            ) : (
              <div className="flex items-center space-x-3">
                <Link href={'/login' as any}>
                  <Button variant="ghost" size="sm">
                    ログイン
                  </Button>
                </Link>
                <Link href={'/register' as any}>
                  <Button variant="primary" size="sm">
                    新規登録
                  </Button>
                </Link>
              </div>
            )}
          </div>
        </div>
      </Container>
    </header>
  )
}

export default Header
