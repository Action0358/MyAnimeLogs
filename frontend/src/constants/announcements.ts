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

export interface Announcement {
  /** 一意なID。日付 + 短い識別子（例: 2026-09-06-beta） */
  id: string
  /** 公開日 YYYY-MM-DD */
  date: string
  category: AnnouncementCategory
  title: string
  /**
   * 本文。改行はそのまま表示に反映されるが、段落を分けるときだけに使う。
   * 見た目の行の長さを整えるための改行は入れない（画面幅によって自動の
   * 折り返しと重なり、「す。」のように1文字だけの行ができるため）
   */
  body: string
  /** 任意。本文の後に置く導線 */
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
    a.date === b.date ? b.id.localeCompare(a.id) : b.date.localeCompare(a.date),
  )

/** 新しい順に並べ替えたお知らせを返す */
export const getAnnouncements = (): Announcement[] => sortAnnouncements(ANNOUNCEMENTS)

/**
 * 一覧ページ内の該当お知らせを指すURL。
 *
 * 1件ごとの詳細ページは持たない（#848）ため、遷移先は一覧ページ + アンカー。
 * リンクを張る側と id を出力する側でずれないよう、組み立てはここに集約する。
 *
 * id は日付始まりで数字から始まるが、HTML の id 属性としては有効で
 * getElementById・アンカー遷移・:target のいずれも動く。
 * ただし CSS や querySelector の id セレクタ（#2026-...）は不正になるため、
 * 要素を取るときは getElementById を使うこと。
 */
export const announcementHref = (id: string): string => `/news#${id}`

/** トップページに出す件数。アニメカードが主役のため控えめにする */
export const HOME_ANNOUNCEMENT_COUNT = 3
