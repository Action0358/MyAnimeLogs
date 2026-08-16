'use client'

import React from 'react'
import { PRESET_AVATAR_IDS, resolveAvatarUrl } from '@/lib/avatars'

interface AvatarPickerProps {
  selectedIcon: string | null
  onSelect: (iconID: string) => void
  disabled?: boolean
}

/**
 * プリセットアバターの選択UI。
 *
 * 画像アップロードは #763 で廃止した。任意の画像を持ち込めると第三者の
 * 著作物が保存されうるため、運営が用意した候補からのみ選ばせる。
 *
 * 選択済みのものを再度押すと解除できる。解除専用のボタンを置くより
 * 操作が少なく、未設定へ戻す導線も確保できる。
 */
export const AvatarPicker: React.FC<AvatarPickerProps> = ({
  selectedIcon,
  onSelect,
  disabled = false,
}) => (
  <div
    role="radiogroup"
    aria-label="アイコンを選択"
    className="flex flex-wrap gap-2.5 md:gap-3"
  >
    {PRESET_AVATAR_IDS.map((id) => {
      const isSelected = selectedIcon === id
      return (
        <button
          key={id}
          type="button"
          role="radio"
          aria-checked={isSelected}
          aria-label={`アイコン ${id.replace('preset-', '')}`}
          onClick={() => onSelect(id)}
          disabled={disabled}
          className={`h-11 w-11 overflow-hidden rounded-full transition-all focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 md:h-12 md:w-12 ${
            isSelected
              ? 'ring-2 ring-primary ring-offset-2'
              : 'opacity-70 hover:opacity-100'
          }`}
        >
          {/* プリセットはフロントエンドの静的アセット。next/image を通す必要がない */}
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src={resolveAvatarUrl(id) ?? ''}
            alt=""
            className="h-full w-full object-cover"
          />
        </button>
      )
    })}
  </div>
)
