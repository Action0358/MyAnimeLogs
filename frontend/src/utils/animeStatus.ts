export interface AnimeStatusConfig {
  label: string;
  bgColor: string;
  textColor: string;
}

export const ANIME_STATUS_CONFIG: Record<string, AnimeStatusConfig> = {
  watching: {
    label: '視聴中',
    bgColor: 'bg-blue-100',
    textColor: 'text-blue-700',
  },
  completed: {
    label: '完了',
    bgColor: 'bg-green-100',
    textColor: 'text-green-700',
  },
  planned: {
    label: '予定',
    bgColor: 'bg-gray-100',
    textColor: 'text-gray-700',
  },
  dropped: {
    label: '中断',
    bgColor: 'bg-orange-100',
    textColor: 'text-orange-700',
  },
};

export const getAnimeStatusConfig = (status: string): AnimeStatusConfig => {
  return ANIME_STATUS_CONFIG[status] || {
    label: status,
    bgColor: 'bg-gray-100',
    textColor: 'text-gray-700',
  };
};
