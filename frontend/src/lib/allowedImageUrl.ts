/**
 * CSP の img-src で許可している外部画像のホスト。
 *
 * public/_headers / _headers.staging / _headers.production の img-src と一致させる
 * （API のドメインは環境ごとに違い、アニメ画像には使わないため除く）。
 * ずれはテストで検出する。
 */
export const ALLOWED_IMAGE_HOSTS = [
  '*.anilist.co',
  'cdn.myanimelist.net',
  '*.twimg.com',
  'i.imgur.com',
  'imgur.com',
  'avatars.githubusercontent.com',
  'github.com',
  'raw.githubusercontent.com',
  '*.gravatar.com',
  '*.cloudinary.com',
  'image.tmdb.org',
  'lh3.googleusercontent.com',
] as const

const matchesHost = (hostname: string, pattern: string): boolean =>
  pattern.startsWith('*.')
    ? // CSP のワイルドカードはサブドメインだけに一致し、ドメインそのものには一致しない
      hostname.endsWith(pattern.slice(1))
    : hostname === pattern

/**
 * CSP で読み込みが許可される画像URLかどうか。
 *
 * MAL ID のない作品は image_url に公式サイトの OGP 画像（作品ごとに別ドメイン）が入り、
 * そのまま img に渡すと CSP に止められてコンソールにエラーが出る。
 * 読み込む前に判定し、許可されないURLは使わない。
 */
export const isAllowedImageUrl = (raw: string | null | undefined): boolean => {
  if (!raw) return false
  try {
    const url = new URL(raw)
    if (url.protocol !== 'https:') return false
    const hostname = url.hostname.toLowerCase()
    return ALLOWED_IMAGE_HOSTS.some((pattern) => matchesHost(hostname, pattern))
  } catch {
    return false
  }
}

/** 候補のうち、CSP で許可される最初の画像URLを返す。なければ undefined */
export const pickAllowedImageUrl = (
  ...candidates: Array<string | null | undefined>
): string | undefined => candidates.find(isAllowedImageUrl) ?? undefined
