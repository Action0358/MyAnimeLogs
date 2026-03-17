export interface AnnictWorkImagesFacebook {
  og_image_url?: string;
}

export interface AnnictWorkImagesTwitter {
  mini_avatar_url?: string;
  normal_avatar_url?: string;
  bigger_avatar_url?: string;
  original_avatar_url?: string;
  image_url?: string;
}

export interface AnnictWorkImages {
  recommended_url?: string;
  facebook?: AnnictWorkImagesFacebook;
  twitter?: AnnictWorkImagesTwitter;
}

export interface AnnictWork {
  id: number;
  title: string;
  title_kana?: string;
  title_en?: string;
  media: string;
  media_text: string;
  season_name: string;
  season_year: number;
  released_on?: string;
  released_on_about?: string;
  official_site_url?: string;
  wikipedia_url?: string;
  twitter_username?: string;
  twitter_hashtag?: string;
  mal_anime_id?: string;
  images?: AnnictWorkImages;
  episodes_count?: number;
  watchers_count: number;
  reviews_count: number;
  no_episodes?: boolean;
  // Anilist APIから取得した画像（優先的に使用）
  cover_image_large?: string;
  cover_image_medium?: string;
}

export interface AnnictWorkResponse {
  works: AnnictWork[];
  total_count: number;
  next_page?: number;
  prev_page?: number;
}

// バックエンドDBから取得するアニメの型
export interface DBAnime {
  id: number;
  title: string;
  image_url?: string;
  season?: string;
  status?: string;
  released_on?: string;
  watchers_count: number;
  mal_anime_id?: string;
  annict_id?: number;
  media?: string;
  created_at: string;
  updated_at: string;
  cover_image_large?: string;
  cover_image_medium?: string;
}

export interface DBAnimeResponse {
  animes: DBAnime[];
  total: number;
}

import { get, ApiClientError } from './client'

export const annictApi = {
  async getCurrentSeasonAnime(): Promise<AnnictWorkResponse> {
    try {
      return await get<AnnictWorkResponse>('/api/annict/works')
    } catch (error) {
      if (error instanceof ApiClientError) {
        throw new Error('今期アニメの取得に失敗しました')
      }
      throw error
    }
  },

  async getPopularAnime(): Promise<AnnictWorkResponse> {
    try {
      return await get<AnnictWorkResponse>('/api/annict/popular-works')
    } catch (error) {
      if (error instanceof ApiClientError) {
        throw new Error('人気アニメの取得に失敗しました')
      }
      throw error
    }
  },

  async getAnimesBySeason(season: string): Promise<AnnictWorkResponse> {
    try {
      return await get<AnnictWorkResponse>(
        `/api/annict/works?season=${encodeURIComponent(season)}`
      )
    } catch (error) {
      if (error instanceof ApiClientError) {
        throw new Error('シーズンアニメの取得に失敗しました')
      }
      throw error
    }
  },
}

// バックエンドDBからアニメを取得するAPI
export const animeApi = {
  async getCurrentSeasonAnimes(limit?: number): Promise<DBAnimeResponse> {
    try {
      const params = limit ? `?limit=${limit}` : ''
      return await get<DBAnimeResponse>(`/api/animes/current${params}`)
    } catch (error) {
      if (error instanceof ApiClientError) {
        throw new Error('今期アニメの取得に失敗しました')
      }
      throw error
    }
  },

  async getPopularAnimes(limit?: number): Promise<DBAnimeResponse> {
    try {
      const params = limit ? `?limit=${limit}` : ''
      return await get<DBAnimeResponse>(`/api/animes/popular${params}`)
    } catch (error) {
      if (error instanceof ApiClientError) {
        throw new Error('人気アニメの取得に失敗しました')
      }
      throw error
    }
  },
}