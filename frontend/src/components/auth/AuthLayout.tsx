import React from 'react'

interface AuthLayoutProps {
  /** 左（モバイルでは上）のビジュアルパネル */
  panel: React.ReactNode
  children: React.ReactNode
}

/**
 * 認証画面の2分割レイアウト。
 * デスクトップは左ビジュアル / 右フォームの等幅2カラム、モバイルでは縦積みになる。
 */
export const AuthLayout: React.FC<AuthLayoutProps> = ({ panel, children }) => (
  <div className="min-h-screen bg-background">
    <div className="mx-auto grid max-w-[1180px] grid-cols-1 lg:min-h-screen lg:grid-cols-2">
      {panel}
      <div className="flex items-center justify-center px-5 py-8 lg:px-14 lg:py-12">
        <div className="w-full max-w-[360px]">{children}</div>
      </div>
    </div>
  </div>
)
