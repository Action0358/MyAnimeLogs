'use client'

import React, { useRef, useEffect, useState } from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import { LogOut } from 'lucide-react'
import { useAuth } from '@/contexts/AuthContext'
import { USER_MENU_ITEMS } from './navigationConfig'
import { cn } from '@/lib/utils'

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
  const [imageError, setImageError] = useState(false)

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

  const handleLogout = () => {
    logout()
    onClose()
  }

  const avatarUrl = user?.avatar_url ?? null

  return (
    <>
      {/*
        画面全体を覆う透明なレイヤー。メニューの外側であればどこを押しても閉じる。
        document のリスナーだけに頼ると、クリック先の要素や重なり順によっては
        閉じないことがあるため、受け止める実体を置いて確実にする。
        見た目を変えないよう背景色は持たせない。isOpenがfalseの間は
        pointer-events-noneにして、裏の要素のクリックを邪魔しないようにする。
      */}
      <div
        className={cn('fixed inset-0 z-40', isOpen ? 'block' : 'hidden')}
        aria-hidden="true"
        onMouseDown={onClose}
        onTouchStart={onClose}
      />
      {/*
        常時マウントしてopacity/transformで開閉する（MobileNavと同じパターン）。
        isOpenでreturn nullすると、開閉のたびに瞬間表示/瞬間消失になり
        フェード・スライドのアニメーションが機能しない。
      */}
      <div
        ref={menuRef}
        role="menu"
        aria-hidden={!isOpen}
        className={cn(
          'absolute right-0 top-full z-50 mt-2 w-[248px] origin-top-right rounded-lg border border-border bg-popover p-1.5 text-popover-foreground shadow-lg transition-[opacity,transform] duration-150 ease-out',
          isOpen ? 'translate-y-0 opacity-100' : 'pointer-events-none -translate-y-1.5 opacity-0'
        )}
      >
        {/* ユーザー情報セクション。アバター＋名前をヘッダーアイコンと同じ見た目にして、
            誰のメニューかを一目で分かるようにする */}
        <div className="flex items-center gap-2.5 px-2.5 pb-3 pt-2">
          <span className="flex h-10 w-10 flex-shrink-0 items-center justify-center overflow-hidden rounded-full bg-primary/10 text-sm font-bold text-primary">
            {avatarUrl && !imageError ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img
                src={avatarUrl}
                alt=""
                className="h-full w-full object-cover"
                onError={() => setImageError(true)}
              />
            ) : (
              user?.name.charAt(0).toUpperCase()
            )}
          </span>
          <span className="min-w-0 truncate text-sm font-bold text-foreground">{user?.name}</span>
        </div>

        <div className="mx-1 mb-1.5 h-px bg-border" />

        {/* メインメニュー。項目の先頭にアイコンを出し、モバイルドロワーと見た目を揃える */}
        <nav>
          {USER_MENU_ITEMS.map((item) => {
            const Icon = item.icon
            return (
              <Link
                key={item.href}
                href={item.href as Route}
                onClick={onClose}
                role="menuitem"
                className="flex h-10 items-center gap-2.5 rounded-md px-2.5 text-sm font-medium text-foreground/90 transition-colors hover:bg-muted"
              >
                {Icon && (
                  <Icon className="h-[17px] w-[17px] text-muted-foreground" aria-hidden="true" />
                )}
                {item.label}
              </Link>
            )
          })}
        </nav>

        <div className="mx-1 my-1.5 h-px bg-border" />

        {/* ログアウトは操作の重さが他と違うため、赤系の色とホバーで分離する。
            高さ40pxで誤タップを防ぐ */}
        <button
          onClick={handleLogout}
          role="menuitem"
          className="flex h-10 w-full items-center gap-2.5 rounded-md px-2.5 text-left text-sm font-semibold text-destructive transition-colors hover:bg-destructive/5"
        >
          <LogOut className="h-[17px] w-[17px]" aria-hidden="true" />
          ログアウト
        </button>
      </div>
    </>
  )
}
