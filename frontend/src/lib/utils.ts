import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

// フォームバリデーション用ユーティリティ
export const validateEmail = (email: string): boolean => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
};

export const validatePassword = (password: string): boolean => {
  return password.length >= 8;
};

// エラーメッセージ取得
export const getErrorMessage = (error: unknown): string => {
  if (error instanceof Error) return error.message;
  if (typeof error === 'string') return error;
  return 'An unexpected error occurred';
};

// ローカルストレージ操作
export const storage = {
  /**
   * localStorageから値を取得（エラーをスローする）
   * @throws {Error} パースエラーや不正なデータの場合
   */
  getUnsafe: <T = any>(key: string): T | null => {
    if (typeof window === 'undefined') return null;
    const item = localStorage.getItem(key);
    if (!item) return null;
    return JSON.parse(item) as T; // エラーをスローする可能性がある
  },

  /**
   * localStorageから値を取得（エラーを捕捉してnullを返す）
   */
  get: <T = any>(key: string): T | null => {
    if (typeof window === 'undefined') return null;
    try {
      return storage.getUnsafe<T>(key);
    } catch (error) {
      console.error(`Failed to get item from localStorage with key "${key}":`, error);
      return null;
    }
  },

  set: (key: string, value: any) => {
    if (typeof window === 'undefined') return;
    try {
      localStorage.setItem(key, JSON.stringify(value));
    } catch (error) {
      console.error(`Failed to set item in localStorage with key "${key}":`, error);
    }
  },

  remove: (key: string) => {
    if (typeof window === 'undefined') return;
    try {
      localStorage.removeItem(key);
    } catch (error) {
      console.error(`Failed to remove item from localStorage with key "${key}":`, error);
    }
  },
};

/**
 * アイコンURLを構築する
 * @param iconPath - アイコンのパス（例: "/uploads/icons/user123.png"）
 * @returns 完全なURL（例: "http://localhost:8080/uploads/icons/user123.png"）
 */
export const buildIconUrl = (iconPath: string | null | undefined): string | null => {
  // null, undefined, 空文字、空白のみの文字列をnullとして扱う
  if (!iconPath || !iconPath.trim()) return null;
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
  return `${apiUrl}${iconPath}`;
};
