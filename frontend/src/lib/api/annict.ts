interface AnnictWorkImages {
  recommended_url?: string;
}

export interface AnnictWork {
  id: number;
  title: string;
  title_kana?: string;
  media: string;
  media_text: string;
  season_name: string;
  season_year: number;
  watchers_count: number;
  reviews_count: number;
  no_episodes?: boolean;
  images?: AnnictWorkImages;
  official_site_url?: string;
}

export interface AnnictWorkResponse {
  works: AnnictWork[];
  total_count: number;
  next_page?: number;
  prev_page?: number;
}

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

export const annictApi = {
  async getCurrentSeasonAnime(): Promise<AnnictWorkResponse> {
    const response = await fetch(`${API_BASE_URL}/annict/works`);
    if (!response.ok) {
      throw new Error('今期アニメの取得に失敗しました');
    }
    return response.json();
  },

  async getPopularAnime(): Promise<AnnictWorkResponse> {
    const response = await fetch(`${API_BASE_URL}/annict/popular-works`);
    if (!response.ok) {
      throw new Error('人気アニメの取得に失敗しました');
    }
    return response.json();
  },

  async getAnimesBySeason(season: string): Promise<AnnictWorkResponse> {
    const response = await fetch(`${API_BASE_URL}/annict/works?season=${encodeURIComponent(season)}`);
    if (!response.ok) {
      throw new Error('シーズンアニメの取得に失敗しました');
    }
    return response.json();
  }
};