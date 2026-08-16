import { isValidAnimeStatus, toAnimeStatus, LOGOUT_REDIRECT_PATHS } from '../constants'
import type { AnimeStatus } from '../constants'

describe('isValidAnimeStatus', () => {
  it('should return true for valid statuses', () => {
    expect(isValidAnimeStatus('airing')).toBe(true)
    expect(isValidAnimeStatus('finished')).toBe(true)
    expect(isValidAnimeStatus('upcoming')).toBe(true)
  })

  it('should return false for invalid statuses', () => {
    expect(isValidAnimeStatus('currently_airing')).toBe(false)
    expect(isValidAnimeStatus('not_yet_aired')).toBe(false)
    expect(isValidAnimeStatus('')).toBe(false)
    expect(isValidAnimeStatus(null)).toBe(false)
    expect(isValidAnimeStatus(undefined)).toBe(false)
    expect(isValidAnimeStatus(123)).toBe(false)
  })
})

describe('toAnimeStatus', () => {
  it('should return AnimeStatus for valid values', () => {
    expect(toAnimeStatus('airing')).toBe('airing')
    expect(toAnimeStatus('finished')).toBe('finished')
    expect(toAnimeStatus('upcoming')).toBe('upcoming')
  })

  it('should return undefined for invalid values', () => {
    expect(toAnimeStatus('currently_airing')).toBeUndefined()
    expect(toAnimeStatus('invalid')).toBeUndefined()
    expect(toAnimeStatus(null)).toBeUndefined()
    expect(toAnimeStatus(undefined)).toBeUndefined()
  })
})

describe('LOGOUT_REDIRECT_PATHS', () => {
  // アカウント削除後にトップへ戻すと削除できたか分からないため、
  // 専用の完了画面へ送る必要がある
  it('アカウント削除は専用の完了画面へ遷移する', () => {
    expect(LOGOUT_REDIRECT_PATHS['account-deleted']).toBe('/account-deleted')
  })

  it('通常のログアウトはトップページへ戻す', () => {
    expect(LOGOUT_REDIRECT_PATHS.manual).toBe('/')
  })

  it('セッション切れはログイン画面へ理由付きで送る', () => {
    expect(LOGOUT_REDIRECT_PATHS.timeout).toContain('/signin?reason=')
    expect(LOGOUT_REDIRECT_PATHS.unauthorized).toContain('/signin?reason=')
  })
})
