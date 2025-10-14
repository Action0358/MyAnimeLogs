import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import MyPage from './page';
import { useAuth } from '@/contexts/AuthContext';
import { useRouter } from 'next/navigation';

// Mock dependencies
jest.mock('@/contexts/AuthContext');
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
}));

// Mock fetch
global.fetch = jest.fn();

describe('MyPage', () => {
  const mockRouter = {
    push: jest.fn(),
  };

  const mockUser = {
    id: 1,
    name: 'テストユーザー',
    email: 'test@example.com',
    description: 'テスト説明',
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

    render(<MyPage />);

    expect(mockRouter.push).toHaveBeenCalledWith('/login');
  });

  it('認証済みユーザーの情報を表示する', async () => {
    (useAuth as jest.Mock).mockReturnValue({
      user: mockUser,
      token: 'test-token',
      isLoading: false,
    });

    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => mockStats,
    });

    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    });

    render(<MyPage />);

    await waitFor(() => {
      expect(screen.getByText('テストユーザー')).toBeInTheDocument();
      expect(screen.getByText('test@example.com')).toBeInTheDocument();
    });
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

    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    });

    render(<MyPage />);

    await waitFor(() => {
      expect(screen.getByText('100')).toBeInTheDocument(); // totalEpisodesWatched
      expect(screen.getByText('10')).toBeInTheDocument(); // totalAnimesCompleted
      expect(screen.getByText('15')).toBeInTheDocument(); // thisMonthEpisodes
    });
  });

  it('ローディング中は読み込み表示を出す', () => {
    (useAuth as jest.Mock).mockReturnValue({
      user: mockUser,
      token: 'test-token',
      isLoading: true,
    });

    render(<MyPage />);

    expect(screen.getByText('読み込み中...')).toBeInTheDocument();
  });
});
