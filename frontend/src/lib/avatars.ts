/**
 * プリセットアバター。
 *
 * ユーザーによる画像アップロードは #763 で廃止した。任意の画像を持ち込めると
 * 第三者の著作物がアップロードされうること、保存先の永続ボリュームが
 * Fly.io のマシンをホストに固定してデプロイを妨げること（#760）が理由。
 *
 * DB の users.icon にはこのIDを格納する。URLではなくIDにすることで、
 * バックエンド側で許可リスト検証ができ、任意のURLを混入させられない。
 *
 * 【重要】このリストは backend/internal/models/user.go の PresetAvatarIDs と
 * 同期させること。片方だけ増やすと、選べるのに保存できない（またはその逆）
 * という不整合が起きる。
 */
export const PRESET_AVATAR_IDS = [
  'preset-01',
  'preset-02',
  'preset-03',
  'preset-04',
  'preset-05',
  'preset-06',
  'preset-07',
  'preset-08',
] as const

export type PresetAvatarID = (typeof PRESET_AVATAR_IDS)[number]

export const isPresetAvatarID = (value: unknown): value is PresetAvatarID =>
  typeof value === 'string' && (PRESET_AVATAR_IDS as readonly string[]).includes(value)

/**
 * アバターIDを画像パスへ解決する。
 *
 * 画像はフロントエンドの静的アセットとして配信するため、バックエンドの
 * URLは前置しない（従来のアップロード画像は API サーバーが配信していた）。
 *
 * 未設定・未知のIDの場合は null を返し、呼び出し側でプレースホルダーを出す。
 * 移行前のアップロードURLが残っていても、ここで弾かれて壊れた画像にならない。
 */
export const resolveAvatarUrl = (iconID: string | null | undefined): string | null => {
  if (!isPresetAvatarID(iconID)) return null
  return `/avatars/${iconID}.svg`
}
