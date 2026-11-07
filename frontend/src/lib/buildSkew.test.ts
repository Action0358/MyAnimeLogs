import { TextDecoder, TextEncoder } from 'util'
import { ReadableStream } from 'stream/web'
import { extractBuildId, fetchLatestBuildId, getClientBuildId } from './buildSkew'

// jsdom には無いため、Node のものを使う
Object.assign(globalThis, { TextDecoder, TextEncoder, ReadableStream })

const streamOf = (...parts: string[]) => {
  const encoder = new TextEncoder()
  return new ReadableStream<Uint8Array>({
    start(controller) {
      parts.forEach((part) => controller.enqueue(encoder.encode(part)))
      controller.close()
    },
  })
}

const respond = (body: ReadableStream<Uint8Array> | null, ok = true) =>
  ({ ok, body }) as unknown as Response

describe('extractBuildId', () => {
  it('ペイロードの "b" からビルドIDを取り出す', () => {
    expect(extractBuildId('0:{"P":null,"b":"abc_123-XYZ","p":""}')).toBe('abc_123-XYZ')
  })

  it('無ければ null', () => {
    expect(extractBuildId('1:"$Sreact.fragment"')).toBeNull()
  })
})

describe('getClientBuildId', () => {
  afterEach(() => {
    delete (self as unknown as { __next_f?: unknown }).__next_f
  })

  it('self.__next_f の RSC データから読む', () => {
    ;(self as unknown as { __next_f: unknown[][] }).__next_f = [
      [0],
      [1, '1:"x"\n'],
      [1, '0:{"P":null,"b":"build-A","p":""}'],
    ]

    expect(getClientBuildId()).toBe('build-A')
  })

  it('読めない環境（開発サーバーなど）では null', () => {
    expect(getClientBuildId()).toBeNull()
  })
})

describe('fetchLatestBuildId', () => {
  const originalFetch = global.fetch

  afterEach(() => {
    global.fetch = originalFetch
  })

  it('/index.txt をキャッシュなしで取得し、先頭のビルドIDを返す', async () => {
    global.fetch = jest
      .fn()
      .mockResolvedValue(respond(streamOf('0:{"P":null,"b":"build-B",', '"p":""}'))) as typeof fetch

    await expect(fetchLatestBuildId()).resolves.toBe('build-B')
    expect(global.fetch).toHaveBeenCalledWith(
      '/index.txt',
      expect.objectContaining({ cache: 'no-store', headers: { RSC: '1' } })
    )
  })

  it('ビルドIDが分割されて届いても読める', async () => {
    global.fetch = jest
      .fn()
      .mockResolvedValue(respond(streamOf('0:{"P":null,"b":"bui', 'ld-C","p":""}'))) as typeof fetch

    await expect(fetchLatestBuildId()).resolves.toBe('build-C')
  })

  it('取得に失敗したら null（開発サーバーの 404 など）', async () => {
    global.fetch = jest.fn().mockResolvedValue(respond(null, false)) as typeof fetch
    await expect(fetchLatestBuildId()).resolves.toBeNull()

    global.fetch = jest.fn().mockRejectedValue(new Error('network')) as typeof fetch
    await expect(fetchLatestBuildId()).resolves.toBeNull()
  })

  it('ビルドIDが見つからなければ null', async () => {
    global.fetch = jest
      .fn()
      .mockResolvedValue(respond(streamOf('<html>not payload</html>'))) as typeof fetch

    await expect(fetchLatestBuildId()).resolves.toBeNull()
  })
})
