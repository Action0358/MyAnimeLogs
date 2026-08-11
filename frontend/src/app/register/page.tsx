'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { useAuth, DEFAULT_REGISTER_ERROR } from '@/contexts/AuthContext';
import { Input } from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';
import { AuthCard } from '@/components/auth/AuthCard';
import { AuthMessages } from '@/components/auth/AuthMessages';

const RegisterPage: React.FC = () => {
  const { register } = useAuth();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [passwordConfirm, setPasswordConfirm] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    // パスワード確認
    if (password !== passwordConfirm) {
      setError('パスワードが一致しません');
      return;
    }

    // パスワード長チェック
    if (password.length < 8) {
      setError('パスワードは8文字以上で入力してください');
      return;
    }

    setIsLoading(true);

    try {
      await register(name, email, password);
    } catch (err) {
      setError(err instanceof Error ? err.message : DEFAULT_REGISTER_ERROR);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <AuthCard
      title="新規登録"
      subtitle="アカウントを作成して、アニメ視聴管理を始めましょう"
      icon="✨"
    >
      <form onSubmit={handleSubmit} className="space-y-5">
        <AuthMessages error={error} />

            <Input
              label="ユーザー名"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="マイアニメログ"
              required
              autoComplete="name"
              helperText="1〜50文字で入力してください"
            />

            <Input
              label="メールアドレス"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="example@email.com"
              required
              autoComplete="email"
            />

            <Input
              label="パスワード"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="8文字以上"
              required
              autoComplete="new-password"
              helperText="8文字以上で入力してください"
            />

            <Input
              label="パスワード（確認）"
              type="password"
              value={passwordConfirm}
              onChange={(e) => setPasswordConfirm(e.target.value)}
              placeholder="パスワードを再入力"
              required
              autoComplete="new-password"
            />

        <Button
          type="submit"
          variant="primary"
          size="lg"
          isLoading={isLoading}
          className="w-full mt-6"
        >
          登録する
        </Button>
      </form>

      <div className="mt-6 pt-6 border-t border-gray-200 text-center">
        <p className="text-sm text-gray-600">
          既にアカウントをお持ちの方は{' '}
          <Link
            href="/login"
            className="text-blue-600 hover:text-blue-700 font-medium transition-colors"
          >
            ログイン
          </Link>
        </p>
      </div>
    </AuthCard>
  );
};

export default RegisterPage;
