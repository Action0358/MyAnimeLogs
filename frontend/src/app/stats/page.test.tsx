import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import StatsPage from './page';
import { useAuth } from '@/contexts/AuthContext';
import { useRouter } from 'next/navigation';

// Mock dependencies
jest.mock('@/contexts/AuthContext');
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
}));

// Mock fetch
global.fetch = jest.fn();

describe('StatsPage', () => {
  const mockRouter = {
    push: jest.fn(),
  };

  const mockUser = {
    id: 1,
    name: 'テストユーザー',
    email: 'test@example.com',
  };

  const mockStats = {
    totalEpisodesWatched: 100,
    totalAnimesCompleted: 10,
    totalWatchTimeHours: 40.0,
    favoriteGenres: ['アクション', 'コメディ', 'ドラマ'],
    thisMonthEpisodes: 15,
  };

  beforeEach(() => {
    jest.clearAllMocks();
    (useRouter as jest.Mock).mockReturnValue(mockRouter);
  });

  it('未認証時にログインページにリダイレクトする', () => {
    (useAuth as jest.Mock).mockReturnValue({
      user: null,
      token: null,
      isLoading: false,
    });

    render(<StatsPage />);

    expect(mockRouter.push).toHaveBeenCalledWith('/login');
  });

  it('統計情報を正しく表示する', async () => {
    (useAuth as jest.Mock).mockReturnValue({
      user: mockUser,
      token: 'test-token',
      isLoading: false,
    });

    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => mockStats,
    });

    render(<StatsPage />);

    await waitFor(() => {
      expect(screen.getByText('視聴統計')).toBeInTheDocument();
      expect(screen.getByText('100')).toBeInTheDocument();
      expect(screen.getByText('10')).toBeInTheDocument();
    });
  });

  it('お気に入りジャンルを表示する', async () => {
    (useAuth as jest.Mock).mockReturnValue({
      user: mockUser,
      token: 'test-token',
      isLoading: false,
    });

    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => mockStats,
    });

    render(<StatsPage />);

    await waitFor(() => {
      expect(screen.getByText('アクション')).toBeInTheDocument();
      expect(screen.getByText('コメディ')).toBeInTheDocument();
      expect(screen.getByText('ドラマ')).toBeInTheDocument();
    });
  });

  it('エラー時にエラーメッセージを表示する', async () => {
    (useAuth as jest.Mock).mockReturnValue({
      user: mockUser,
      token: 'test-token',
      isLoading: false,
    });

    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: false,
    });

    render(<StatsPage />);

    await waitFor(() => {
      expect(screen.getByText('統計情報の取得に失敗しました')).toBeInTheDocument();
    });
  });
});
