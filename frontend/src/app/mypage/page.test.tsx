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
      isAuthenticated: false,
    });

    render(<MyPage />);

    expect(mockRouter.push).toHaveBeenCalledWith('/login');
  });

  it('認証済みユーザーの情報を表示する', async () => {
    (useAuth as jest.Mock).mockReturnValue({
      user: mockUser,
      token: 'test-token',
      isLoading: false,
      isAuthenticated: true,
    });

    (global.fetch as jest.Mock).mockImplementation((url) => {
      if (url.includes('/stats')) {
        return Promise.resolve({
          ok: true,
          json: async () => mockStats,
        });
      }
      if (url.includes('/favorites')) {
        return Promise.resolve({
          ok: true,
          json: async () => [],
        });
      }
      return Promise.reject(new Error('Unknown URL'));
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
      isAuthenticated: true,
    });

    (global.fetch as jest.Mock).mockImplementation((url) => {
      if (url.includes('/stats')) {
        return Promise.resolve({
          ok: true,
          json: async () => mockStats,
        });
      }
      if (url.includes('/favorites')) {
        return Promise.resolve({
          ok: true,
          json: async () => [],
        });
      }
      return Promise.reject(new Error('Unknown URL'));
    });

    render(<MyPage />);

    await waitFor(() => {
      expect(screen.getByText('100')).toBeInTheDocument();
      expect(screen.getByText('10')).toBeInTheDocument();
      expect(screen.getByText('15')).toBeInTheDocument();
    });
  });

  it('ローディング中は認証情報確認中を表示する', () => {
    (useAuth as jest.Mock).mockReturnValue({
      user: null,
      token: null,
      isLoading: true,
      isAuthenticated: false,
    });

    render(<MyPage />);

    expect(screen.getByText('認証情報を確認中...')).toBeInTheDocument();
  });

  it('データ取得エラー時にエラーメッセージを表示する', async () => {
    (useAuth as jest.Mock).mockReturnValue({
      user: mockUser,
      token: 'test-token',
      isLoading: false,
      isAuthenticated: true,
    });

    (global.fetch as jest.Mock).mockImplementation(() => {
      return Promise.resolve({
        ok: false,
      });
    });

    render(<MyPage />);

    await waitFor(() => {
      expect(screen.getByText('統計情報の取得に失敗しました')).toBeInTheDocument();
    });
  });

  it('ネットワークエラー時にエラーメッセージを表示する', async () => {
    (useAuth as jest.Mock).mockReturnValue({
      user: mockUser,
      token: 'test-token',
      isLoading: false,
      isAuthenticated: true,
    });

    (global.fetch as jest.Mock).mockRejectedValue(new Error('Network error'));

    render(<MyPage />);

    await waitFor(() => {
      expect(screen.getByText('データの取得に失敗しました')).toBeInTheDocument();
    });
  });
});
