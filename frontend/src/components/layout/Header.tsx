'use client'

import React from 'react'
import Link from 'next/link'
import Container from './Container'
import { Button } from '../ui/Button'
import { useAuth } from '@/contexts/AuthContext'

const Header: React.FC = () => {
  const { user, isLoading, logout } = useAuth()

  const handleLogout = () => {
    logout()
  }
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
            {/* 将来的にナビゲーションリンクを追加する場合はここに配置 */}
          </nav>

          {/* ユーザーメニュー */}
          <div className="flex items-center space-x-4">
            {isLoading ? (
              <div className="flex items-center space-x-2">
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600"></div>
                <span className="text-sm text-gray-500">読み込み中...</span>
              </div>
            ) : user ? (
              <div className="flex items-center space-x-3">
                <Button variant="ghost" size="sm" onClick={handleLogout}>
                  ログアウト
                </Button>
              </div>
            ) : (
              <div className="flex items-center space-x-3">
                <Link href="/register">
                  <Button variant="primary" size="sm">
                    新規登録
                  </Button>
                </Link>
                <Link href="/login">
                  <Button variant="ghost" size="sm">
                    ログイン
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
