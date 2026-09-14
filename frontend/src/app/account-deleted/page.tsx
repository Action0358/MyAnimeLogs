import React from 'react'
import Link from 'next/link'
import type { Route } from 'next'
import { Container } from '@/components/layout'

export const metadata = {
  title: 'アカウント削除完了 | MyAnimeLogs',
  description: 'アカウントの削除が完了しました',
}

/**
 * アカウント削除の完了画面。
 *
 * 削除後にトップページへ戻すだけだと、削除できたのか分からない。
 * さらに Google ログインは「初回なら登録・以降はログイン」を兼ねるため、
 * 再度ログインすると新しいアカウントが作られる。これを事前に伝えないと
 * 「削除したのに復活した」と誤解される。
 *
 * 未ログイン状態で表示される画面のため ProtectedRoute で囲まない。
 */
const AccountDeletedPage: React.FC = () => (
  <Container className="py-16">
    <div className="mx-auto max-w-lg text-center">
      <h1 className="text-2xl font-bold text-foreground">アカウントを削除しました</h1>
      <p className="mt-4 text-sm leading-relaxed text-muted-foreground">
        ご利用ありがとうございました。
        <br />
        プロフィール・お気に入り・視聴記録はすべて削除されました。
      </p>

      <div className="mt-8 rounded-lg border border-border bg-card p-5 text-left">
        <h2 className="text-sm font-bold text-foreground">再度ご利用になる場合</h2>
        <p className="mt-2 text-[13px] leading-relaxed text-muted-foreground">
          同じGoogleアカウントでログインすると、<strong className="text-foreground">新しいアカウントとして登録されます</strong>。
          削除した視聴記録やお気に入りは復元されません。
        </p>
      </div>

      <div className="mt-4 rounded-lg border border-border bg-card p-5 text-left">
        <h2 className="text-sm font-bold text-foreground">Googleアカウント側の連携について</h2>
        <p className="mt-2 text-[13px] leading-relaxed text-muted-foreground">
          Googleアカウントの設定に「MyAnimeLogs へのアクセス許可」が残っている場合があります。
          必要に応じて、Googleのアカウント設定画面から解除してください。
        </p>
      </div>

      <div className="mt-8 flex flex-col items-center gap-3">
        <Link
          href="/"
          className="inline-flex h-11 items-center justify-center rounded-md bg-primary px-6 text-sm font-semibold text-primary-foreground transition-colors hover:bg-primary/90"
        >
          トップページへ
        </Link>
        {/* 削除後はログインできないため、ログイン不要で送れるフォームを主に案内する。
            退会理由のフィードバックもここから受け取れる */}
        <p className="text-xs text-muted-foreground">
          お問い合わせ:{' '}
          <Link href={'/contact' as Route} className="text-primary hover:underline">
            お問い合わせフォーム
          </Link>
        </p>
      </div>
    </div>
  </Container>
)

export default AccountDeletedPage
