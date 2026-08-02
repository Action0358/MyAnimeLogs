'use client'

import React, { useRef } from 'react'
import { Container } from '@/components/layout'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import {
  MAX_NAME_LENGTH,
  MAX_DESCRIPTION_LENGTH,
  MAX_FAVORITE_ANIME_LENGTH,
} from '@/lib/profile'
import { useProfileForm } from '@/hooks/useProfileForm'

// Shared styles
const TEXTAREA_CLASS_NAME =
  'block w-full px-2 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none'
const HELPER_TEXT_CLASS_NAME = 'text-sm text-gray-500 mt-2 px-2'

export const ProfileEditContent: React.FC = () => {
  const fileInputRef = useRef<HTMLInputElement>(null)

  const {
    name,
    description,
    favoriteAnime,
    snsLink,
    iconPreview,
    error,
    isLoading,
    isUploadingIcon,
    hasChanges,
    handleNameChange,
    handleDescriptionChange,
    handleFavoriteAnimeChange,
    handleSnsLinkChange,
    handleIconChange,
    handleSubmit,
    handleCancel,
  } = useProfileForm()

  const handleIconClick = () => {
    fileInputRef.current?.click()
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
