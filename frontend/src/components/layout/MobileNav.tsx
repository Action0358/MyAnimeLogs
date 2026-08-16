'use client'

import React, { useEffect, useRef } from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import { ChevronRight, LogOut, X } from 'lucide-react'
import { NavLink } from './NavLink'
import { HOME_NAV_ITEM, NAV_ITEMS, USER_MENU_ITEMS, NAV_TEXT_STYLES } from './navigationConfig'
import { useAuth } from '@/contexts/AuthContext'
import { resolveAvatarUrl } from '@/lib/avatars'
import { cn } from '@/lib/utils'

interface MobileNavProps {
  isOpen: boolean
  onClose: () => void
  /**
   * 開閉トグルのボタン。外側クリック判定から除外するために受け取る。
   * 渡さないと、開いている状態でトグルを押したときに mousedown で閉じた直後
   * click でまた開き、ボタンを押しても閉じないように見える。
   */
  triggerRef?: React.RefObject<HTMLElement>
}

/** セクションの見出しラベル。単なる区切り線より、group の意味が伝わる */
const SectionLabel: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <div className="px-2 pb-1.5 text-[11px] font-bold tracking-[0.06em] text-muted-foreground">
    {children}
  </div>
)

export const MobileNav: React.FC<MobileNavProps> = ({ isOpen, onClose, triggerRef }) => {
  const menuRef = useRef<HTMLDivElement>(null)
  const { user, logout } = useAuth()

  // メニュー外クリックで閉じる
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Node
      // トグル自身のクリックはここで閉じない。閉じる処理はボタンの onClick に任せる
      if (triggerRef?.current?.contains(target)) return
      if (menuRef.current && !menuRef.current.contains(target)) {
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
  }, [isOpen, onClose, triggerRef])

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

  const handleLogout = () => {
    logout()
    onClose()
  }

  const avatarUrl = user ? resolveAvatarUrl(user.icon) : null

  return (
    <>
      {/* オーバーレイ */}
      {/* ドロワーの外側であればどこを押しても閉じる。
          document のリスナー任せにせず、受け止める実体側でも閉じる */}
      <div
        className="fixed inset-0 bg-black bg-opacity-50 z-40 md:hidden"
        aria-hidden="true"
        onMouseDown={onClose}
        onTouchStart={onClose}
      />

      {/* サイドメニュー（左側から表示） */}
      <div
        ref={menuRef}
        className="fixed inset-y-0 left-0 flex w-72 flex-col bg-background shadow-xl transform transition-transform duration-300 ease-in-out z-50 md:hidden"
        role="dialog"
        aria-modal="true"
        aria-label="モバイルナビゲーション"
      >
        {/* ヘッダー。ブランド名で今どこにいるかを示し、閉じる操作を明示する */}
        <div className="flex flex-shrink-0 items-center justify-between px-4 pb-3.5 pt-4">
          <span className="text-base font-bold tracking-tight text-foreground">MyAnimeLogs</span>
          <button
            onClick={onClose}
            className="-mr-1.5 rounded-full p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            aria-label="メニューを閉じる"
          >
            <X className="h-[18px] w-[18px]" />
          </button>
        </div>

        {/*
          ユーザー名を出すだけだと行き止まりになるため、マイページへのカードにする。
          アイコン・名前・遷移先を1つのタップ領域にまとめる。
        */}
        {user && (
          <div className="flex-shrink-0 border-b border-border px-3 pb-3.5">
            <Link
              href={'/mypage' as Route}
              onClick={onClose}
              className="flex items-center gap-3 rounded-lg bg-muted/50 px-2 py-2.5 transition-colors hover:bg-muted focus:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            >
              <span className="flex h-[38px] w-[38px] flex-shrink-0 items-center justify-center overflow-hidden rounded-full bg-primary/10 text-sm font-bold text-primary">
                {avatarUrl ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img src={avatarUrl} alt="" className="h-full w-full object-cover" />
                ) : (
                  user.name.charAt(0).toUpperCase()
                )}
              </span>
              <span className="min-w-0">
                <span className="block truncate text-[13.5px] font-bold leading-tight text-foreground">
                  {user.name}
                </span>
                <span className="mt-px block text-[11px] text-muted-foreground">
                  マイページを見る
                </span>
              </span>
              <ChevronRight
                className="ml-auto h-4 w-4 flex-shrink-0 text-muted-foreground"
                aria-hidden="true"
              />
            </Link>
          </div>
        )}

        {/* ナビゲーションリンク */}
        <nav className="flex-1 overflow-y-auto px-3 py-3.5" aria-label="モバイルナビゲーション">
          <SectionLabel>メイン</SectionLabel>
          <ul className="space-y-0.5">
            {[HOME_NAV_ITEM, ...NAV_ITEMS].map((item) => (
              <li key={item.href}>
                <NavLink
                  href={item.href}
                  onClick={onClose}
                  variant="drawer"
                  icon={item.icon}
                  className="w-full"
                >
                  {item.label}
                </NavLink>
              </li>
            ))}
          </ul>

          {/* ログイン中のみ。ヘッダーのユーザーアイコンを廃したぶんをここに集約する */}
          {user && (
            <>
              <div className="pt-3.5">
                <SectionLabel>マイアカウント</SectionLabel>
              </div>
              <ul className="space-y-0.5">
                {USER_MENU_ITEMS.map((item) => (
                  <li key={item.href}>
                    <NavLink
                      href={item.href}
                      onClick={onClose}
                      variant="drawer"
                      icon={item.icon}
                      className="w-full"
                    >
                      {item.label}
                    </NavLink>
                  </li>
                ))}
              </ul>
            </>
          )}
        </nav>

        {/* フッター。ログアウトは他のリンクと役割が違うため色で分け、最下部に隔離する */}
        <div className="flex-shrink-0 border-t border-border px-3 pb-4 pt-2.5">
          {user ? (
            <button
              onClick={handleLogout}
              className={cn(
                NAV_TEXT_STYLES,
                'flex w-full items-center gap-3 rounded-lg px-2 py-2.5 font-semibold text-destructive transition-colors hover:bg-destructive/5 focus:outline-none focus-visible:ring-2 focus-visible:ring-ring'
              )}
            >
              <LogOut className="h-[18px] w-[18px] flex-shrink-0" aria-hidden="true" />
              ログアウト
            </button>
          ) : (
            /*
              Google認証ではログインと新規登録が同じ操作になるため、ボタンは1つ。
              ヘッダーから移設し、ヒーロー画像の上に単独で浮かないようにする。
            */
            <Link
              href={'/signin' as Route}
              onClick={onClose}
              className="flex h-11 items-center justify-center rounded-md bg-primary px-4 text-sm font-semibold text-primary-foreground transition-colors hover:bg-primary/90 focus:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            >
              ログイン
            </Link>
          )}
        </div>
      </div>
    </>
  )
}
