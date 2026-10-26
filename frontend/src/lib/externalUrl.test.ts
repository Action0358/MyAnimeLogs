import { toSafeExternalUrl } from './externalUrl'

describe('toSafeExternalUrl', () => {
  it.each([
    ['https://example.com/anime', 'https://example.com/anime'],
    ['http://example.com/', 'http://example.com/'],
    ['  https://example.com/  ', 'https://example.com/'],
    ['HTTPS://Example.com/', 'https://example.com/'],
  ])('http/https の %s は URL として返す', (raw, expected) => {
    expect(toSafeExternalUrl(raw)?.href).toBe(expected)
  })

  it.each([
    'javascript:alert(document.cookie)',
    'JavaScript:alert(1)',
    '  javascript:alert(1)',
    'java\tscript:alert(1)',
    'data:text/html,<script>alert(1)</script>',
    'vbscript:msgbox(1)',
    '//evil.example',
    '/anime/1',
    'not a url',
    '',
  ])('%s は null を返す', (raw) => {
    expect(toSafeExternalUrl(raw)).toBeNull()
  })

  it('null と undefined は null を返す', () => {
    expect(toSafeExternalUrl(null)).toBeNull()
    expect(toSafeExternalUrl(undefined)).toBeNull()
  })
})
