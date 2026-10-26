// バックエンドDBから取得するアニメの型
export interface DBAnime {
  id: number
  title: string
  image_url?: string
  season?: string
  status?: string
  released_on?: string
  watchers_count: number
  mal_anime_id?: string
  annict_id?: number
  media?: string
  created_at: string
  updated_at: string
  cover_image_large?: string
  cover_image_medium?: string
  genres?: string[]
}

export interface DBAnimeResponse {
  animes: DBAnime[]
  total: number
}

import { get, ApiClientError } from './client'

// バックエンドDBからアニメを取得するAPI
export const animeApi = {
  async getCurrentSeasonAnimes(limit?: number, random?: boolean): Promise<DBAnimeResponse> {
    try {
      const query = new URLSearchParams()
      if (limit) query.set('limit', String(limit))
      if (random) query.set('random', 'true')
      const params = query.toString() ? `?${query.toString()}` : ''
      return await get<DBAnimeResponse>(`/api/animes/current${params}`)
    } catch (error) {
      if (error instanceof ApiClientError) {
        throw new Error('今期アニメの取得に失敗しました')
      }
      throw error
    }
  },

  async getPopularAnimes(limit?: number, random?: boolean): Promise<DBAnimeResponse> {
    try {
      const query = new URLSearchParams()
      if (limit) query.set('limit', String(limit))
      if (random) query.set('random', 'true')
      const params = query.toString() ? `?${query.toString()}` : ''
      return await get<DBAnimeResponse>(`/api/animes/popular${params}`)
    } catch (error) {
      if (error instanceof ApiClientError) {
        throw new Error('人気アニメの取得に失敗しました')
      }
      throw error
    }
  },
}
