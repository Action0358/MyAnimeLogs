'use client'

import { useState } from 'react'
import { uploadFile } from '@/lib/api/client'
import { buildIconUrl } from '@/lib/utils'

// アイコン画像の制約
const MAX_ICON_SIZE = 5 * 1024 * 1024 // 5MB
const ALLOWED_ICON_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']

const ICON_SIZE_ERROR = '画像サイズは5MB以下にしてください'
const ICON_TYPE_ERROR = 'JPEG、PNG、GIF、WebP形式の画像を選択してください'

interface UseIconUploadParams {
  // 現在のユーザーアイコンパス（初期プレビュー生成用）
  initialIcon?: string
  // アップロード成功時にコンテキストのユーザー情報を更新するコールバック
  updateUser: (user: any) => void
}

export interface UseIconUploadResult {
  iconFile: File | null
  iconPreview: string | null
  isUploadingIcon: boolean
  // ファイルを検証してプレビューを生成する。検証に失敗した場合はエラーメッセージを返す
  selectIcon: (file: File) => { error?: string }
  // 選択済みのアイコンをアップロードする（未選択なら何もしない）
  uploadIcon: () => Promise<void>
}

// アイコン画像の選択・プレビュー・アップロードを扱うフック。
// ProfileEditContent から画像まわりのロジックを分離したもの。挙動は不変。
export const useIconUpload = ({
  initialIcon,
  updateUser,
}: UseIconUploadParams): UseIconUploadResult => {
  const [iconFile, setIconFile] = useState<File | null>(null)
  const [iconPreview, setIconPreview] = useState<string | null>(
    buildIconUrl(initialIcon)
  )
  const [isUploadingIcon, setIsUploadingIcon] = useState(false)

  const selectIcon = (file: File): { error?: string } => {
    // ファイルサイズチェック（5MB以下）
    if (file.size > MAX_ICON_SIZE) {
      return { error: ICON_SIZE_ERROR }
    }

    // ファイル形式チェック
    if (!ALLOWED_ICON_TYPES.includes(file.type)) {
      return { error: ICON_TYPE_ERROR }
    }

    setIconFile(file)

    // プレビュー生成
    const reader = new FileReader()
    reader.onloadend = () => {
      setIconPreview(reader.result as string)
    }
    reader.readAsDataURL(file)

    return {}
  }

  const uploadIcon = async (): Promise<void> => {
    if (!iconFile) return

    setIsUploadingIcon(true)
    try {
      const formData = new FormData()
      formData.append('icon', iconFile)

      const data = await uploadFile<{ user: any }>(
        '/api/users/me/icon',
        formData
      )
      updateUser(data.user)
    } finally {
      setIsUploadingIcon(false)
    }
  }

  return {
    iconFile,
    iconPreview,
    isUploadingIcon,
    selectIcon,
    uploadIcon,
  }
}
