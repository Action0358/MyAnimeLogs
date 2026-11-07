/**
 * デプロイをまたいだタブの検知（#1311）。
 *
 * 静的エクスポートの Next.js は、クライアント遷移で `<パス>/index.txt`（RSC ペイロード）を
 * 取得して画面を組み立てる。ペイロードのビルド ID が開いているページの JS と食い違うと
 * （= デプロイ前から開いていたタブ）、遷移をやめて **その .txt の URL へブラウザごと移る**ため、
 * 生テキストが表示される。遷移が失敗する前に食い違いを見つけ、読み込み直すために使う。
 */

/** ペイロードの先頭に入っているビルド ID（`"b":"<id>"`）を取り出す */
export const extractBuildId = (payload: string): string | null =>
  payload.match(/"b":"([^"]+)"/)?.[1] ?? null

/** インラインの script に埋め込まれた RSC データは、引用符が `\"` とエスケープされている */
const EMBEDDED_BUILD_ID = /\\"b\\":\\"([^"\\]+)\\"/

/**
 * 開いているページのビルド ID。ページに埋め込まれた RSC データから読む。
 * Next.js は読み込み後に self.__next_f を空にするため、まず self.__next_f を見て、
 * 無ければ HTML に残っているインラインの script を読む。
 * 読めない環境では null
 */
export const getClientBuildId = (): string | null => {
  const chunks = (self as unknown as { __next_f?: unknown[][] }).__next_f
  if (Array.isArray(chunks)) {
    for (const chunk of chunks) {
      const id = typeof chunk?.[1] === 'string' ? extractBuildId(chunk[1]) : null
      if (id) return id
    }
  }
  for (const script of Array.from(document.scripts)) {
    if (script.src) continue
    const id = script.textContent?.match(EMBEDDED_BUILD_ID)?.[1]
    if (id) return id
  }
  return null
}

/** ビルド ID が出てくる先頭部分だけ読めば足りる。全体を取得しない */
const PEEK_BYTES = 2048

/**
 * 最新のビルド ID。`/index.txt`（トップの RSC ペイロード）の先頭だけを、キャッシュを使わずに読む。
 * 取得できない・読めないときは null（開発サーバーには /index.txt が無い）
 */
export const fetchLatestBuildId = async (): Promise<string | null> => {
  const controller = new AbortController()
  try {
    const res = await fetch('/index.txt', {
      cache: 'no-store',
      headers: { RSC: '1' },
      signal: controller.signal,
    })
    if (!res.ok || !res.body) return null

    const reader = res.body.getReader()
    const decoder = new TextDecoder()
    let text = ''
    while (text.length < PEEK_BYTES) {
      const { done, value } = await reader.read()
      if (done) break
      text += decoder.decode(value, { stream: true })
      const id = extractBuildId(text)
      if (id) {
        controller.abort()
        return id
      }
    }
    controller.abort()
    return extractBuildId(text)
  } catch {
    return null
  }
}
