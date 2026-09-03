'use client'

import React, { useState } from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import { Container } from '@/components/layout'
import { useAuth } from '@/contexts/AuthContext'
import { deleteAccount } from '@/lib/api/account'
import { DeleteAccountDialog } from './DeleteAccountDialog'

/**
 * 設定画面。
 *
 * #789 でプロフィールの自由記述項目を廃止したため、編集フォームは持たない。
 * 表示名とアイコンはGoogleアカウント由来で固定（#779）。
 * アカウント削除の導線がここにあるため、画面自体は残している。
 */
export const SettingsContent: React.FC = () => {
  const { user, logout } = useAuth()
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  const [deleteError, setDeleteError] = useState('')

  const handleDeleteAccount = async () => {
    setDeleteError('')
    setIsDeleting(true)
    try {
      await deleteAccount()
      // 削除後もJWTは有効期限内のため、必ずサインアウトしてトークンを破棄する。
      // 遷移先を専用の完了画面にするのは、トップページに戻すだけでは削除できたか
      // 分からず、再ログインで新しいアカウントが作られることも伝わらないため。
      await logout('account-deleted')
    } catch (err) {
      setDeleteError(err instanceof Error ? err.message : 'アカウントの削除に失敗しました')
      setIsDeleting(false)
    }
  }

  return (
    <div className="bg-muted/30 min-h-screen">
      <Container className="py-6 md:py-10">
        <div className="max-w-[640px] mx-auto">
          <nav
            aria-label="パンくずリスト"
            className="text-[11.5px] md:text-[12.5px] mb-3 md:mb-[18px]"
          >
            <Link
              href={'/mypage' as Route}
              className="text-muted-foreground hover:text-primary transition-colors"
            >
              マイページ
            </Link>
            <span className="text-muted-foreground mx-1.5" aria-hidden="true">
              /
            </span>
            <span className="text-foreground">設定</span>
          </nav>
          <h1 className="text-xl md:text-[26px] font-bold text-foreground mb-1.5">設定</h1>
          <p className="text-[12.5px] md:text-[13.5px] text-muted-foreground mb-[18px] md:mb-6">
            アカウントに関する設定を確認できます
          </p>

          {/*
            表示名とアイコンはGoogleアカウントの情報をそのまま使うため、
            ここでは変更できない旨だけを伝える。編集できない項目を
            入力欄の形で並べると操作できるように見えてしまう。
          */}
          <div className="rounded-lg border border-border bg-card p-[18px] md:px-7 md:py-6">
            <h2 className="text-[13.5px] font-bold text-foreground md:text-sm">アカウント情報</h2>
            <dl className="mt-3.5 grid grid-cols-[88px_1fr] gap-x-4 gap-y-2.5 text-[12.5px] md:grid-cols-[104px_1fr] md:gap-x-5 md:text-[13px]">
              <dt className="text-muted-foreground">表示名</dt>
              <dd className="min-w-0 break-words text-foreground">{user?.name}</dd>
              <dt className="text-muted-foreground">メールアドレス</dt>
              <dd className="min-w-0 break-words text-foreground">{user?.email}</dd>
            </dl>
            <p className="mt-3.5 text-[11px] leading-[1.75] text-muted-foreground md:text-[11.5px]">
              表示名・アイコンはGoogleアカウントに準拠します。
            </p>
          </div>

          {/*
            アカウント削除は破壊的操作のため独立させる。
            警告は削除ボタンの配色1点に集約し、枠や本文を赤で塗り重ねない。
          */}
          <div className="mt-4 rounded-lg border border-border bg-card p-[18px] md:mt-6 md:px-7 md:py-6">
            <div className="flex flex-col md:flex-row md:items-start md:justify-between md:gap-8">
              <div className="min-w-0">
                <h2 className="text-[13.5px] font-bold text-foreground md:text-sm">
                  アカウントの削除
                </h2>
                <p className="mt-[7px] text-[12.5px] leading-[1.8] text-muted-foreground md:text-[13px]">
                  アカウントと記録をすべて削除します。
                </p>
              </div>
              <button
                type="button"
                onClick={() => setIsDeleteDialogOpen(true)}
                className="mt-3.5 h-[42px] w-full flex-shrink-0 whitespace-nowrap rounded-md border border-destructive/40 bg-card px-4 text-[13.5px] font-semibold text-destructive transition-colors hover:bg-destructive/5 focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 md:mt-0 md:h-[38px] md:w-auto md:text-[13px]"
              >
                アカウントを削除
              </button>
            </div>
          </div>
        </div>
      </Container>

      <DeleteAccountDialog
        isOpen={isDeleteDialogOpen}
        isDeleting={isDeleting}
        error={deleteError}
        onConfirm={handleDeleteAccount}
        onCancel={() => {
          setIsDeleteDialogOpen(false)
          setDeleteError('')
        }}
      />
    </div>
  )
}
