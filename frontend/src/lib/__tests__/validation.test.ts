import { validateEmail, validatePassword, getErrorMessage, cn } from '../utils'

describe('validateEmail', () => {
  it('should return true for valid emails', () => {
    expect(validateEmail('test@example.com')).toBe(true)
    expect(validateEmail('user.name@domain.co.jp')).toBe(true)
    expect(validateEmail('a@b.c')).toBe(true)
  })

  it('should return false for invalid emails', () => {
    expect(validateEmail('')).toBe(false)
    expect(validateEmail('test')).toBe(false)
    expect(validateEmail('test@')).toBe(false)
    expect(validateEmail('@example.com')).toBe(false)
    expect(validateEmail('test @example.com')).toBe(false)
  })
})

describe('validatePassword', () => {
  it('should return true for passwords with 8+ characters', () => {
    expect(validatePassword('12345678')).toBe(true)
    expect(validatePassword('abcdefgh')).toBe(true)
    expect(validatePassword('a very long password')).toBe(true)
  })

  it('should return false for passwords shorter than 8 characters', () => {
    expect(validatePassword('')).toBe(false)
    expect(validatePassword('1234567')).toBe(false)
    expect(validatePassword('abc')).toBe(false)
  })
})

describe('getErrorMessage', () => {
  it('should return message from Error object', () => {
    expect(getErrorMessage(new Error('test error'))).toBe('test error')
  })

  it('should return string as-is', () => {
    expect(getErrorMessage('string error')).toBe('string error')
  })

  it('should return default message for unknown types', () => {
    expect(getErrorMessage(123)).toBe('An unexpected error occurred')
    expect(getErrorMessage(null)).toBe('An unexpected error occurred')
    expect(getErrorMessage(undefined)).toBe('An unexpected error occurred')
    expect(getErrorMessage({})).toBe('An unexpected error occurred')
  })
})

describe('cn', () => {
  it('should merge class names', () => {
    expect(cn('foo', 'bar')).toBe('foo bar')
  })

  it('should handle conditional classes', () => {
    expect(cn('foo', false && 'bar', 'baz')).toBe('foo baz')
  })

  it('should merge tailwind classes', () => {
    expect(cn('px-2', 'px-4')).toBe('px-4')
    expect(cn('text-red-500', 'text-blue-500')).toBe('text-blue-500')
  })
})
