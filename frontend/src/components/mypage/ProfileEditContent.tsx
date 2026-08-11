'use client'

import React, { useState, useRef } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import type { Route } from 'next'
import { useAuth } from '@/contexts/AuthContext'
import { Container } from '@/components/layout'
import { put, uploadFile, ApiClientError } from '@/lib/api/client'
import { buildIconUrl } from '@/lib/utils'

// Validation constants
const MAX_NAME_LENGTH = 50
const MAX_DESCRIPTION_LENGTH = 500
const MAX_FAVORITE_ANIME_LENGTH = 200

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
  const { user, updateUser } = useAuth()
  const router = useRouter()
  const fileInputRef = useRef<HTMLInputElement>(null)

  // 初期値を保持（変更検知用）
  const [initialValues] = useState(() => ({
    name: user?.name || '',
    description: user?.description || '',
    favoriteAnime: user?.favorite_anime || '',
  }))

  const [name, setName] = useState(initialValues.name)
  const [description, setDescription] = useState(initialValues.description)
  const [favoriteAnime, setFavoriteAnime] = useState(
    initialValues.favoriteAnime
  )

  const [iconFile, setIconFile] = useState<File | null>(null)
  const [iconPreview, setIconPreview] = useState<string | null>(
    buildIconUrl(user?.icon)
  )
  const [error, setError] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [isUploadingIcon, setIsUploadingIcon] = useState(false)
  const [hasChanges, setHasChanges] = useState(false)

  // 画像アップロードAPI
  const uploadIcon = async (file: File): Promise<void> => {
    setIsUploadingIcon(true)
    try {
      const formData = new FormData()
      formData.append('icon', file)

      const data = await uploadFile<{ user: any }>(
        '/api/users/me/icon',
        formData
      )
      updateUser(data.user)
    } catch (err) {
      // エラーの型情報を保持したまま再スロー
      throw err
    } finally {
      setIsUploadingIcon(false)
    }
  }

  // 変更検知（初期値と現在値を比較）
  const checkForChanges = (
    newName: string,
    newDescription: string,
    newFavoriteAnime: string,
    newIconFile: File | null = iconFile
  ) => {
    const nameChanged = newName !== initialValues.name
    const descriptionChanged = newDescription !== initialValues.description
    const favoriteAnimeChanged =
      newFavoriteAnime !== initialValues.favoriteAnime
    const iconChanged = newIconFile !== null

    setHasChanges(
      nameChanged || descriptionChanged || favoriteAnimeChanged || iconChanged
    )
  }

  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newName = e.target.value
    setName(newName)
    checkForChanges(newName, description, favoriteAnime)
  }

  const handleDescriptionChange = (
    e: React.ChangeEvent<HTMLTextAreaElement>
  ) => {
    const newDescription = e.target.value
    setDescription(newDescription)
    checkForChanges(name, newDescription, favoriteAnime)
  }

  const handleFavoriteAnimeChange = (
    e: React.ChangeEvent<HTMLTextAreaElement>
  ) => {
    const newFavoriteAnime = e.target.value
    setFavoriteAnime(newFavoriteAnime)
    checkForChanges(name, description, newFavoriteAnime)
  }

  const handleIconChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    // ファイルサイズチェック（5MB以下）
    if (file.size > 5 * 1024 * 1024) {
      setError('画像サイズは5MB以下にしてください')
      return
    }

    // ファイル形式チェック
    const allowedTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']
    if (!allowedTypes.includes(file.type)) {
      setError('JPEG、PNG、GIF、WebP形式の画像を選択してください')
      return
    }

    setIconFile(file)
    setError('')

    // プレビュー生成
    const reader = new FileReader()
    reader.onloadend = () => {
      setIconPreview(reader.result as string)
    }
    reader.readAsDataURL(file)

    checkForChanges(name, description, favoriteAnime, file)
  }

  const handleIconClick = () => {
    fileInputRef.current?.click()
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    // バリデーション
    if (!name.trim()) {
      setError('ユーザー名を入力してください')
      return
    }

    if (name.length > MAX_NAME_LENGTH) {
      setError(`ユーザー名は${MAX_NAME_LENGTH}文字以内で入力してください`)
      return
    }

    if (description.length > MAX_DESCRIPTION_LENGTH) {
      setError(`自己紹介は${MAX_DESCRIPTION_LENGTH}文字以内で入力してください`)
      return
    }

    setIsLoading(true)

    try {
      // 1. 画像がある場合は先にアップロード
      if (iconFile) {
        await uploadIcon(iconFile)
      }

      // 2. プロフィール情報を更新
      const requestBody = {
        name: name.trim(),
        description: description.trim(),
        favorite_anime: favoriteAnime.trim() || null,
      }

      const data = await put<{ user: any }>('/api/users/me', requestBody)
      const updatedUser = data.user || data

      // コンテキストのユーザー情報を更新
      updateUser(updatedUser)

      // マイページに戻る
      router.push('/mypage')
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.message)
      } else {
        setError(
          err instanceof Error
            ? err.message
            : 'プロフィールの更新に失敗しました'
        )
      }
    } finally {
      setIsLoading(false)
    }
  }

  const handleCancel = () => {
    if (hasChanges) {
      if (confirm('変更内容が保存されていません。破棄してもよろしいですか？')) {
        router.push('/mypage')
      }
    } else {
      router.push('/mypage')
    }
  }

  const isBusy = isLoading || isUploadingIcon

  return (
    <div className="bg-muted/30 min-h-screen">
      <Container className="py-6 md:py-10">
        <div className="max-w-[640px] mx-auto">
          {/* ヘッダー */}
          <nav aria-label="パンくずリスト" className="text-[11.5px] md:text-[12.5px] mb-3 md:mb-[18px]">
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
            <div
              className={`${SECTION_CLASS_NAME} border-b border-border flex flex-col gap-[18px] md:gap-[22px]`}
            >
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
        </div>
      </Container>
    </div>
  )
}
