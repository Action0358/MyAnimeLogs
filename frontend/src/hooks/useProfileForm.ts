'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/contexts/AuthContext'
import { put, ApiClientError } from '@/lib/api/client'
import {
  getProfileFieldValues,
  MAX_NAME_LENGTH,
  MAX_DESCRIPTION_LENGTH,
} from '@/lib/profile'

export interface UseProfileFormResult {
  name: string
  description: string
  favoriteAnime: string
  error: string
  isLoading: boolean
  hasChanges: boolean
  handleNameChange: (e: React.ChangeEvent<HTMLInputElement>) => void
  handleDescriptionChange: (e: React.ChangeEvent<HTMLTextAreaElement>) => void
  handleFavoriteAnimeChange: (e: React.ChangeEvent<HTMLTextAreaElement>) => void
  handleSubmit: (e: React.FormEvent) => Promise<void>
  handleCancel: () => void
}

// プロフィール編集フォームの状態・バリデーション・変更検知・保存処理を集約したフック。
// アイコンは #779 でGoogleアカウントのプロフィール画像に一本化し、編集対象から外した。
export const useProfileForm = (): UseProfileFormResult => {
  const { user, updateUser } = useAuth()
  const router = useRouter()

  // 初期値を保持（変更検知用）
  const [initialValues] = useState(() => getProfileFieldValues(user))

  const [name, setName] = useState(initialValues.name)
  const [description, setDescription] = useState(initialValues.description)
  const [favoriteAnime, setFavoriteAnime] = useState(initialValues.favoriteAnime)

  const [error, setError] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [hasChanges, setHasChanges] = useState(false)

  // 変更検知（初期値と現在値を比較）
  const checkForChanges = (
    newName: string,
    newDescription: string,
    newFavoriteAnime: string
  ) => {
    const nameChanged = newName !== initialValues.name
    const descriptionChanged = newDescription !== initialValues.description
    const favoriteAnimeChanged =
      newFavoriteAnime !== initialValues.favoriteAnime

    setHasChanges(nameChanged || descriptionChanged || favoriteAnimeChanged)
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
          err instanceof Error ? err.message : 'プロフィールの更新に失敗しました'
        )
      }
    } finally {
      setIsLoading(false)
    }
  }

  const handleCancel = () => {
    if (hasChanges) {
      if (window.confirm('変更内容が保存されていません。破棄してもよろしいですか？')) {
        router.push('/mypage')
      }
    } else {
      router.push('/mypage')
    }
  }

  return {
    name,
    description,
    favoriteAnime,
    error,
    isLoading,
    hasChanges,
    handleNameChange,
    handleDescriptionChange,
    handleFavoriteAnimeChange,
    handleSubmit,
    handleCancel,
  }
}
