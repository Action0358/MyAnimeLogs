import { PRESET_AVATAR_IDS, isPresetAvatarID, resolveAvatarUrl } from '../avatars'

describe('avatars', () => {
  it('許可リストのIDだけを受け付ける', () => {
    expect(isPresetAvatarID('preset-01')).toBe(true)
    expect(isPresetAvatarID('preset-99')).toBe(false)
    expect(isPresetAvatarID(null)).toBe(false)
    expect(isPresetAvatarID(undefined)).toBe(false)
  })

  it('IDを静的アセットのパスへ解決する', () => {
    expect(resolveAvatarUrl('preset-04')).toBe('/avatars/preset-04.svg')
  })

  // 移行前のアップロードURLが残っていても壊れた画像を出さないための保険
  it('未設定や未知の値は null を返す', () => {
    expect(resolveAvatarUrl(null)).toBeNull()
    expect(resolveAvatarUrl('')).toBeNull()
    expect(resolveAvatarUrl('/uploads/icons/user_14_1786463713.jpeg')).toBeNull()
    expect(resolveAvatarUrl('https://example.com/evil.png')).toBeNull()
  })

  it('プリセットのIDと実ファイル名の規則が一致する', () => {
    PRESET_AVATAR_IDS.forEach((id) => {
      expect(resolveAvatarUrl(id)).toBe(`/avatars/${id}.svg`)
    })
  })
})
