import { validateEmail, validatePassword, getErrorMessage, cn, storage } from '../utils'

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

describe('storage', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  describe('set and get', () => {
    it('should store and retrieve a string', () => {
      storage.set('key', 'value')
      expect(storage.get('key')).toBe('value')
    })

    it('should store and retrieve a number', () => {
      storage.set('num', 42)
      expect(storage.get('num')).toBe(42)
    })

    it('should store and retrieve an object', () => {
      const obj = { a: 1, b: 'two' }
      storage.set('obj', obj)
      expect(storage.get('obj')).toEqual(obj)
    })

    it('should return null for non-existent key', () => {
      expect(storage.get('missing')).toBeNull()
    })
  })

  describe('getString', () => {
    it('should return string value', () => {
      storage.set('str', 'hello')
      expect(storage.getString('str')).toBe('hello')
    })

    it('should return null for non-string value', () => {
      storage.set('num', 123)
      expect(storage.getString('num')).toBeNull()
    })
  })

  describe('getNumber', () => {
    it('should return number value', () => {
      storage.set('num', 99)
      expect(storage.getNumber('num')).toBe(99)
    })

    it('should return null for non-number value', () => {
      storage.set('str', 'hello')
      expect(storage.getNumber('str')).toBeNull()
    })
  })

  describe('getObject', () => {
    it('should return object when validator passes', () => {
      const obj = { name: 'test' }
      storage.set('obj', obj)
      const isValid = (v: unknown): v is { name: string } =>
        typeof v === 'object' && v !== null && 'name' in v
      expect(storage.getObject('obj', isValid)).toEqual(obj)
    })

    it('should return null when validator fails', () => {
      storage.set('num', 42)
      const isValid = (v: unknown): v is { name: string } =>
        typeof v === 'object' && v !== null && 'name' in v
      expect(storage.getObject('num', isValid)).toBeNull()
    })
  })

  describe('remove', () => {
    it('should remove a stored item', () => {
      storage.set('key', 'value')
      storage.remove('key')
      expect(storage.get('key')).toBeNull()
    })
  })
})

