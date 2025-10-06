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
  // Anilist API から取得した画像（優先的に使用）
  cover_image_large?: string;
  cover_image_medium?: string;
}

export interface AnnictWorkResponse {
  works: AnnictWork[];
  total_count: number;
  next_page?: number;
  prev_page?: number;
}

// API_BASE_URLから末尾の/apiを安全に削除
const getApiBaseUrl = () => {
  const baseUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
  // 末尾の/apiを削除（重複を防ぐため）
  return baseUrl.replace(/\/api\/?$/, '');
};

const API_BASE_URL = getApiBaseUrl();

export const annictApi = {
  async getCurrentSeasonAnime(): Promise<AnnictWorkResponse> {
    const response = await fetch(`${API_BASE_URL}/api/annict/works`);
    if (!response.ok) {
      throw new Error('今期アニメの取得に失敗しました');
    }
    return response.json();
  },

  async getPopularAnime(): Promise<AnnictWorkResponse> {
    const response = await fetch(`${API_BASE_URL}/api/annict/popular-works`);
    if (!response.ok) {
      throw new Error('人気アニメの取得に失敗しました');
    }
    return response.json();
  },

  async getAnimesBySeason(season: string): Promise<AnnictWorkResponse> {
    const response = await fetch(`${API_BASE_URL}/api/annict/works?season=${encodeURIComponent(season)}`);
    if (!response.ok) {
      throw new Error('シーズンアニメの取得に失敗しました');
    }
    return response.json();
  }
};