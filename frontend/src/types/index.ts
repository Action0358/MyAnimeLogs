// 共通型定義

export interface User {
  id: number
  name: string
  email: string
  description: string
  icon: string | null
}

export interface LoginCredentials {
  email: string
  password: string
}

export interface RegisterData {
  name: string
  email: string
  password: string
}

export interface UpdateProfileData {
  name?: string
  description?: string
  icon?: string
}

export interface LoginResponse {
  access_token: string
  token_type: string
  user: User
}

export interface RegisterResponse {
  message: string
  user: User
}

export interface ApiError {
  error: string
  message: string
  code?: string
}

// UI関連の型定義
export type ButtonVariant = 'primary' | 'secondary' | 'outline' | 'ghost' | 'danger'
export type ButtonSize = 'xs' | 'sm' | 'md' | 'lg' | 'xl'
