'use client'

import React, { useEffect, useRef } from 'react'
import { NavLink } from './NavLink'
import { X } from 'lucide-react'
import { NAV_ITEMS, USER_MENU_ITEMS } from './navigationConfig'
import { useAuth } from '@/contexts/AuthContext'

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
        className="fixed inset-y-0 left-0 w-64 bg-background shadow-xl transform transition-transform duration-300 ease-in-out z-50 md:hidden"
        role="dialog"
        aria-modal="true"
        aria-label="モバイルナビゲーション"
      >
        {/* ヘッダー */}
        <div className="flex items-center px-4 py-4 border-b border-border">
          <button
            onClick={onClose}
            // 開くボタンと同じ48pxのタップ領域に揃える
            className="-ml-3 p-3 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted focus:outline-none focus:ring-2 focus:ring-ring"
            aria-label="メニューを閉じる"
          >
            <X className="h-6 w-6" />
          </button>
        </div>

        {/* ナビゲーションリンク */}
        <nav className="px-3 py-4" aria-label="モバイルナビゲーション">
          <ul className="space-y-1">
            {/* ホームリンク（モバイルのみ） */}
            <li>
              <NavLink href="/" onClick={onClose} variant="drawer" className="block">
                ホーム
              </NavLink>
            </li>
            {NAV_ITEMS.map((item) => (
              <li key={item.href}>
                <NavLink href={item.href} onClick={onClose} variant="drawer" className="block">
                  {item.label}
                </NavLink>
              </li>
            ))}
          </ul>

          {/* ログイン中のみ。ヘッダーのユーザーアイコンを廃したぶんをここに集約する */}
          {user && (
            <>
              <div className="mt-4 pt-4 border-t border-border">
                <p className="px-3 pb-2 text-sm font-bold text-foreground truncate">
                  {user.name}
                </p>
                <ul className="space-y-1">
                  {USER_MENU_ITEMS.map((item) => (
                    <li key={item.href}>
                      <NavLink
                        href={item.href}
                        onClick={onClose}
                        variant="drawer"
                        className="block"
                      >
                        {item.label}
                      </NavLink>
                    </li>
                  ))}
                </ul>
              </div>

              <div className="mt-4 pt-4 border-t border-border">
                <button
                  onClick={handleLogout}
                  className="block w-full text-left px-3 py-2 rounded-md text-foreground hover:bg-muted transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                >
                  ログアウト
                </button>
              </div>
            </>
          )}
        </nav>
      </div>
    </>
  )
}
