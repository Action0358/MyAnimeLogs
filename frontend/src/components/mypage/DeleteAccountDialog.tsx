'use client'

import React, { useEffect, useRef } from 'react'
import { API_BASE_URL } from '@/lib/api/client'

interface DeleteAccountDialogProps {
  isOpen: boolean
  isDeleting: boolean
  error?: string
  onConfirm: () => void
  onCancel: () => void
}

const DELETED_ITEMS = [
  'プロフィール（表示名・自己紹介・好きな作品・アイコン画像）',
  'お気に入りに登録した作品',
  '視聴予定・視聴中・視聴完了の記録',
]

/**
 * アカウント削除の確認ダイアログ。
 *
 * 復旧不可の操作のため、削除される対象を具体的に列挙したうえで確認させる。
 * テキスト入力による確認は採用しない（実際に削除したいユーザーの妨げになるため）。
 * 代わりに、キャンセルを既定の操作として扱い、削除ボタンは destructive 配色にする。
 */
export const DeleteAccountDialog: React.FC<DeleteAccountDialogProps> = ({
  isOpen,
  isDeleting,
  error,
  onConfirm,
  onCancel,
}) => {
  const cancelButtonRef = useRef<HTMLButtonElement>(null)

  // 開いた直後のフォーカスはキャンセル側に置き、Enter の連打で
  // 誤って削除が実行されないようにする
  useEffect(() => {
    if (isOpen) cancelButtonRef.current?.focus()
  }, [isOpen])

  useEffect(() => {
    if (!isOpen) return
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && !isDeleting) onCancel()
    }
    document.addEventListener('keydown', onKeyDown)
    return () => document.removeEventListener('keydown', onKeyDown)
  }, [isOpen, isDeleting, onCancel])

  if (!isOpen) return null

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="delete-account-title"
    >
      <div className="w-full max-w-md rounded-lg border border-border bg-card p-6 shadow-lg">
        <h2 id="delete-account-title" className="text-lg font-bold text-foreground">
          アカウントを削除しますか？
        </h2>

        <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
          以下のデータがすべて削除されます。
        </p>
        <ul className="mt-2 list-disc space-y-1 pl-5 text-sm text-muted-foreground marker:text-destructive">
          {DELETED_ITEMS.map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ul>

        <p className="mt-4 rounded-md border border-destructive/20 bg-destructive/5 p-3 text-sm font-semibold text-destructive">
          削除したデータは復元できません。
        </p>

        <p className="mt-3 text-xs leading-relaxed text-muted-foreground">
          削除後に同じGoogleアカウントでログインした場合は、新しいアカウントとして登録されます。以前の記録は引き継がれません。
        </p>

        <p className="mt-3 text-xs leading-relaxed text-muted-foreground">
          記録を手元に残したい場合は、削除前に
          <a
            href={`${API_BASE_URL}/api/users/me/export?format=json`}
            className="text-primary hover:underline"
            target="_blank"
            rel="noopener noreferrer"
          >
            データをエクスポート
          </a>
          してください。
          <br />
          なお、Googleアカウント側に残る「MyAnimeLogsへのアクセス許可」は、
          Googleのアカウント設定から別途解除してください。
        </p>

        {error && (
          <p className="mt-4 rounded-md border border-destructive/20 bg-destructive/5 p-3 text-sm text-destructive">
            {error}
          </p>
        )}

        <div className="mt-6 flex justify-end gap-3">
          <button
            ref={cancelButtonRef}
            type="button"
            onClick={onCancel}
            disabled={isDeleting}
            className="rounded-md border border-border px-4 py-2 text-sm font-semibold text-foreground transition-colors hover:bg-muted disabled:cursor-not-allowed disabled:opacity-60"
          >
            キャンセル
          </button>
          <button
            type="button"
            onClick={onConfirm}
            disabled={isDeleting}
            className="rounded-md bg-destructive px-4 py-2 text-sm font-semibold text-destructive-foreground transition-colors hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {isDeleting ? '削除しています...' : 'アカウントを削除'}
          </button>
        </div>
      </div>
    </div>
  )
}
