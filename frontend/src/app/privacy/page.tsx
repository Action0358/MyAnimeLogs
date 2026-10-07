import React from 'react'
import {
  LegalTocPage,
  LegalArticle,
  LegalDefinitionList,
  LegalDotList,
  LegalContactCard,
  type LegalTocItem,
} from '@/components/legal/LegalTocPage'
import { OPERATOR_NAME, SUPPORT_EMAIL } from '@/lib/constants'

export const metadata = {
  title: 'プライバシーポリシー | MyAnimeLogs',
  description: 'MyAnimeLogsにおける個人情報の取り扱いについて',
}

const SECTION_IDS = [
  'intro',
  'collected-info',
  'purpose',
  'third-party',
  'external-services',
  'cookies',
  'retention',
  'user-rights',
  'security',
  'policy-changes',
] as const

const TOC_ITEMS: LegalTocItem[] = [
  { id: SECTION_IDS[0], label: 'はじめに' },
  { id: SECTION_IDS[1], label: '取得する情報' },
  { id: SECTION_IDS[2], label: '利用目的' },
  { id: SECTION_IDS[3], label: '第三者提供' },
  { id: SECTION_IDS[4], label: '外部サービスの利用' },
  { id: SECTION_IDS[5], label: 'Cookie等の利用' },
  { id: SECTION_IDS[6], label: '情報の保管期間' },
  { id: SECTION_IDS[7], label: '利用者の権利' },
  { id: SECTION_IDS[8], label: 'セキュリティ' },
  { id: SECTION_IDS[9], label: '本ポリシーの変更' },
]

const PrivacyPage: React.FC = () => (
  <LegalTocPage title="プライバシーポリシー" tocItems={TOC_ITEMS} extraTocLabel="お問い合わせ" extraTocHref="#contact">
    <LegalArticle id={SECTION_IDS[0]} num="第1節" title="はじめに">
      <p>
        本ポリシーは、{OPERATOR_NAME}（以下「運営者」）が提供する MyAnimeLogs（以下「本サービス」）における利用者の情報の取り扱いについて定めるものです。
      </p>
    </LegalArticle>

    <LegalArticle id={SECTION_IDS[1]} num="第2節" title="取得する情報">
      <p>本サービスは、以下の情報を取得します。</p>
      <LegalDefinitionList
        items={[
          {
            term: 'Googleアカウント情報',
            desc: '認証時にGoogleから提供されるメールアドレス、プロフィール名、およびアカウントの識別子。本サービスがGoogleアカウントのパスワードを取得・保存することはありません。',
          },
          {
            term: '利用者が登録した記録データ',
            desc: '視聴状況、お気に入り、プロフィールの自己紹介・好きな作品・SNSリンク、アイコン画像。',
          },
          {
            term: '技術的な情報',
            desc: 'アクセス日時、IPアドレス、ブラウザの種類など、サービスの提供とセキュリティ確保のために自動的に記録される情報。',
          },
        ]}
      />
    </LegalArticle>

    <LegalArticle id={SECTION_IDS[2]} num="第3節" title="利用目的">
      <LegalDotList
        items={[
          '本サービスの提供、維持、および利用者の認証のため',
          '利用者の記録データの保存および表示のため',
          '不正アクセスや規約違反への対応など、セキュリティの確保のため',
          'サービスの改善および新機能の検討のため（個人を特定しない統計処理を含む）',
          '利用者からのお問い合わせへの対応のため',
        ]}
      />
    </LegalArticle>

    <LegalArticle id={SECTION_IDS[3]} num="第4節" title="第三者提供">
      <p>運営者は、以下の場合を除き、利用者の個人情報を第三者に提供しません。</p>
      <LegalDotList
        items={[
          '利用者本人の同意がある場合',
          '法令に基づく場合',
          '人の生命、身体または財産の保護のために必要があり、本人の同意を得ることが困難な場合',
        ]}
      />
      <p>利用者の記録データを、広告目的で第三者へ販売または提供することはありません。</p>
    </LegalArticle>

    <LegalArticle id={SECTION_IDS[4]} num="第5節" title="外部サービスの利用">
      <p>本サービスは、以下の外部サービスを利用しています。</p>
      <LegalDefinitionList
        items={[
          {
            term: 'Google（認証）',
            desc: 'ログイン認証のために利用します。Googleにおける情報の取り扱いは、Googleのプライバシーポリシーに従います。',
          },
          {
            term: 'Supabase（認証基盤・データベース）',
            desc: '認証情報および利用者の記録データの保存に利用します。',
          },
          {
            term: 'Annict / AniList / Jikan（MyAnimeList）/ TMDB',
            desc: 'アニメ作品情報の取得に利用します。これらへ利用者の個人情報を送信することはありません。',
          },
          {
            term: 'Fly.io / Cloudflare',
            desc: '本サービスのホスティングおよび配信に利用します。また、Cloudflareのアクセス解析機能（Cloudflare Web Analytics）により、訪問状況を把握するための統計情報を取得します。',
          },
        ]}
      />
    </LegalArticle>

    <LegalArticle id={SECTION_IDS[5]} num="第6節" title="Cookie等の利用">
      <p>
        本サービスは、ログイン状態を維持するためにブラウザのローカルストレージおよびCookieを使用します。これらは認証の維持を目的とするものであり、広告配信を目的とした追跡は行っていません。
      </p>
      <p>
        また、前節のアクセス解析（Cloudflare Web
        Analytics）はCookieを使用せず、個人を特定できない形で訪問状況の統計情報のみを取得します。
      </p>
    </LegalArticle>

    <LegalArticle id={SECTION_IDS[6]} num="第7節" title="情報の保管期間">
      <p>
        利用者の情報は、アカウントが存続する間、保管されます。アカウントを削除した場合、記録データは削除されます。ただし、法令により保管が義務付けられている情報については、当該期間中保管します。
      </p>
    </LegalArticle>

    <LegalArticle id={SECTION_IDS[7]} num="第8節" title="利用者の権利">
      <p>利用者は、自己の情報について以下を行うことができます。</p>
      <LegalDotList
        items={[
          'プロフィール画面からの情報の確認および訂正',
          'データエクスポート機能による、自己の記録データの取得',
          'アカウントの削除による、情報の削除',
        ]}
      />
    </LegalArticle>

    <LegalArticle id={SECTION_IDS[8]} num="第9節" title="セキュリティ">
      <p>
        運営者は、利用者の情報の漏洩、滅失または毀損の防止のために、通信の暗号化やアクセス制御など合理的な安全管理措置を講じます。認証にGoogleアカウントを用いることで、本サービスがパスワードを保持しない構成としています。
      </p>
    </LegalArticle>

    <LegalArticle id={SECTION_IDS[9]} num="第10節" title="本ポリシーの変更">
      <p>
        運営者は、必要に応じて本ポリシーを変更することがあります。変更後のポリシーは、本ページに掲載した時点から効力を生じるものとします。重要な変更を行う場合は、本サービス上で告知します。
      </p>
    </LegalArticle>

    <LegalContactCard id="contact" operator={OPERATOR_NAME} email={SUPPORT_EMAIL} />
  </LegalTocPage>
)

export default PrivacyPage
