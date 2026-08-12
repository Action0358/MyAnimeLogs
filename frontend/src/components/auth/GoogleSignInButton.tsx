import React from 'react'

/** Google ブランドガイドラインに沿った公式カラーの G ロゴ */
const GoogleLogo: React.FC = () => (
  <svg width="20" height="20" viewBox="0 0 48 48" aria-hidden="true" focusable="false">
    <path
      fill="#EA4335"
      d="M24 9.5c3.54 0 6.71 1.22 9.21 3.6l6.85-6.85C35.9 2.38 30.47 0 24 0 14.62 0 6.51 5.38 2.56 13.22l7.98 6.19C12.43 13.72 17.74 9.5 24 9.5z"
    />
    <path
      fill="#4285F4"
      d="M46.98 24.55c0-1.57-.15-3.09-.38-4.55H24v9.02h12.94c-.58 2.96-2.26 5.48-4.78 7.18l7.73 6c4.51-4.18 7.09-10.36 7.09-17.65z"
    />
    <path
      fill="#FBBC05"
      d="M10.53 28.59c-.48-1.45-.76-2.99-.76-4.59s.27-3.14.76-4.59l-7.98-6.19C.92 16.46 0 20.12 0 24s.92 7.54 2.56 10.78l7.97-6.19z"
    />
    <path
      fill="#34A853"
      d="M24 48c6.48 0 11.93-2.13 15.89-5.81l-7.73-6c-2.15 1.45-4.92 2.3-8.16 2.3-6.26 0-11.57-4.22-13.47-9.91l-7.98 6.19C6.51 42.62 14.62 48 24 48z"
    />
  </svg>
)

interface GoogleSignInButtonProps {
  label: string
  onClick: () => void
  isLoading?: boolean
}

/**
 * Googleログインの唯一の入口。
 * 認証手段がこれ1つのため、フォーム内の主操作としての視覚的な重みを持たせている。
 */
export const GoogleSignInButton: React.FC<GoogleSignInButtonProps> = ({
  label,
  onClick,
  isLoading = false,
}) => (
  <button
    type="button"
    onClick={onClick}
    disabled={isLoading}
    /*
     * 枠線に汎用の border トークンを使わないのは、Google ブランドガイドラインが
     * light テーマのボタン枠線に #747775 相当の濃さを求めているため。
     * トークンの border（hsl(214 32% 91%)）では薄すぎてガイドライン非準拠になる。
     * 背景色も同様に、公式が認めるのは白／ライトグレー／黒の3配色のみで、
     * primary（インディゴ）で塗るのは違反にあたる。
     */
    className="flex h-14 w-full items-center justify-center gap-3 rounded-md border border-[#747775] bg-white text-[15px] font-semibold text-[#1f1f1f] shadow-sm transition-colors hover:bg-[#f8f9fa] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-60"
  >
    {isLoading ? (
      <>
        <span
          className="h-5 w-5 animate-spin rounded-full border-2 border-muted-foreground border-t-transparent"
          aria-hidden="true"
        />
        <span>Googleに接続しています...</span>
      </>
    ) : (
      <>
        <GoogleLogo />
        <span>{label}</span>
      </>
    )}
  </button>
)
