'use client'

import React, { useState, useRef } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/contexts/AuthContext'
import { Container } from '@/components/layout'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'

export const ProfileEditContent: React.FC = () => {
  const { user, token, updateUser } = useAuth()
  const router = useRouter()
  const fileInputRef = useRef<HTMLInputElement>(null)

  const [name, setName] = useState(user?.name || '')
  const [description, setDescription] = useState(user?.description || '')
  const [favoriteAnime, setFavoriteAnime] = useState(
    typeof window !== 'undefined' && user?.id
      ? localStorage.getItem(`user_${user.id}_favorite_anime`) || ''
      : ''
  )
  const [snsLink, setSnsLink] = useState(
    typeof window !== 'undefined' && user?.id
      ? localStorage.getItem(`user_${user.id}_sns_link`) || ''
      : ''
  )
  const [iconFile, setIconFile] = useState<File | null>(null)
  const [iconPreview, setIconPreview] = useState<string | null>(
    user?.icon || null
  )
  const [error, setError] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [hasChanges, setHasChanges] = useState(false)

  const API_BASE_URL =
    process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

  // 変更検知
  const checkForChanges = (
    newName: string,
    newDescription: string,
    newFavoriteAnime: string,
    newSnsLink: string,
    newIconFile: File | null
  ) => {
    const nameChanged = newName !== (user?.name || '')
    const descriptionChanged = newDescription !== (user?.description || '')
    const favoriteAnimeChanged = newFavoriteAnime !== ''
    const snsLinkChanged = newSnsLink !== ''
    // 画像アップロード機能は一時的に無効化
    // const iconChanged = newIconFile !== null
    setHasChanges(
      nameChanged ||
        descriptionChanged ||
        favoriteAnimeChanged ||
        snsLinkChanged
    )
  }

  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newName = e.target.value
    setName(newName)
    checkForChanges(newName, description, favoriteAnime, snsLink, iconFile)
  }

  const handleDescriptionChange = (
    e: React.ChangeEvent<HTMLTextAreaElement>
  ) => {
    const newDescription = e.target.value
    setDescription(newDescription)
    checkForChanges(name, newDescription, favoriteAnime, snsLink, iconFile)
  }

  const handleFavoriteAnimeChange = (
    e: React.ChangeEvent<HTMLTextAreaElement>
  ) => {
    const newFavoriteAnime = e.target.value
    setFavoriteAnime(newFavoriteAnime)
    checkForChanges(name, description, newFavoriteAnime, snsLink, iconFile)
  }

  const handleSnsLinkChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newSnsLink = e.target.value
    setSnsLink(newSnsLink)
    checkForChanges(name, description, favoriteAnime, newSnsLink, iconFile)
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
    if (!file.type.startsWith('image/')) {
      setError('画像ファイルを選択してください')
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

  const handleRemoveIcon = () => {
    setIconFile(null)
    setIconPreview(null)
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
    checkForChanges(name, description, favoriteAnime, snsLink, null)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    // バリデーション
    if (!name.trim()) {
      setError('ユーザー名を入力してください')
      return
    }

    if (name.length > 100) {
      setError('ユーザー名は100文字以内で入力してください')
      return
    }

    if (description.length > 500) {
      setError('自己紹介は500文字以内で入力してください')
      return
    }

    setIsLoading(true)

    try {
      // リクエストボディを構築（JSON形式）
      const requestBody: {
        name: string
        description: string
        icon?: string
      } = {
        name: name.trim(),
        description: description.trim(),
      }

      // 好きなアニメとSNSリンクをlocalStorageに保存（ユーザーIDごとに分離）
      // TODO: バックエンドAPIに対応するフィールドを追加後、requestBodyに含める
      if (user?.id) {
        if (favoriteAnime.trim()) {
          localStorage.setItem(`user_${user.id}_favorite_anime`, favoriteAnime.trim())
        } else {
          localStorage.removeItem(`user_${user.id}_favorite_anime`)
        }

        if (snsLink.trim()) {
          localStorage.setItem(`user_${user.id}_sns_link`, snsLink.trim())
        } else {
          localStorage.removeItem(`user_${user.id}_sns_link`)
        }
      }

      // 画像がある場合はBase64エンコードして送信
      // 注: 現在のバックエンドはiconフィールドにURLを期待しているため、
      // 画像アップロード機能は一旦無効化
      // TODO: バックエンドに画像アップロードAPIを実装後に有効化
      // if (iconFile) {
      //   const base64 = await convertFileToBase64(iconFile);
      //   requestBody.icon = base64;
      // }

      const response = await fetch(`${API_BASE_URL}/api/users/me`, {
        method: 'PUT',
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestBody),
      })

      if (!response.ok) {
        const error = await response.json()
        throw new Error(error.message || 'プロフィールの更新に失敗しました')
      }

      const data = await response.json()
      const updatedUser = data.user || data

      // コンテキストのユーザー情報を更新
      updateUser(updatedUser)

      // マイページに戻る
      router.push('/mypage')
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'プロフィールの更新に失敗しました'
      )
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

            {/* アイコン画像 - 一時的に無効化 */}
            <div className="opacity-50">
              <label className="block text-sm font-medium text-gray-700 mb-3">
                プロフィール画像
                <span className="ml-2 text-xs text-gray-500">（準備中）</span>
              </label>
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

                {/* ボタン */}
                <div className="flex-1">
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept="image/*"
                    onChange={handleIconChange}
                    className="hidden"
                    disabled
                  />
                  <div className="flex space-x-3">
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={handleIconClick}
                      disabled
                    >
                      画像を選択
                    </Button>
                    {iconPreview && (
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={handleRemoveIcon}
                        disabled
                      >
                        削除
                      </Button>
                    )}
                  </div>
                  <p className="text-xs text-gray-500 mt-2">
                    画像アップロード機能は準備中です
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
              helperText="1〜100文字で入力してください"
            />

            {/* プロフィール */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                プロフィール
              </label>
              <textarea
                value={description}
                onChange={handleDescriptionChange}
                placeholder="アニメが大好きです！"
                rows={4}
                maxLength={500}
                className="w-full px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
              />
              <p className="text-xs text-gray-500 mt-2">
                {description.length}/500文字
              </p>
            </div>

            {/* 好きなアニメ */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                好きなアニメ
              </label>
              <textarea
                value={favoriteAnime}
                onChange={handleFavoriteAnimeChange}
                placeholder="アクション、ファンタジー系のアニメが好きです"
                rows={3}
                maxLength={300}
                className="w-full px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
              />
              <p className="text-xs text-gray-500 mt-2">
                {favoriteAnime.length}/300文字
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
                size="lg"
                isLoading={isLoading}
                disabled={!hasChanges}
                className="flex-1"
              >
                保存する
              </Button>
              <Button
                type="button"
                variant="outline"
                size="lg"
                onClick={handleCancel}
                disabled={isLoading}
                className="flex-1"
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
