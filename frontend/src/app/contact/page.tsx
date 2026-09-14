import type React from 'react'
import type { Metadata } from 'next'
import { ExternalLink } from 'lucide-react'
import { Container } from '@/components/layout'
import { SUPPORT_EMAIL, SUPPORT_FORM_URL } from '@/lib/constants'

export const metadata: Metadata = {
  title: 'お問い合わせ | MyAnimeLogs',
  description: 'MyAnimeLogs へのお問い合わせ窓口です。',
}

/**
 * お問い合わせページ。
 *
 * 一般の問い合わせはGoogleフォームへ誘導し、メールは個人情報に関する請求の
 * 窓口として併記する。フォームを主導線にするのは、メールだと設定次第で
 * 届かないまま消える経路があるのに対し、フォームは必ず記録が残るため（#869）。
 *
 * API を叩かないためクライアントコンポーネントにする必要がなく、静的生成される。
 */
const ContactPage: React.FC = () => (
  <Container size="xl" className="max-w-[1180px] py-8">
    {/* 読みやすさを優先して本文幅を絞る。お知らせ一覧（/news）と同じ扱い */}
    <div className="max-w-[720px]">
      {/* 見出しの体裁は今期アニメ／人気アニメ／検索／お知らせと揃える */}
      <h1 className="mb-6 text-xl md:text-[26px] font-bold text-foreground tracking-tight">
        お問い合わせ
      </h1>

      <section className="rounded-lg border border-border bg-card px-5 py-6 md:px-6 md:py-7">
        <h2 className="mb-2 text-base font-bold text-foreground md:text-lg">
          ご意見・不具合のご報告
        </h2>
        <p className="mb-5 text-sm leading-[1.8] text-foreground/80">
          不具合のご報告、機能のご要望、作品情報の誤りなどは、お問い合わせフォームからお寄せください。
          ログインは不要です。ご返信をご希望の場合のみメールアドレスをご記入いただければ、
          そのアドレス宛にご連絡します。
        </p>

        <a
          href={SUPPORT_FORM_URL}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex items-center gap-2 rounded-full bg-primary px-5 py-2.5 text-sm font-bold text-primary-foreground transition-colors hover:bg-primary/90 focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
        >
          お問い合わせフォームを開く
          <ExternalLink className="h-4 w-4" aria-hidden="true" />
        </a>

        {/* 外部サイトへ遷移することを事前に伝える。別タブで開く挙動の予告も兼ねる */}
        <p className="mt-3 text-xs text-muted-foreground">
          Googleフォームが別のタブで開きます。
        </p>
      </section>

      <section className="mt-5 rounded-lg border border-border bg-card px-5 py-6 md:px-6 md:py-7">
        <h2 className="mb-2 text-base font-bold text-foreground md:text-lg">
          個人情報に関するご請求
        </h2>
        <p className="text-sm leading-[1.8] text-foreground/80">
          登録情報の開示・訂正・削除・利用停止のご請求は、以下のメールアドレスへご連絡ください。
          詳しくは
          <a href="/privacy" className="mx-1 text-primary hover:underline">
            プライバシーポリシー
          </a>
          をご確認ください。
        </p>
        <p className="mt-3 text-sm">
          <a
            href={`mailto:${SUPPORT_EMAIL}`}
            className="break-all font-medium text-primary hover:underline"
          >
            {SUPPORT_EMAIL}
          </a>
        </p>
      </section>
    </div>
  </Container>
)

export default ContactPage
