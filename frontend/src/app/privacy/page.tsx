import React from 'react'
import { LegalPage, LegalSection, LegalList } from '@/components/legal/LegalPage'
import { OPERATOR_NAME, SUPPORT_EMAIL } from '@/lib/constants'

export const metadata = {
  title: 'プライバシーポリシー | MyAnimeLogs',
  description: 'MyAnimeLogsにおける個人情報の取り扱いについて',
}

const PrivacyPage: React.FC = () => (
  <LegalPage title="プライバシーポリシー" updatedAt="2026年8月12日">
    <LegalSection title="1. はじめに">
      <p>
        本ポリシーは、{OPERATOR_NAME}（以下「運営者」）が提供する MyAnimeLogs（以下「本サービス」）における利用者の情報の取り扱いについて定めるものです。
      </p>
    </LegalSection>

    <LegalSection title="2. 取得する情報">
      <p>本サービスは、以下の情報を取得します。</p>
      <LegalList
        items={[
          <>
            <strong className="text-foreground">Googleアカウント情報</strong>
            ：認証時にGoogleから提供されるメールアドレス、プロフィール名、およびアカウントの識別子。
            本サービスがGoogleアカウントのパスワードを取得・保存することはありません。
          </>,
          <>
            <strong className="text-foreground">利用者が登録した記録データ</strong>
            ：視聴状況、お気に入り、プロフィールの自己紹介・好きな作品・SNSリンク、アイコン画像。
          </>,
          <>
            <strong className="text-foreground">技術的な情報</strong>
            ：アクセス日時、IPアドレス、ブラウザの種類など、サービスの提供とセキュリティ確保のために自動的に記録される情報。
          </>,
        ]}
      />
    </LegalSection>

    <LegalSection title="3. 利用目的">
      <LegalList
        items={[
          '本サービスの提供、維持、および利用者の認証のため',
          '利用者の記録データの保存および表示のため',
          '不正アクセスや規約違反への対応など、セキュリティの確保のため',
          'サービスの改善および新機能の検討のため（個人を特定しない統計処理を含む）',
          '利用者からのお問い合わせへの対応のため',
        ]}
      />
    </LegalSection>

    <LegalSection title="4. 第三者提供">
      <p>
        運営者は、以下の場合を除き、利用者の個人情報を第三者に提供しません。
      </p>
      <LegalList
        items={[
          '利用者本人の同意がある場合',
          '法令に基づく場合',
          '人の生命、身体または財産の保護のために必要があり、本人の同意を得ることが困難な場合',
        ]}
      />
      <p>
        利用者の記録データを、広告目的で第三者へ販売または提供することはありません。
      </p>
    </LegalSection>

    <LegalSection title="5. 外部サービスの利用">
      <p>本サービスは、以下の外部サービスを利用しています。</p>
      <LegalList
        items={[
          <>
            <strong className="text-foreground">Google（認証）</strong>
            ：ログイン認証のために利用します。Googleにおける情報の取り扱いは、Googleのプライバシーポリシーに従います。
          </>,
          <>
            <strong className="text-foreground">Supabase（認証基盤・データベース）</strong>
            ：認証情報および利用者の記録データの保存に利用します。
          </>,
          <>
            <strong className="text-foreground">
              Annict / AniList / Jikan（MyAnimeList） / TMDB
            </strong>
            ：アニメ作品情報の取得に利用します。これらへ利用者の個人情報を送信することはありません。
          </>,
          <>
            <strong className="text-foreground">Fly.io / Cloudflare</strong>
            ：本サービスのホスティングおよび配信に利用します。
          </>,
        ]}
      />
    </LegalSection>

    <LegalSection title="6. Cookie等の利用">
      <p>
        本サービスは、ログイン状態を維持するためにブラウザのローカルストレージおよびCookieを使用します。
        これらは認証の維持を目的とするものであり、広告配信を目的とした追跡は行っていません。
      </p>
    </LegalSection>

    <LegalSection title="7. 情報の保管期間">
      <p>
        利用者の情報は、アカウントが存続する間、保管されます。
        アカウントを削除した場合、記録データは削除されます。
        ただし、法令により保管が義務付けられている情報については、当該期間中保管します。
      </p>
    </LegalSection>

    <LegalSection title="8. 利用者の権利">
      <p>利用者は、自己の情報について以下を行うことができます。</p>
      <LegalList
        items={[
          'プロフィール画面からの情報の確認および訂正',
          'データエクスポート機能による、自己の記録データの取得',
          'アカウントの削除による、情報の削除',
        ]}
      />
    </LegalSection>

    <LegalSection title="9. セキュリティ">
      <p>
        運営者は、利用者の情報の漏洩、滅失または毀損の防止のために、通信の暗号化やアクセス制御など合理的な安全管理措置を講じます。
        認証にGoogleアカウントを用いることで、本サービスがパスワードを保持しない構成としています。
      </p>
    </LegalSection>

    <LegalSection title="10. 本ポリシーの変更">
      <p>
        運営者は、必要に応じて本ポリシーを変更することがあります。
        変更後のポリシーは、本ページに掲載した時点から効力を生じるものとします。
        重要な変更を行う場合は、本サービス上で告知します。
      </p>
    </LegalSection>

    <LegalSection title="11. お問い合わせ">
      <p>
        本ポリシーに関するお問い合わせ、および保有する情報の開示・訂正・削除のご請求は、以下までご連絡ください。
      </p>
      <p>
        運営者: {OPERATOR_NAME}
        <br />
        連絡先:{' '}
        <a href={`mailto:${SUPPORT_EMAIL}`} className="text-primary hover:underline">
          {SUPPORT_EMAIL}
        </a>
      </p>
    </LegalSection>
  </LegalPage>
)

export default PrivacyPage
