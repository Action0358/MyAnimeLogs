'use client'

import React, { useEffect, useId, useRef, useState } from 'react'
import { AlertTriangle, Info } from 'lucide-react'

interface DeleteAccountDialogProps {
  isOpen: boolean
  isDeleting: boolean
  error?: string
  onConfirm: () => void
  onCancel: () => void
}

const DELETED_ITEMS = [
  'プロフィール',
  'お気に入りに登録した作品',
  '視聴予定・視聴中・視聴済の記録',
  '検索の記録',
]

const NOTES = [
  '同じGoogleアカウントで再ログインしても、以前の記録は引き継がれません。',
  'Googleアカウント側に残る「MyAnimeLogsへのアクセス許可」は、Googleのアカウント設定から別途解除してください。',
]

/**
 * アカウント削除の確認ダイアログ。
 *
 * 復旧不可の操作のため、削除される対象を具体的に列挙したうえで確認させる。
 * テキスト入力による確認は採用しない（実際に削除したいユーザーの妨げになるため）。
 * 代わりに確認チェックボックスで一段挟み、未チェックの間は削除を実行できなくする。
 *
 * レイアウトはヘッダー / スクロールする本文 / 固定フッターの3層。本文が長くても
 * 操作ボタンがビューポート外へ押し出されないようにするため。
 * 危険色は削除ボタン1点に集約し、枠や本文を赤で塗り重ねない。
 */
export const DeleteAccountDialog: React.FC<DeleteAccountDialogProps> = ({
  isOpen,
  isDeleting,
  error,
  onConfirm,
  onCancel,
}) => {
  const cancelButtonRef = useRef<HTMLButtonElement>(null)
  const [agreed, setAgreed] = useState(false)
  const agreeId = useId()

  // 開いた直後のフォーカスはキャンセル側に置き、Enter の連打で
  // 誤って削除が実行されないようにする
  useEffect(() => {
    if (isOpen) cancelButtonRef.current?.focus()
  }, [isOpen])

  // 開き直したときに前回の同意を引き継がない
  useEffect(() => {
    if (!isOpen) setAgreed(false)
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
      className="fixed inset-0 z-50 flex items-end justify-center bg-black/55 md:items-center md:p-10"
      role="dialog"
      aria-modal="true"
      aria-labelledby="delete-account-title"
    >
      <div className="flex max-h-[92%] w-full flex-col overflow-hidden rounded-t-2xl bg-card shadow-xl md:max-h-full md:w-[520px] md:rounded-xl">
        {/* ヘッダー */}
        <div className="flex-shrink-0 px-5 pb-3.5 pt-[18px] md:px-7 md:pb-[18px] md:pt-[26px]">
          <div
            className="mx-auto mb-4 h-1 w-9 rounded-full bg-border md:hidden"
            aria-hidden="true"
          />
          <div className="flex items-center gap-3">
            <div
              className="flex h-[34px] w-[34px] flex-shrink-0 items-center justify-center rounded-full bg-destructive/10 text-destructive md:h-[38px] md:w-[38px]"
              aria-hidden="true"
            >
              <AlertTriangle className="h-[19px] w-[19px]" />
            </div>
            <h2
              id="delete-account-title"
              className="text-[16.5px] font-bold text-foreground md:text-lg"
            >
              アカウントを削除しますか？
            </h2>
          </div>
        </div>

        {/* 本文（長い場合はここだけがスクロールする） */}
        <div className="min-h-0 flex-1 overflow-y-auto px-5 md:px-7">
          <p className="mb-3 text-[12.5px] leading-[1.8] text-foreground md:mb-3.5 md:text-[13.5px]">
            以下のデータがすべて削除され、
            <strong className="font-bold text-destructive">復元はできません</strong>。
          </p>

          <ul className="mb-4 overflow-hidden rounded-md border border-border md:mb-5">
            {DELETED_ITEMS.map((item) => (
              <li
                key={item}
                className="flex items-start gap-2.5 border-b border-border px-3.5 py-2.5 text-[12.5px] leading-[1.7] text-foreground last:border-b-0 md:px-4 md:py-3 md:text-[13px]"
              >
                <span
                  className="mt-[7px] h-1.5 w-1.5 flex-shrink-0 rounded-full bg-destructive"
                  aria-hidden="true"
                />
                {item}
              </li>
            ))}
          </ul>

          <div className="flex flex-col gap-2.5 pb-4 md:gap-3 md:pb-5">
            {NOTES.map((note) => (
              <p
                key={note}
                className="flex gap-2 text-xs leading-[1.75] text-muted-foreground md:gap-2.5 md:text-[12.5px]"
              >
                <Info
                  className="mt-[3px] h-[13px] w-[13px] flex-shrink-0 text-muted-foreground/70"
                  aria-hidden="true"
                />
                <span>{note}</span>
              </p>
            ))}

            {error && (
              <p className="rounded-md border border-destructive/20 bg-destructive/5 p-3 text-sm text-destructive">
                {error}
              </p>
            )}
          </div>
        </div>

        {/* フッター（同意と操作。本文の長さに関わらず常に見える） */}
        <div className="flex-shrink-0 border-t border-border bg-muted/30">
          <label
            htmlFor={agreeId}
            className="flex cursor-pointer items-start gap-2.5 px-5 pb-3 pt-3.5 md:px-7 md:pb-3.5 md:pt-4"
          >
            <input
              id={agreeId}
              type="checkbox"
              checked={agreed}
              onChange={(e) => setAgreed(e.target.checked)}
              disabled={isDeleting}
              className="mt-[1px] h-[18px] w-[18px] flex-shrink-0 cursor-pointer rounded-[4px] border-border accent-destructive focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed"
            />
            <span className="text-xs leading-[1.6] text-foreground md:text-[12.5px]">
              削除するとデータを復元できないことを理解しました
            </span>
          </label>

          <div className="flex flex-col-reverse gap-2.5 px-5 pb-5 md:flex-row md:items-center md:justify-end md:gap-3 md:px-7 md:pb-[18px]">
            <button
              ref={cancelButtonRef}
              type="button"
              onClick={onCancel}
              disabled={isDeleting}
              className="h-[42px] rounded-md border border-border bg-card px-[18px] text-[13.5px] font-semibold text-foreground transition-colors hover:bg-muted focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-60 md:h-10"
            >
              キャンセル
            </button>
            <button
              type="button"
              onClick={onConfirm}
              disabled={isDeleting || !agreed}
              className="h-[46px] rounded-md bg-destructive px-5 text-sm font-semibold text-destructive-foreground transition-colors hover:opacity-90 focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:bg-muted disabled:text-muted-foreground disabled:opacity-100 disabled:hover:opacity-100 md:h-10 md:text-[13.5px]"
            >
              {isDeleting ? '削除しています...' : 'アカウントを削除'}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
