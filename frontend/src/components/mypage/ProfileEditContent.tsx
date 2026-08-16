'use client'

import React, { useRef, useState } from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import { Container } from '@/components/layout'
import { useAuth } from '@/contexts/AuthContext'
import { deleteAccount } from '@/lib/api/account'
import { DeleteAccountDialog } from './DeleteAccountDialog'
import {
  MAX_NAME_LENGTH,
  MAX_DESCRIPTION_LENGTH,
  MAX_FAVORITE_ANIME_LENGTH,
} from '@/lib/profile'
import { useProfileForm } from '@/hooks/useProfileForm'

// Shared styles
const FIELD_BASE_CLASS_NAME =
  'w-full border border-border rounded-md bg-card text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent'
const INPUT_CLASS_NAME = `${FIELD_BASE_CLASS_NAME} h-[42px] px-3.5 text-sm`
const TEXTAREA_CLASS_NAME = `${FIELD_BASE_CLASS_NAME} px-3.5 py-2.5 text-[13.5px] md:text-sm leading-[1.7] resize-y`
const HELPER_TEXT_CLASS_NAME = 'text-[11px] md:text-[11.5px] text-muted-foreground mt-1.5'
// 区切られた各グループの余白
const SECTION_CLASS_NAME = 'px-[18px] py-[18px] md:px-7 md:py-[26px]'

/**
 * ラベル行。文字数カウンターを入力欄の下ではなくラベルの右端に置き、
 * 「何の項目が何文字か」を1行で読めるようにする。
 */
const FieldLabel: React.FC<{
  htmlFor: string
  children: React.ReactNode
  required?: boolean
  current?: number
  max?: number
}> = ({ htmlFor, children, required = false, current, max }) => (
  <div className="flex items-baseline justify-between mb-[7px] md:mb-2 gap-3">
    <label
      htmlFor={htmlFor}
      className="text-[12.5px] md:text-[13.5px] font-semibold text-foreground"
    >
      {children}
      {required && (
        <span className="text-destructive font-medium ml-1" aria-hidden="true">
          *
        </span>
      )}
    </label>
    {max !== undefined && (
      <span className="text-[11px] md:text-[11.5px] text-muted-foreground tabular-nums flex-shrink-0">
        {current} / {max}
      </span>
    )}
  </div>
)

