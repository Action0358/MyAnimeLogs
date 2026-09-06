import {
  ANNOUNCEMENTS,
  ANNOUNCEMENT_CATEGORY_LABELS,
  sortAnnouncements,
  type Announcement,
  type AnnouncementCategory,
} from '../announcements'

const build = (id: string, date: string, extra: Partial<Announcement> = {}): Announcement => ({
  id,
  date,
  category: 'release',
  title: `${id} のタイトル`,
  body: '本文',
  ...extra,
})

describe('sortAnnouncements', () => {
  it('新しい順に並べる', () => {
    const sorted = sortAnnouncements([
      build('a', '2026-08-01'),
      build('b', '2026-09-06'),
      build('c', '2026-08-20'),
    ])

    expect(sorted.map((a) => a.id)).toEqual(['b', 'c', 'a'])
  })

  it('同じ日付でも順序がぶれない', () => {
    const items = [build('z', '2026-09-06'), build('a', '2026-09-06')]

    // 入力順を変えても結果が同じであること
    expect(sortAnnouncements(items).map((a) => a.id)).toEqual(['z', 'a'])
    expect(sortAnnouncements([...items].reverse()).map((a) => a.id)).toEqual(['z', 'a'])
  })

  it('元の配列を書き換えない', () => {
    const items = [build('a', '2026-08-01'), build('b', '2026-09-06')]
    sortAnnouncements(items)

    expect(items.map((a) => a.id)).toEqual(['a', 'b'])
  })

  // 月・日が1桁でも文字列比較で正しく並ぶことを担保する（YYYY-MM-DD 固定が前提）
  it('月をまたいでも桁数に引きずられない', () => {
    const sorted = sortAnnouncements([
      build('a', '2026-09-09'),
      build('b', '2026-10-01'),
      build('c', '2026-09-10'),
    ])

    expect(sorted.map((a) => a.id)).toEqual(['b', 'c', 'a'])
  })
})

describe('カテゴリ定義', () => {
  it('すべてのカテゴリに表示名がある', () => {
    const categories: AnnouncementCategory[] = ['release', 'feature', 'maintenance', 'important']

    for (const category of categories) {
      expect(ANNOUNCEMENT_CATEGORY_LABELS[category]).toBeTruthy()
    }
  })
})

/**
 * お知らせは手書きのファイルで管理するため、書き間違いをここで止める。
 * 現時点で ANNOUNCEMENTS は空だが、追加したときに効くようにしておく。
 */
describe('ANNOUNCEMENTS の記述', () => {
  it('IDが重複していない', () => {
    const ids = ANNOUNCEMENTS.map((a) => a.id)

    expect(new Set(ids).size).toBe(ids.length)
  })

  it('日付が YYYY-MM-DD 形式', () => {
    for (const announcement of ANNOUNCEMENTS) {
      expect(announcement.date).toMatch(/^\d{4}-\d{2}-\d{2}$/)
    }
  })

  it('タイトルと本文が空でない', () => {
    for (const announcement of ANNOUNCEMENTS) {
      expect(announcement.title.trim()).not.toBe('')
      expect(announcement.body.trim()).not.toBe('')
    }
  })
})
