import React from 'react'
import { act, render } from '@testing-library/react'
import { BuildSkewGuard } from './BuildSkewGuard'
import { fetchLatestBuildId, getClientBuildId } from '@/lib/buildSkew'

jest.mock('@/lib/buildSkew')

const mockGetClientBuildId = getClientBuildId as jest.Mock
const mockFetchLatestBuildId = fetchLatestBuildId as jest.Mock

const reload = jest.fn()
const assign = jest.fn()
const originalLocation = window.location

const setLocation = (pathname: string) => {
  Object.defineProperty(window, 'location', {
    configurable: true,
    value: {
      href: `http://localhost${pathname}`,
      origin: 'http://localhost',
      pathname,
      search: '',
      reload,
      assign,
    },
  })
}

const setVisibility = (state: 'visible' | 'hidden') => {
  Object.defineProperty(document, 'visibilityState', { configurable: true, value: state })
}

/** タブが前面に戻ったことにして、確認の完了まで待つ */
const returnToTab = async () => {
  setVisibility('visible')
  await act(async () => {
    document.dispatchEvent(new Event('visibilitychange'))
  })
}

describe('BuildSkewGuard', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    jest.useFakeTimers()
    sessionStorage.clear()
    setLocation('/mypage/')
    setVisibility('visible')
    mockGetClientBuildId.mockReturnValue('build-OLD')
    mockFetchLatestBuildId.mockResolvedValue('build-OLD')
  })

  afterEach(() => {
    jest.useRealTimers()
    Object.defineProperty(window, 'location', { configurable: true, value: originalLocation })
    document.body.innerHTML = ''
  })

  it('最新と同じなら、何もしない', async () => {
    render(<BuildSkewGuard />)

    await returnToTab()

    expect(mockFetchLatestBuildId).toHaveBeenCalled()
    expect(reload).not.toHaveBeenCalled()
  })

  it('古いタブが前面に戻ったら、読み込み直す', async () => {
    mockFetchLatestBuildId.mockResolvedValue('build-NEW')
    render(<BuildSkewGuard />)

    await returnToTab()

    expect(reload).toHaveBeenCalledTimes(1)
  })

  it('ビルドIDを読めない環境（開発サーバーなど）では、確認しない', async () => {
    mockGetClientBuildId.mockReturnValue(null)
    render(<BuildSkewGuard />)

    await returnToTab()

    expect(mockFetchLatestBuildId).not.toHaveBeenCalled()
    expect(reload).not.toHaveBeenCalled()
  })

  it('最新を取得できなかったときは、読み込み直さない', async () => {
    mockFetchLatestBuildId.mockResolvedValue(null)
    render(<BuildSkewGuard />)

    await returnToTab()

    expect(reload).not.toHaveBeenCalled()
  })

  it('/auth/callback では読み込み直さない（認可コードは一度しか使えない）', async () => {
    setLocation('/auth/callback/')
    mockFetchLatestBuildId.mockResolvedValue('build-NEW')
    render(<BuildSkewGuard />)

    await returnToTab()

    expect(reload).not.toHaveBeenCalled()
  })

  it('直前に読み込み直していたら、繰り返さない', async () => {
    sessionStorage.setItem('build_skew_reloaded_at', String(Date.now()))
    mockFetchLatestBuildId.mockResolvedValue('build-NEW')
    render(<BuildSkewGuard />)

    await returnToTab()

    expect(reload).not.toHaveBeenCalled()
  })

  it('復帰のたびに確認しない（間隔をあける）', async () => {
    render(<BuildSkewGuard />)

    await returnToTab()
    await returnToTab()

    expect(mockFetchLatestBuildId).toHaveBeenCalledTimes(1)
  })

  describe('前面で使用中に古くなったとき', () => {
    const makeLink = (href: string, attrs: Record<string, string> = {}) => {
      const a = document.createElement('a')
      a.href = href
      Object.entries(attrs).forEach(([k, v]) => a.setAttribute(k, v))
      a.textContent = 'link'
      document.body.appendChild(a)
      return a
    }

    const click = (el: Element, init: MouseEventInit = {}) => {
      const event = new MouseEvent('click', { bubbles: true, cancelable: true, ...init })
      el.dispatchEvent(event)
      return event
    }

    const goStale = async () => {
      mockFetchLatestBuildId.mockResolvedValue('build-NEW')
      render(<BuildSkewGuard />)
      await act(async () => {
        jest.advanceTimersByTime(5 * 60 * 1000 + 100)
      })
    }

    it('サイト内リンクを、クライアント遷移ではなくブラウザごとの移動にする', async () => {
      await goStale()
      const link = makeLink('http://localhost/search/')

      const event = click(link)

      expect(event.defaultPrevented).toBe(true)
      expect(assign).toHaveBeenCalledWith('http://localhost/search/')
      // 使用中は、操作の途中で読み込み直さない
      expect(reload).not.toHaveBeenCalled()
    })

    it('外部リンク・新しいタブ・修飾キー付きは、そのままにする', async () => {
      await goStale()

      expect(click(makeLink('https://example.com/')).defaultPrevented).toBe(false)
      expect(click(makeLink('http://localhost/a/', { target: '_blank' })).defaultPrevented).toBe(
        false
      )
      expect(click(makeLink('http://localhost/b/'), { metaKey: true }).defaultPrevented).toBe(false)
      expect(assign).not.toHaveBeenCalled()
    })

    it('古くなっていなければ、リンクをそのままにする', async () => {
      render(<BuildSkewGuard />)
      await act(async () => {
        jest.advanceTimersByTime(5 * 60 * 1000 + 100)
      })

      const event = click(makeLink('http://localhost/search/'))

      expect(event.defaultPrevented).toBe(false)
      expect(assign).not.toHaveBeenCalled()
    })
  })
})