export const ProfileEditContent: React.FC = () => {
  const fileInputRef = useRef<HTMLInputElement>(null)

  const {
    name,
    description,
    favoriteAnime,
    iconPreview,
    error,
    isLoading,
    isUploadingIcon,
    hasChanges,
    handleNameChange,
    handleDescriptionChange,
    handleFavoriteAnimeChange,
    handleIconChange,
    handleSubmit,
    handleCancel,
  } = useProfileForm()

  const { logout } = useAuth()
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

  const handleIconClick = () => {
    fileInputRef.current?.click()
  }

  const isBusy = isLoading || isUploadingIcon

  return (
    <div className="bg-muted/30 min-h-screen">
      <Container className="py-6 md:py-10">
        <div className="max-w-[640px] mx-auto">
          {/* ヘッダー */}
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
            <span className="text-foreground">プロフィール編集</span>
          </nav>
          <h1 className="text-xl md:text-[26px] font-bold text-foreground mb-1.5">
            プロフィール編集
          </h1>
          <p className="text-[12.5px] md:text-[13.5px] text-muted-foreground mb-[18px] md:mb-6">
            公開されるプロフィール情報を編集できます
          </p>

          {/* 編集フォーム。項目を意味のまとまりごとに区切り線でグループ化する */}
          <form onSubmit={handleSubmit} className="bg-card border border-border rounded-lg">
            {error && (
              <div className={`${SECTION_CLASS_NAME} border-b border-border`}>
                <div className="p-3 bg-destructive/10 border border-destructive/30 rounded-md">
                  <p className="text-sm text-destructive">{error}</p>
                </div>
              </div>
            )}

            {/* アイコン */}
            <div
              className={`${SECTION_CLASS_NAME} border-b border-border flex items-center gap-3.5 md:gap-5`}
            >
              <div className="w-[60px] h-[60px] md:w-[72px] md:h-[72px] bg-muted rounded-full flex items-center justify-center overflow-hidden flex-shrink-0">
                {iconPreview ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img
                    src={iconPreview}
                    alt="プロフィール画像"
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <span className="text-muted-foreground text-3xl md:text-4xl">👤</span>
                )}
              </div>

              <div className="flex-1 min-w-0">
                <div className="text-[12.5px] md:text-[13.5px] font-semibold text-foreground mb-[7px] md:mb-2">
                  アイコン
                </div>
                <input
                  ref={fileInputRef}
                  type="file"
                  accept="image/jpeg,image/png,image/gif,image/webp"
                  onChange={handleIconChange}
                  className="hidden"
                />
                <div className="flex flex-col md:flex-row md:items-center gap-1.5 md:gap-3">
                  <button
                    type="button"
                    onClick={handleIconClick}
                    disabled={isBusy}
                    className="h-8 md:h-[34px] px-3 md:px-3.5 self-start border border-border rounded-md bg-card text-xs md:text-[12.5px] font-semibold text-foreground whitespace-nowrap hover:bg-muted transition-colors disabled:opacity-50 disabled:cursor-not-allowed focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                  >
                    {isUploadingIcon ? 'アップロード中...' : '画像を選択'}
                  </button>
                  <span className="text-[10.5px] md:text-[11.5px] text-muted-foreground">
                    JPEG / PNG / GIF / WebP・最大5MB
                  </span>
                </div>
              </div>
            </div>

            {/* 基本情報 */}
            <div className={`${SECTION_CLASS_NAME} border-b border-border`}>
              <FieldLabel
                htmlFor="profile-name"
                required
                current={name.length}
                max={MAX_NAME_LENGTH}
              >
                ユーザー名
              </FieldLabel>
              <input
                id="profile-name"
                type="text"
                value={name}
                onChange={handleNameChange}
                placeholder="マイアニメログ"
                required
                maxLength={MAX_NAME_LENGTH}
                className={INPUT_CLASS_NAME}
              />
              <p className={HELPER_TEXT_CLASS_NAME}>
                1〜{MAX_NAME_LENGTH}文字。マイページや一覧に表示されます
              </p>
            </div>

            {/* 紹介 */}
            <div className={`${SECTION_CLASS_NAME} flex flex-col gap-[18px] md:gap-[22px]`}>
              <div>
                <FieldLabel
                  htmlFor="profile-description"
                  current={description.length}
                  max={MAX_DESCRIPTION_LENGTH}
                >
                  プロフィール
                </FieldLabel>
                <textarea
                  id="profile-description"
                  value={description}
                  onChange={handleDescriptionChange}
                  placeholder="アニメが大好きです！"
                  rows={4}
                  maxLength={MAX_DESCRIPTION_LENGTH}
                  className={TEXTAREA_CLASS_NAME}
                />
              </div>
              <div>
                <FieldLabel
                  htmlFor="profile-favorite-anime"
                  current={favoriteAnime.length}
                  max={MAX_FAVORITE_ANIME_LENGTH}
                >
                  好きなアニメ
                </FieldLabel>
                <textarea
                  id="profile-favorite-anime"
                  value={favoriteAnime}
                  onChange={handleFavoriteAnimeChange}
                  placeholder="アクション、ファンタジー系のアニメが好きです"
                  rows={3}
                  maxLength={MAX_FAVORITE_ANIME_LENGTH}
                  className={TEXTAREA_CLASS_NAME}
                />
              </div>
            </div>

            {/* 操作列。主操作を「変更を保存」1つに絞る */}
            <div className="px-[18px] py-4 md:px-7 md:py-[18px] border-t border-border bg-muted/30 rounded-b-lg flex flex-col-reverse md:flex-row md:items-center md:justify-end gap-2.5 md:gap-3.5">
              <button
                type="button"
                onClick={handleCancel}
                disabled={isBusy}
                className="h-10 px-2 rounded-md text-[13px] md:text-[13.5px] font-semibold text-muted-foreground hover:text-foreground transition-colors disabled:opacity-50 disabled:cursor-not-allowed focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
              >
                キャンセル
              </button>
              <button
                type="submit"
                disabled={!hasChanges || isBusy}
                className="h-11 md:h-10 px-[22px] rounded-md text-sm md:text-[13.5px] font-semibold transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 bg-primary text-primary-foreground hover:bg-primary/90 disabled:bg-muted disabled:text-muted-foreground disabled:cursor-not-allowed disabled:hover:bg-muted"
              >
                {isBusy ? '保存中...' : '変更を保存'}
              </button>
            </div>
          </form>

          {/*
            アカウント削除は破壊的操作のため、保存フォームの外に独立させる。
            ただし幅はフォームカードに揃え、同じ設定ページの続きとして読ませる。
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
