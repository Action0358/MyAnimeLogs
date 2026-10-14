import React from 'react'
import { ImageOff } from 'lucide-react'
import { cn } from '@/lib/utils'

interface PosterFallbackProps {
  className?: string
}

/**
 * ポスター画像が無い/読み込みに失敗した際の共通フォールバック。
 * 検索・一覧・関連作品・詳細ページのメインポスターすべてで使う
 * （モックアップ「No Poster Redesign」A案採用）。
 *
 * 絵文字や「NO IMAGE」+タイトル重複など画面ごとにバラバラだった表示を統一する。
 * 色はアクセント(primary)を使わずslateのみにし、どのポスターに混ざっても浮かないようにする。
 */
export const PosterFallback: React.FC<PosterFallbackProps> = ({ className }) => (
  <div
    className={cn(
      'absolute inset-0 flex flex-col items-center justify-center gap-1.5 bg-slate-100 text-slate-400 sm:gap-2',
      className
    )}
  >
    <ImageOff className="h-[22px] w-[22px] sm:h-7 sm:w-7" strokeWidth={1.5} aria-hidden="true" />
    <span className="text-[9.5px] font-semibold tracking-wide sm:text-[10.5px]">画像準備中</span>
  </div>
)

export default PosterFallback
