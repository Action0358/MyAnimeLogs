import React from 'react'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import '@testing-library/jest-dom'
import { RelatedAnimeGrid, RELATED_INITIAL_COUNT } from './RelatedAnimeGrid'
import type { RelatedAnime } from '@/lib/api/generated/myanimelogs'

// AnimeCard は next/link や画像最適化に依存するため、
// ここでは「何件描画されたか」だけを見たい。グリッドはタイトル一覧に差し替える。
jest.mock('./AnimeGrid', () => ({
  AnimeGrid: ({ animes }: { animes: { id: number; title: string }[] }) => (
    <ul data-testid="anime-grid">
      {animes.map((a) => (
        <li key={a.id}>{a.title}</li>
      ))}
    </ul>
  ),
}))

const buildRelated = (count: number): RelatedAnime[] =>
  Array.from({ length: count }, (_, i) => ({
    id: i + 1,
    title: `作品${i + 1}`,
    season: '2020-spring',
    image_url: null,
    cover_image_extra_large: undefined,
    cover_image_large: undefined,
    cover_image_medium: undefined,
  })) as RelatedAnime[]

const renderedTitles = () =>
  Array.from(screen.getByTestId('anime-grid').querySelectorAll('li')).map((li) => li.textContent)

/**
 * 関連作品の折りたたみを検証する（#829）。
 *
 * バックエンドの上限が12→48に上がったため、名探偵コナン36作品のような
 * シリーズをそのまま並べると詳細ページが極端に長くなる。
 */
describe('RelatedAnimeGrid', () => {
  it('12件以下なら全部出し、展開ボタンを出さない', () => {
    render(<RelatedAnimeGrid related={buildRelated(RELATED_INITIAL_COUNT)} />)

    expect(renderedTitles()).toHaveLength(RELATED_INITIAL_COUNT)
    expect(screen.queryByRole('button')).not.toBeInTheDocument()
  })

  it('12件を超えたら12件だけ出し、残り件数を明示する', () => {
    render(<RelatedAnimeGrid related={buildRelated(36)} />)

    expect(renderedTitles()).toHaveLength(RELATED_INITIAL_COUNT)
    expect(screen.getByRole('button', { name: /他24件を表示/ })).toBeInTheDocument()
  })

  it('展開すると全件出る', async () => {
    render(<RelatedAnimeGrid related={buildRelated(36)} />)

    await userEvent.click(screen.getByRole('button', { name: /他24件を表示/ }))

    expect(renderedTitles()).toHaveLength(36)
    // 打ち切られていた新しい側の作品まで到達できること
    expect(screen.getByText('作品36')).toBeInTheDocument()
  })

  it('展開後は折りたたみに戻せる', async () => {
    render(<RelatedAnimeGrid related={buildRelated(20)} />)

    const trigger = screen.getByRole('button')
    expect(trigger).toHaveAttribute('aria-expanded', 'false')

    await userEvent.click(trigger)
    expect(trigger).toHaveAttribute('aria-expanded', 'true')

    await userEvent.click(screen.getByRole('button', { name: /折りたたむ/ }))
    expect(renderedTitles()).toHaveLength(RELATED_INITIAL_COUNT)
  })

  it('残りが1行（6件）以内なら、ボタンを出さず全件表示する', () => {
    // 12件+6件=18件。最も列数が多いPC（6列）の1行ぶんに収まるため、
    // クリックさせずに全件出す（アニメ詳細UI刷新v2）
    render(<RelatedAnimeGrid related={buildRelated(RELATED_INITIAL_COUNT + 6)} />)

    expect(renderedTitles()).toHaveLength(RELATED_INITIAL_COUNT + 6)
    expect(screen.queryByRole('button')).not.toBeInTheDocument()
  })

  it('残りが1行を超えたら、12件だけ出してボタンで折りたたむ（境界）', () => {
    // 12件+7件=19件。1行(6件)を超えるため折りたたむ
    render(<RelatedAnimeGrid related={buildRelated(RELATED_INITIAL_COUNT + 7)} />)

    expect(renderedTitles()).toHaveLength(RELATED_INITIAL_COUNT)
    expect(screen.getByRole('button', { name: /他7件を表示/ })).toBeInTheDocument()
  })

  it('先頭から順に表示する（放送順を崩さない）', () => {
    render(<RelatedAnimeGrid related={buildRelated(36)} />)

    expect(renderedTitles()[0]).toBe('作品1')
    expect(renderedTitles()[RELATED_INITIAL_COUNT - 1]).toBe(`作品${RELATED_INITIAL_COUNT}`)
  })
})
