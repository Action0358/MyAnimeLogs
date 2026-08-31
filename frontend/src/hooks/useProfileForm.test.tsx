import { renderHook, act, waitFor } from '@testing-library/react'
import { useProfileForm } from './useProfileForm'
import { useAuth } from '@/contexts/AuthContext'
import { useRouter } from 'next/navigation'
import { put, uploadFile, ApiClientError } from '@/lib/api/client'

jest.mock('@/contexts/AuthContext')
jest.mock('next/navigation')
// put / uploadFile のみモックし、ApiClientError は実クラスを保持（instanceof 判定のため）
jest.mock('@/lib/api/client', () => {
  const actual = jest.requireActual('@/lib/api/client')
  return { ...actual, put: jest.fn(), uploadFile: jest.fn() }
})

const mockPut = put as jest.Mock
const mockUploadFile = uploadFile as jest.Mock
const mockUpdateUser = jest.fn()
const mockPush = jest.fn()

const baseUser = {
  id: 1,
  name: '初期名',
  email: 'a@b.c',
  description: '初期説明',
  favorite_anime: '初期アニメ',
}

const inputEvent = (value: string) => ({ target: { value } }) as any
const submitEvent = () => ({ preventDefault: jest.fn() }) as any

beforeEach(() => {
  jest.clearAllMocks()
  ;(useRouter as jest.Mock).mockReturnValue({ push: mockPush })
  ;(useAuth as jest.Mock).mockReturnValue({
    user: baseUser,
    updateUser: mockUpdateUser,
  })
})

describe('useProfileForm', () => {
  it('ユーザー情報から初期値を取得し、hasChanges は false', () => {
    const { result } = renderHook(() => useProfileForm())

    expect(result.current.name).toBe('初期名')
    expect(result.current.description).toBe('初期説明')
    expect(result.current.favoriteAnime).toBe('初期アニメ')
    expect(result.current.hasChanges).toBe(false)
  })

  it('値を変更すると hasChanges が true になり、初期値に戻すと false に戻る', () => {
    const { result } = renderHook(() => useProfileForm())

    act(() => result.current.handleNameChange(inputEvent('新しい名前')))
    expect(result.current.name).toBe('新しい名前')
    expect(result.current.hasChanges).toBe(true)

    act(() => result.current.handleNameChange(inputEvent('初期名')))
    expect(result.current.hasChanges).toBe(false)
  })

  it('ユーザー名が空の場合はエラーを表示し、API を呼ばない', async () => {
    const { result } = renderHook(() => useProfileForm())

    act(() => result.current.handleNameChange(inputEvent('   ')))
    await act(async () => {
      await result.current.handleSubmit(submitEvent())
    })

    expect(result.current.error).toBe('ユーザー名を入力してください')
    expect(mockPut).not.toHaveBeenCalled()
  })

  it('ユーザー名が上限超過の場合はエラーを表示し、API を呼ばない', async () => {
    const { result } = renderHook(() => useProfileForm())

    act(() => result.current.handleNameChange(inputEvent('あ'.repeat(51))))
    await act(async () => {
      await result.current.handleSubmit(submitEvent())
    })

    expect(result.current.error).toBe(
      'ユーザー名は50文字以内で入力してください'
    )
    expect(mockPut).not.toHaveBeenCalled()
  })

  it('正常時は trim / null 変換した body で put し、updateUser と /mypage 遷移を行う', async () => {
    mockPut.mockResolvedValue({ user: { ...baseUser, name: '更新名' } })
    const { result } = renderHook(() => useProfileForm())

    act(() => result.current.handleNameChange(inputEvent('  更新名  ')))
    act(() => result.current.handleFavoriteAnimeChange(inputEvent('   ')))

    await act(async () => {
      await result.current.handleSubmit(submitEvent())
    })

    expect(mockPut).toHaveBeenCalledWith('/api/users/me', {
      name: '更新名',
      description: '初期説明',
      favorite_anime: null,
    })
    expect(mockUpdateUser).toHaveBeenCalledWith({ ...baseUser, name: '更新名' })
    expect(mockPush).toHaveBeenCalledWith('/mypage')
  })

  it('ApiClientError 時はそのメッセージを表示し、遷移しない', async () => {
    mockPut.mockRejectedValue(new ApiClientError('サーバーエラー', 500))
    const { result } = renderHook(() => useProfileForm())

    act(() => result.current.handleNameChange(inputEvent('新名')))
    await act(async () => {
      await result.current.handleSubmit(submitEvent())
    })

    expect(result.current.error).toBe('サーバーエラー')
    expect(mockPush).not.toHaveBeenCalled()
  })


  it('変更なしのキャンセルは確認なしで /mypage へ遷移する', () => {
    const { result } = renderHook(() => useProfileForm())

    act(() => result.current.handleCancel())

    expect(mockPush).toHaveBeenCalledWith('/mypage')
  })
})
