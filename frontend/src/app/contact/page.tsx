import type React from 'react'
import type { Metadata } from 'next'
import { ExternalLink } from 'lucide-react'
import { Container } from '@/components/layout'
import { CopyEmailButton } from '@/components/contact/CopyEmailButton'
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
 * ページ本体は API を叩かないため静的生成。コピー操作だけ CopyEmailButton に
 * 切り出してクライアント化している。
 */
const ContactPage: React.FC = () => (
  <Container size="xl" className="max-w-[1180px] py-8">
    {/* 本文はコンテナ中央の読み幅列（760px）に置く。ヘッダー・フッターと同じ
        共通コンテナ幅に揃え、/news と同じ扱いにする（モックアップ「Contact Redesign」修正版） */}
    <div className="mx-auto max-w-[760px]">
      {/* 見出しの体裁は今期アニメ／人気アニメ／検索／お知らせと揃える */}
      <h1 className="mb-6 text-xl md:text-[26px] font-bold text-foreground tracking-tight">
        お問い合わせ
      </h1>

      <section className="rounded-lg border border-border bg-card px-5 py-6 md:px-6 md:py-7">
        <h2 className="mb-2 text-base font-bold text-foreground md:text-lg">
          ご意見・不具合のご報告
        </h2>
        {/* text-pretty で追い出し改行を防ぐ。JSX内で改行すると半角スペースが
            挿入されて「ください。 詳しくは」のような不要な空白ができるため、
            1行にまとめている */}
        <p className="mb-5 text-sm leading-[1.8] text-foreground/80 text-pretty">
          不具合のご報告、機能のご要望、作品情報の誤りなどは、お問い合わせフォームからお寄せください。ご返信をご希望の場合のみメールアドレスをご記入いただければ、そのアドレス宛にご連絡します。
        </p>

        <a
          href={SUPPORT_FORM_URL}
          target="_blank"
          rel="noopener noreferrer"
          className="flex h-11 w-full items-center justify-center gap-2 rounded-md bg-primary px-5 text-sm font-bold text-primary-foreground transition-colors hover:bg-primary/90 focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 sm:w-auto"
        >
          お問い合わせフォームを開く
          {/* 別タブで開くことは視覚的にはアイコンで示す。アイコンは aria-hidden のため、
              読み上げにはここで伝える（#879） */}
          <span className="sr-only">（新しいタブで開きます）</span>
          <ExternalLink className="h-4 w-4" aria-hidden="true" />
        </a>
      </section>

      <section className="mt-5 rounded-lg border border-border bg-card px-5 py-6 md:px-6 md:py-7">
        <h2 className="mb-2 text-base font-bold text-foreground md:text-lg">
          個人情報に関するご請求
        </h2>
        <p className="mb-5 text-sm leading-[1.8] text-foreground/80 text-pretty">
          登録情報の開示・訂正・削除・利用停止のご請求は、以下のメールアドレスへご連絡ください。詳しくは<a href="/privacy" className="text-primary hover:underline">プライバシーポリシー</a>をご確認ください。
        </p>

        <div className="flex items-center justify-between gap-4 rounded-md border border-border bg-muted/40 px-4 py-3">
          <div className="min-w-0">
            <div className="mb-0.5 text-[11px] text-muted-foreground">請求の送付先</div>
            <a
              href={`mailto:${SUPPORT_EMAIL}`}
              className="break-all text-sm font-semibold text-primary hover:underline"
            >
              {SUPPORT_EMAIL}
            </a>
          </div>
          <CopyEmailButton email={SUPPORT_EMAIL} />
        </div>
      </section>
    </div>
  </Container>
)

export default ContactPage
