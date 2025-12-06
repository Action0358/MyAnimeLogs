'use client'

import React, { useState, useRef } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/contexts/AuthContext'
import { Container } from '@/components/layout'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { put, uploadFile, ApiClientError } from '@/lib/api/client'
import { buildIconUrl } from '@/lib/utils'

// Validation constants
const MAX_NAME_LENGTH = 50
const MAX_DESCRIPTION_LENGTH = 500
const MAX_FAVORITE_ANIME_LENGTH = 200

// Shared styles
const TEXTAREA_CLASS_NAME =
  'block w-full px-2 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none'
const HELPER_TEXT_CLASS_NAME = 'text-sm text-gray-500 mt-2 px-2'

export const ProfileEditContent: React.FC = () => {
  const { user, updateUser } = useAuth()
  const router = useRouter()
  const fileInputRef = useRef<HTMLInputElement>(null)

  // 初期値を保持（変更検知用）
  const [initialValues] = useState(() => ({
    name: user?.name || '',
    description: user?.description || '',
    favoriteAnime: user?.favorite_anime || '',
    snsLink: user?.sns_link || '',
  }))

  const [name, setName] = useState(initialValues.name)
  const [description, setDescription] = useState(initialValues.description)
  const [favoriteAnime, setFavoriteAnime] = useState(
    initialValues.favoriteAnime
  )
  const [snsLink, setSnsLink] = useState(initialValues.snsLink)

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
    newSnsLink: string,
    newIconFile: File | null = iconFile
  ) => {
    const nameChanged = newName !== initialValues.name
    const descriptionChanged = newDescription !== initialValues.description
    const favoriteAnimeChanged =
      newFavoriteAnime !== initialValues.favoriteAnime
    const snsLinkChanged = newSnsLink !== initialValues.snsLink
    const iconChanged = newIconFile !== null

    setHasChanges(
      nameChanged ||
        descriptionChanged ||
        favoriteAnimeChanged ||
        snsLinkChanged ||
        iconChanged
    )
  }

  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newName = e.target.value
    setName(newName)
    checkForChanges(newName, description, favoriteAnime, snsLink)
  }

  const handleDescriptionChange = (
    e: React.ChangeEvent<HTMLTextAreaElement>
  ) => {
    const newDescription = e.target.value
    setDescription(newDescription)
    checkForChanges(name, newDescription, favoriteAnime, snsLink)
  }

  const handleFavoriteAnimeChange = (
    e: React.ChangeEvent<HTMLTextAreaElement>
  ) => {
    const newFavoriteAnime = e.target.value
    setFavoriteAnime(newFavoriteAnime)
    checkForChanges(name, description, newFavoriteAnime, snsLink)
  }

  const handleSnsLinkChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newSnsLink = e.target.value
    setSnsLink(newSnsLink)
    checkForChanges(name, description, favoriteAnime, newSnsLink)
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

    checkForChanges(name, description, favoriteAnime, snsLink, file)
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
        sns_link: snsLink.trim() || null,
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

  return (
    <Container className="py-12">
      <div className="max-w-2xl mx-auto">
        {/* ヘッダー */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            プロフィール編集
          </h1>
          <p className="text-gray-600">
            あなたのプロフィール情報を編集できます
          </p>
        </div>

        {/* 編集フォーム */}
        <div className="bg-white rounded-lg border border-gray-200 p-8 shadow-sm">
          <form onSubmit={handleSubmit} className="space-y-6">
            {error && (
              <div className="p-4 bg-red-50 border border-red-200 rounded-md">
                <p className="text-sm text-red-600">{error}</p>
              </div>
            )}

            {/* アイコン画像 */}
            <div>
              <div className="flex items-center space-x-6">
                {/* プレビュー */}
                <div className="w-24 h-24 bg-gray-100 rounded-full flex items-center justify-center overflow-hidden">
                  {iconPreview ? (
                    <img
                      src={iconPreview}
                      alt="プロフィール画像"
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    <span className="text-gray-400 text-4xl">👤</span>
                  )}
                </div>

                {/* ボタンと説明 */}
                <div className="flex-1">
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept="image/jpeg,image/png,image/gif,image/webp"
                    onChange={handleIconChange}
                    className="hidden"
                  />
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={handleIconClick}
                    disabled={isLoading || isUploadingIcon}
                  >
                    {isUploadingIcon ? 'アップロード中...' : 'アイコン変更'}
                  </Button>
                  <p className="text-xs text-gray-500 mt-2">
                    JPEG、PNG、GIF、WebP形式（最大5MB）
                  </p>
                </div>
              </div>
            </div>

            {/* ユーザー名 */}
            <Input
              label="ユーザー名"
              type="text"
              value={name}
              onChange={handleNameChange}
              placeholder="マイアニメログ"
              required
              helperText={`1〜${MAX_NAME_LENGTH}文字で入力してください`}
            />

            {/* プロフィール */}
            <div>
              <label className="block text-base font-bold text-gray-700 mb-2 px-2">
                プロフィール
              </label>
              <textarea
                value={description}
                onChange={handleDescriptionChange}
                placeholder="アニメが大好きです！"
                rows={4}
                maxLength={MAX_DESCRIPTION_LENGTH}
                className={TEXTAREA_CLASS_NAME}
              />
              <p className={HELPER_TEXT_CLASS_NAME}>
                {description.length}/{MAX_DESCRIPTION_LENGTH}文字
              </p>
            </div>

            {/* 好きなアニメ */}
            <div>
              <label className="block text-base font-bold text-gray-700 mb-2 px-2">
                好きなアニメ
              </label>
              <textarea
                value={favoriteAnime}
                onChange={handleFavoriteAnimeChange}
                placeholder="アクション、ファンタジー系のアニメが好きです"
                rows={3}
                maxLength={MAX_FAVORITE_ANIME_LENGTH}
                className={TEXTAREA_CLASS_NAME}
              />
              <p className={HELPER_TEXT_CLASS_NAME}>
                {favoriteAnime.length}/{MAX_FAVORITE_ANIME_LENGTH}文字
              </p>
            </div>

            {/* SNSリンク */}
            <Input
              label="SNSリンク"
              type="url"
              value={snsLink}
              onChange={handleSnsLinkChange}
              placeholder="https://twitter.com/username"
            />

            {/* ボタン */}
            <div className="flex space-x-4 pt-4">
              <Button
                type="submit"
                variant="primary"
                size="md"
                isLoading={isLoading || isUploadingIcon}
                disabled={!hasChanges || isLoading || isUploadingIcon}
                className="flex-1 text-sm"
              >
                {isLoading || isUploadingIcon ? '保存中...' : '保存する'}
              </Button>
              <Button
                type="button"
                variant="outline"
                size="md"
                onClick={handleCancel}
                disabled={isLoading || isUploadingIcon}
                className="flex-1 text-sm"
              >
                キャンセル
              </Button>
            </div>
          </form>
        </div>
      </div>
    </Container>
  )
}
