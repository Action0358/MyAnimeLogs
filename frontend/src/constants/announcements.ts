/**
 * 運営者からのお知らせ。
 *
 * DBではなくこのファイルで管理する。バックエンド・管理画面が不要で静的生成でき、
 * 変更が git 履歴に残ってレビューできるため。お知らせは月に数件の想定で、
 * この運用で足りる。日次で増える・運営者が複数人になるといった状況になったら
 * DB + 管理画面への移行を検討する。
 *
 * **このファイルに書いて develop へマージした時点で公開される。**
 * 予約公開の仕組みは持たないので、下書きを置かないこと。
 * （静的エクスポートではビルド時刻と閲覧時刻がずれ、公開日での出し分けは
 *   ハイドレーション不整合の原因になる）
 */

export type AnnouncementCategory = 'release' | 'feature' | 'maintenance' | 'important'

/** 段落・リスト項目の中身。文字列か、リンクつきの文字列 */
export type AnnouncementInline = string | { text: string; href: string }

/** 詳細ページの本文を組むブロック */
export type AnnouncementBlock =
  | { type: 'p'; children: AnnouncementInline[] }
  | { type: 'heading'; text: string }
  | { type: 'ul'; items: AnnouncementInline[][] }

export interface Announcement {
  /** 一意なID。日付 + 短い識別子（例: 2026-09-06-beta） */
  id: string
  /** 公開日 YYYY-MM-DD */
  date: string
  category: AnnouncementCategory
  title: string
  /**
   * 本文（プレーン文字列）。一覧では2行に省略して出す導入文になる。
   * 改行はそのまま表示に反映されるが、段落を分けるときだけに使う。
   * 見た目の行の長さを整えるための改行は入れない（画面幅によって自動の
   * 折り返しと重なり、「す。」のように1文字だけの行ができるため）
   */
  body: string
  /**
   * 任意。詳細ページの本文を見出し・リスト・リンクつきで組みたいときに指定する。
   * 無いときは `body` を段落として詳細ページに出す。Markdown は導入せず、
   * 表せる要素を型で絞っている（依存を増やさず、管理画面も要らないため）
   */
  content?: AnnouncementBlock[]
  /** 任意。詳細ページの本文の後に置く導線 */
  link?: {
    href: string
    label: string
  }
}

/** カテゴリの表示名 */
export const ANNOUNCEMENT_CATEGORY_LABELS: Record<AnnouncementCategory, string> = {
  release: 'リリース',
  feature: '機能',
  maintenance: 'メンテナンス',
  important: '重要',
}

/**
 * お知らせの一覧。並び順は問わない（表示側で日付の降順に並べ替える）。
 * 新しいものを上に足していくと差分が読みやすい。
 */
