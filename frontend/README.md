## 概要

- Next.js + TypeScript
- App Router (Next.js 13+)
- Tailwind CSS
- Storybook
- MSW (Mock Service Worker)
- Orval (OpenAPI code generation)
- Dockerによる環境構築
- ホットリロード対応

## 前提条件

- Docker
- Docker Compose
- Node.js 18+
- Next.js 14
- React 19
- TypeScript 5.0+
- Tailwind CSS 3.4+

## インストール手順

### 1. Dockerコンテナの起動

```bash
docker compose build
docker compose up -d
```

### 2. アプリケーションへのアクセス

- Next.jsアプリ: http://localhost:3000/
- Storybook: http://localhost:6006/ (npm run storybook実行時)

## 開発環境設定

### ホットリロード

ホットリロードが有効になっています。ソースコードを変更すると自動的にブラウザが更新されます。

### TypeScript

TypeScriptの設定は `tsconfig.json` で管理されています。

### App Router

Next.js 13+のApp Routerを使用したファイルベースルーティングです。

### API通信

- OpenAPIスキーマから自動生成されたクライアントコード
- Orvalを使用したコード生成
- MSWによるモック対応

### スタイリング

Tailwind CSSを使用したユーティリティファーストCSS

### コンポーネント開発

Storybookを使用したコンポーネント駆動開発

## ファイル構造

```
frontend/
├── src/                         # ソースコード
│   ├── app/                     # App Router (Next.js 13+)
│   │   ├── anime/[id]/         # 動的ルート（アニメ詳細）
│   │   ├── auth/signin/        # 認証ページ
│   │   ├── profile/            # プロフィールページ
│   │   ├── watchlist/          # ウォッチリストページ
│   │   ├── globals.css         # グローバルスタイル
│   │   ├── layout.tsx          # ルートレイアウト
│   │   └── page.tsx            # ホームページ
│   ├── features/               # 機能別コンポーネント
│   │   └── anime-trend/        # トレンドアニメ機能
│   │       ├── components/     # コンポーネント
│   │       ├── fetchers/       # データフェッチャー
│   │       └── mocks/          # モックデータ
│   ├── lib/                    # ライブラリとユーティリティ
│   │   ├── api/                # API関連
│   │   │   ├── generated/      # 自動生成されたクライアント
│   │   │   ├── fetchers/       # データフェッチャー
│   │   │   └── schemas/        # APIスキーマ
│   │   ├── mocks/              # モック設定
│   │   └── utils.ts            # ユーティリティ関数
│   ├── hooks/                  # カスタムフック
│   ├── stores/                 # 状態管理
│   ├── testing/                # テスト設定
│   │   └── mocks/              # テスト用モック
│   ├── types/                  # TypeScript型定義
│   └── utils/                  # ユーティリティ
├── .storybook/                 # Storybook設定
├── public/                     # 静的ファイル
├── .env                        # 環境変数設定
├── Dockerfile                  # Docker設定
├── orval.config.js             # API生成設定
├── tailwind.config.js          # Tailwind CSS設定
└── package.json                # プロジェクト設定
```

## その他の設定

### 環境変数

Next.jsでは `NEXT_PUBLIC_` で始まる変数がクライアントサイドで使用できます。
必要な場合、`.env.local`、`.env.development`、`.env.production` を作成してください。

### API生成

OpenAPIスキーマからクライアントコードを生成：
```bash
docker compose exec frontend npm run generate:api
```

### Storybook

コンポーネントの開発・テスト：
```bash
docker compose exec frontend npm run storybook
```

### ビルド

```bash
docker compose exec frontend npm run build
```

### テスト

```bash
# 単体テスト
docker compose exec frontend npm test

# テスト（ウォッチモード）
docker compose exec frontend npm run test:watch

# カバレッジ
docker compose exec frontend npm run test:coverage
```

### リント・フォーマット

```bash
# ESLint
docker compose exec frontend npm run lint

# Prettier
docker compose exec frontend npm run format
```
