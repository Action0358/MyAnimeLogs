import { renderHook, act, waitFor } from '@testing-library/react'
import { useIconUpload } from './useIconUpload'
import { uploadFile } from '@/lib/api/client'

jest.mock('@/lib/api/client', () => {
  const actual = jest.requireActual('@/lib/api/client')
  return { ...actual, uploadFile: jest.fn() }
})

const mockUploadFile = uploadFile as jest.Mock
const mockUpdateUser = jest.fn()

beforeEach(() => {
  jest.clearAllMocks()
})

describe('useIconUpload', () => {
  it('5MB を超えるファイルは拒否し、iconFile を設定しない', () => {
    const { result } = renderHook(() =>
      useIconUpload({ updateUser: mockUpdateUser })
    )

    const big = new File([new ArrayBuffer(6 * 1024 * 1024)], 'big.png', {
      type: 'image/png',
    })
    let ret: { error?: string } = {}
    act(() => {
      ret = result.current.selectIcon(big)
    })

    expect(ret.error).toBe('画像サイズは5MB以下にしてください')
    expect(result.current.iconFile).toBeNull()
  })

  it('許可されていない形式は拒否する', () => {
    const { result } = renderHook(() =>
      useIconUpload({ updateUser: mockUpdateUser })
    )

    const txt = new File(['x'], 'a.txt', { type: 'text/plain' })
    let ret: { error?: string } = {}
    act(() => {
      ret = result.current.selectIcon(txt)
    })

    expect(ret.error).toBe('JPEG、PNG、GIF、WebP形式の画像を選択してください')
    expect(result.current.iconFile).toBeNull()
  })

  it('有効な画像は受理し、プレビューを生成する', async () => {
    const { result } = renderHook(() =>
      useIconUpload({ updateUser: mockUpdateUser })
    )

    const png = new File(['data'], 'a.png', { type: 'image/png' })
    let ret: { error?: string } = {}
    act(() => {
      ret = result.current.selectIcon(png)
    })

    expect(ret.error).toBeUndefined()
    expect(result.current.iconFile).toBe(png)
    await waitFor(() => {
      expect(result.current.iconPreview).toMatch(/^data:/)
    })
  })

  it('uploadIcon は uploadFile を呼び、返却ユーザーで updateUser する', async () => {
    mockUploadFile.mockResolvedValue({ user: { id: 42 } })
    const { result } = renderHook(() =>
      useIconUpload({ updateUser: mockUpdateUser })
    )

    const png = new File(['data'], 'a.png', { type: 'image/png' })
    act(() => {
      result.current.selectIcon(png)
    })

    await act(async () => {
      await result.current.uploadIcon()
    })

    expect(mockUploadFile).toHaveBeenCalledWith(
      '/api/users/me/icon',
      expect.any(FormData)
    )
    expect(mockUpdateUser).toHaveBeenCalledWith({ id: 42 })
  })

  it('ファイル未選択で uploadIcon を呼んでも何もしない', async () => {
    const { result } = renderHook(() =>
      useIconUpload({ updateUser: mockUpdateUser })
    )

    await act(async () => {
      await result.current.uploadIcon()
    })

    expect(mockUploadFile).not.toHaveBeenCalled()
    expect(mockUpdateUser).not.toHaveBeenCalled()
  })
})