export const ANNOUNCEMENTS: Announcement[] = [
  {
    id: '2026-10-28-dummy-01',
    date: '2026-10-28',
    category: 'feature',
    title: '【ダミー】feature のお知らせ 1（受け入れテスト用。削除予定）',
    body: '受け入れテスト（AT-035・036）用のダミーです。カテゴリ絞り込みとページネーションの確認後に削除します。',
  },
  {
    id: '2026-10-27-dummy-02',
    date: '2026-10-27',
    category: 'feature',
    title: '【ダミー】feature のお知らせ 2（受け入れテスト用。削除予定）',
    body: '受け入れテスト（AT-035・036）用のダミーです。カテゴリ絞り込みとページネーションの確認後に削除します。',
  },
  {
    id: '2026-10-26-dummy-03',
    date: '2026-10-26',
    category: 'feature',
    title: '【ダミー】feature のお知らせ 3（受け入れテスト用。削除予定）',
    body: '受け入れテスト（AT-035・036）用のダミーです。カテゴリ絞り込みとページネーションの確認後に削除します。',
  },
  {
    id: '2026-10-25-dummy-04',
    date: '2026-10-25',
    category: 'feature',
    title: '【ダミー】feature のお知らせ 4（受け入れテスト用。削除予定）',
    body: '受け入れテスト（AT-035・036）用のダミーです。カテゴリ絞り込みとページネーションの確認後に削除します。',
  },
  {
    id: '2026-10-24-dummy-05',
    date: '2026-10-24',
    category: 'feature',
    title: '【ダミー】feature のお知らせ 5（受け入れテスト用。削除予定）',
    body: '受け入れテスト（AT-035・036）用のダミーです。カテゴリ絞り込みとページネーションの確認後に削除します。',
  },
  {
    id: '2026-10-23-dummy-06',
    date: '2026-10-23',
    category: 'feature',
    title: '【ダミー】feature のお知らせ 6（受け入れテスト用。削除予定）',
    body: '受け入れテスト（AT-035・036）用のダミーです。カテゴリ絞り込みとページネーションの確認後に削除します。',
  },
  {
    id: '2026-10-22-dummy-07',
    date: '2026-10-22',
    category: 'feature',
    title: '【ダミー】feature のお知らせ 7（受け入れテスト用。削除予定）',
    body: '受け入れテスト（AT-035・036）用のダミーです。カテゴリ絞り込みとページネーションの確認後に削除します。',
  },
  {
    id: '2026-10-21-dummy-08',
    date: '2026-10-21',
    category: 'feature',
    title: '【ダミー】feature のお知らせ 8（受け入れテスト用。削除予定）',
    body: '受け入れテスト（AT-035・036）用のダミーです。カテゴリ絞り込みとページネーションの確認後に削除します。',
  },
  {
    id: '2026-10-20-dummy-09',
    date: '2026-10-20',
    category: 'feature',
    title: '【ダミー】feature のお知らせ 9（受け入れテスト用。削除予定）',
    body: '受け入れテスト（AT-035・036）用のダミーです。カテゴリ絞り込みとページネーションの確認後に削除します。',
  },
  {
    id: '2026-10-19-dummy-10',
    date: '2026-10-19',
    category: 'feature',
    title: '【ダミー】feature のお知らせ 10（受け入れテスト用。削除予定）',
    body: '受け入れテスト（AT-035・036）用のダミーです。カテゴリ絞り込みとページネーションの確認後に削除します。',
  },
  {
    id: '2026-10-18-dummy-11',
    date: '2026-10-18',
    category: 'release',
    title: '【ダミー】release のお知らせ 11（受け入れテスト用。削除予定）',
    body: '受け入れテスト（AT-035・036）用のダミーです。カテゴリ絞り込みとページネーションの確認後に削除します。',
  },
  {
    id: '2026-10-17-dummy-12',
    date: '2026-10-17',
    category: 'release',
    title: '【ダミー】release のお知らせ 12（受け入れテスト用。削除予定）',
    body: '受け入れテスト（AT-035・036）用のダミーです。カテゴリ絞り込みとページネーションの確認後に削除します。',
  },
  {
    id: '2026-10-16-dummy-13',
    date: '2026-10-16',
    category: 'release',
    title: '【ダミー】release のお知らせ 13（受け入れテスト用。削除予定）',
    body: '受け入れテスト（AT-035・036）用のダミーです。カテゴリ絞り込みとページネーションの確認後に削除します。',
  },
  {
    id: '2026-10-15-dummy-14',
    date: '2026-10-15',
    category: 'release',
    title: '【ダミー】release のお知らせ 14（受け入れテスト用。削除予定）',
    body: '受け入れテスト（AT-035・036）用のダミーです。カテゴリ絞り込みとページネーションの確認後に削除します。',
  },
  {
    id: '2026-10-14-dummy-15',
    date: '2026-10-14',
    category: 'release',
    title: '【ダミー】release のお知らせ 15（受け入れテスト用。削除予定）',
    body: '受け入れテスト（AT-035・036）用のダミーです。カテゴリ絞り込みとページネーションの確認後に削除します。',
  },
  {
    id: '2026-10-13-dummy-16',
    date: '2026-10-13',
    category: 'release',
    title: '【ダミー】release のお知らせ 16（受け入れテスト用。削除予定）',
    body: '受け入れテスト（AT-035・036）用のダミーです。カテゴリ絞り込みとページネーションの確認後に削除します。',
  },
  {
    id: '2026-10-12-dummy-17',
    date: '2026-10-12',
    category: 'release',
    title: '【ダミー】release のお知らせ 17（受け入れテスト用。削除予定）',
    body: '受け入れテスト（AT-035・036）用のダミーです。カテゴリ絞り込みとページネーションの確認後に削除します。',
  },
  {
    id: '2026-10-11-dummy-18',
    date: '2026-10-11',
    category: 'release',
    title: '【ダミー】release のお知らせ 18（受け入れテスト用。削除予定）',
    body: '受け入れテスト（AT-035・036）用のダミーです。カテゴリ絞り込みとページネーションの確認後に削除します。',
  },
  {
    id: '2026-10-10-dummy-19',
    date: '2026-10-10',
    category: 'maintenance',
    title: '【ダミー】maintenance のお知らせ 19（受け入れテスト用。削除予定）',
    body: '受け入れテスト（AT-035・036）用のダミーです。カテゴリ絞り込みとページネーションの確認後に削除します。',
  },
  {
    id: '2026-10-09-dummy-20',
    date: '2026-10-09',
    category: 'maintenance',
    title: '【ダミー】maintenance のお知らせ 20（受け入れテスト用。削除予定）',
    body: '受け入れテスト（AT-035・036）用のダミーです。カテゴリ絞り込みとページネーションの確認後に削除します。',
  },
  {
    id: '2026-10-08-dummy-21',
    date: '2026-10-08',
    category: 'maintenance',
    title: '【ダミー】maintenance のお知らせ 21（受け入れテスト用。削除予定）',
    body: '受け入れテスト（AT-035・036）用のダミーです。カテゴリ絞り込みとページネーションの確認後に削除します。',
  },
  {
    id: '2026-10-07-dummy-22',
    date: '2026-10-07',
    category: 'maintenance',
    title: '【ダミー】maintenance のお知らせ 22（受け入れテスト用。削除予定）',
    body: '受け入れテスト（AT-035・036）用のダミーです。カテゴリ絞り込みとページネーションの確認後に削除します。',
  },
  {
    id: '2026-10-06-dummy-23',
    date: '2026-10-06',
    category: 'important',
    title: '【ダミー】important のお知らせ 23（受け入れテスト用。削除予定）',
    body: '受け入れテスト（AT-035・036）用のダミーです。カテゴリ絞り込みとページネーションの確認後に削除します。',
  },
  {
    id: '2026-10-05-dummy-24',
    date: '2026-10-05',
    category: 'important',
    title: '【ダミー】important のお知らせ 24（受け入れテスト用。削除予定）',
    body: '受け入れテスト（AT-035・036）用のダミーです。カテゴリ絞り込みとページネーションの確認後に削除します。',
  },
  {
    id: '2026-10-04-dummy-25',
    date: '2026-10-04',
    category: 'important',
    title: '【ダミー】important のお知らせ 25（受け入れテスト用。削除予定）',
    body: '受け入れテスト（AT-035・036）用のダミーです。カテゴリ絞り込みとページネーションの確認後に削除します。',
  },
  {
    id: '2026-10-03-dummy-26',
    date: '2026-10-03',
    category: 'important',
    title: '【ダミー】important のお知らせ 26（受け入れテスト用。削除予定）',
    body: '受け入れテスト（AT-035・036）用のダミーです。カテゴリ絞り込みとページネーションの確認後に削除します。',
  },
  {
    id: '2026-09-06-news-start',
    date: '2026-09-06',
    category: 'feature',
    title: 'お知らせの掲載を開始しました',
    body: 'このページで、運営からのお知らせを掲載していきます。新機能の追加やメンテナンスの予定などをお伝えします。',
  },
]

