'use client'

import React, { useRef, useEffect } from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import { useAuth } from '@/contexts/AuthContext'
import { USER_MENU_ITEMS } from './navigationConfig'

interface UserMenuProps {
  isOpen: boolean
  onClose: () => void
  /**
   * 開閉トグルのボタン。外側クリック判定から除外するために受け取る。
   * 渡さないと、開いている状態でアイコンを押したときに mousedown で閉じた直後
   * click でまた開き、アイコンを押しても閉じないように見える。
   */
  triggerRef?: React.RefObject<HTMLElement>
}

export const UserMenu: React.FC<UserMenuProps> = ({ isOpen, onClose, triggerRef }) => {
  const { user, logout } = useAuth()
  const menuRef = useRef<HTMLDivElement>(null)

  // メニュー外をクリックしたら閉じる
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
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [isOpen, onClose, triggerRef])

  if (!isOpen) return null

  const handleLogout = () => {
    logout()
    onClose()
  }

  return (
    <>
      {/*
        画面全体を覆う透明なレイヤー。メニューの外側であればどこを押しても閉じる。
        document のリスナーだけに頼ると、クリック先の要素や重なり順によっては
        閉じないことがあるため、受け止める実体を置いて確実にする。
        見た目を変えないよう背景色は持たせない。
      */}
      <div
        className="fixed inset-0 z-40"
        aria-hidden="true"
        onMouseDown={onClose}
        onTouchStart={onClose}
      />
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
    </>
  )
}
