'use client'

import React, { useRef, useState } from 'react'
import Link from 'next/link'
import Image from 'next/image'
import { usePathname } from 'next/navigation'
import type { Route } from 'next'
import { Menu } from 'lucide-react'
import { cn } from '@/lib/utils'
import { SUB_LOGO_PATH } from '@/lib/constants'
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
  // 外側クリック判定からトグル自身を除外するために渡す
  const mobileNavTriggerRef = useRef<HTMLButtonElement>(null)
  const userMenuTriggerRef = useRef<HTMLButtonElement>(null)
  const pathname = usePathname()

  const isHomePage = pathname === '/'

  const toggleUserMenu = () => {
    setIsUserMenuOpen(!isUserMenuOpen)
  }

  const toggleMobileNav = () => {
    setIsMobileNavOpen(!isMobileNavOpen)
  }

  return (
    <header
      className={cn(
        'h-20 sm:h-24',
        isHomePage
          ? 'bg-transparent absolute top-0 left-0 right-0 z-10'
          : 'bg-background border-b border-border'
      )}
    >
      {/* ヘッダーは本文の列（最大1180px）に収めず、画面幅いっぱいに置く（左右40px、SPは16px）。
          トップではヒーロー画像が全幅に広がるため、ヘッダーだけ列に収めると広い画面で
          左右に大きな空きができ、内側に寄って見える（モックアップ「Header Alignment
          Redesign」）。白ヘッダーの通常ページも同じ位置にし、遷移でロゴが動かないようにする */}
      <div className="flex items-center justify-between h-20 sm:h-24 px-4 md:px-10">
        {/* 左側: ハンバーガーメニュー（モバイル）/ ロゴ + グローバルナビゲーション（デスクトップ） */}
        <div className="flex items-center gap-6 flex-1">
          {/* モバイルメニューボタン（左側に配置） */}
          <button
            ref={mobileNavTriggerRef}
            onClick={toggleMobileNav}
            className={cn(
              // タップ領域を48pxまで広げる。負のマージンでアイコンの見た目の位置は変えない
              'md:hidden p-3 -m-1 rounded-md transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2',
              isHomePage
                ? 'text-white hover:bg-white/10 focus:ring-white'
                : 'text-foreground hover:bg-muted focus:ring-ring'
            )}
            aria-label="メニューを開く"
            aria-expanded={isMobileNavOpen}
          >
            <Menu className="h-6 w-6" />
          </button>

          {/* ロゴ・ブランド（デスクトップのみ表示） */}
          <Link href={"/" as Route} className="hidden md:flex items-center group flex-shrink-0">
            <Image
              src={SUB_LOGO_PATH}
              alt="MyAnimeLogs"
              width={1792}
              height={576}
              priority
              className={`h-10 sm:h-12 w-auto transition-transform duration-300 group-hover:scale-105 ${isHomePage ? 'brightness-0 invert' : ''}`}
            />
          </Link>

          {/* グローバルナビゲーション（デスクトップ） */}
          <GlobalNav />
        </div>

        {/* 右側のメニュー */}
        <div className="flex items-center gap-2 sm:gap-3">
          {isLoading ? (
            <div className="flex items-center space-x-2">
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-primary"></div>
              <span className="text-sm text-muted-foreground hidden sm:inline">
                読み込み中...
              </span>
            </div>
          ) : user ? (
            // 狭い画面ではハンバーガーメニューに統合するため、アイコンを出さない
            <div className="relative hidden md:flex items-center gap-2">
              {/* ユーザーアイコン（クリックでメニュー表示） */}
              <button
                ref={userMenuTriggerRef}
                onClick={toggleUserMenu}
                className={cn(
                  'w-12 h-12 rounded-full flex items-center justify-center transition-colors cursor-pointer',
                  'focus:outline-none focus:ring-2 focus:ring-offset-2',
                  // ホームはヘッダーが透過してヒーロー画像に重なる。
                  // primary系の淡い色だと背景に埋もれるため、白の縁と塗りで分離する
                  isHomePage
                    ? 'bg-white/20 ring-1 ring-white/70 hover:bg-white/30 focus:ring-white'
                    : 'bg-primary/10 ring-1 ring-border hover:bg-primary/20 focus:ring-ring'
                )}
                aria-label="ユーザーメニュー"
                aria-expanded={isUserMenuOpen}
              >
                {user.avatar_url && !imageError ? (
                  <img
                    src={user.avatar_url}
                    alt={user.name}
                    className="w-10 h-10 rounded-full object-cover"
                    onError={() => setImageError(true)}
                  />
                ) : (
                  <span
                    className={cn(
                      'font-bold text-base',
                      isHomePage ? 'text-white' : 'text-primary'
                    )}
                  >
                    {user.name.charAt(0).toUpperCase()}
                  </span>
                )}
              </button>

              {/* ドロップダウンメニュー */}
              <UserMenu
                isOpen={isUserMenuOpen}
                onClose={() => setIsUserMenuOpen(false)}
                triggerRef={userMenuTriggerRef}
              />
            </div>
          ) : (
            /*
              Google認証ではログインと新規登録が同じ操作になるため、
              ボタンは1つに統合し /signin へ送る。
              狭い画面でもヘッダー右上に出す（ドロワーのフッターにも残す）。
              ログインは記録・お気に入りの入口で、ドロワーの中だけだと
              開くまで存在に気づけないため。ヒーロー上では白地のボタンに
              して浮かないようにする（モックアップ「Header Alignment Redesign」SP）
            */
            <div className="flex items-center gap-2 sm:gap-3">
              <Link href={"/signin" as Route}>
                <Button
                  variant={isHomePage ? 'primary-on-dark' : 'primary'}
                  className={cn(
                    'h-auto text-xs px-[13px] py-[7px] font-bold',
                    'md:h-11 md:text-base md:px-3 md:py-2'
                  )}
                >
                  ログイン
                </Button>
              </Link>
            </div>
          )}
        </div>
      </div>

      {/* モバイルナビゲーション */}
      <MobileNav
        isOpen={isMobileNavOpen}
        onClose={() => setIsMobileNavOpen(false)}
        triggerRef={mobileNavTriggerRef}
      />
    </header>
  )
}

export default Header