/**
 * 新しい順に並べ替える。同じ日付のときはIDで並べ、順序が実行ごとにぶれないようにする。
 * 日付は YYYY-MM-DD 固定なので文字列比較で日付順になる（Date にパースしない）。
 */
export const sortAnnouncements = (items: Announcement[]): Announcement[] =>
  [...items].sort((a, b) =>
    a.date === b.date ? b.id.localeCompare(a.id) : b.date.localeCompare(a.date)
  )

/** 新しい順に並べ替えたお知らせを返す */
export const getAnnouncements = (): Announcement[] => sortAnnouncements(ANNOUNCEMENTS)

/** IDからお知らせを探す。無ければ undefined */
export const getAnnouncementById = (id: string): Announcement | undefined =>
  ANNOUNCEMENTS.find((a) => a.id === id)

/**
 * 該当お知らせの詳細ページのURL。
 *
 * 一覧・トップ・旧アンカー互換の転送が同じURLを指すよう、組み立てはここに集約する。
 * 詳細ページは #1111 で追加した（それ以前は #848 の決定で一覧 + アンカーだった）。
 */
export const announcementHref = (id: string): string => `/news/${id}`

/** トップページに出す件数。アニメカードが主役のため控えめにする */
export const HOME_ANNOUNCEMENT_COUNT = 3
