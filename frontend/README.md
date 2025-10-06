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

### 1. 依存関係のインストール

```bash
# 依存関係のインストール
npm install
# または
yarn install
```

### 2. 環境変数の設定

```bash
# 環境変数のコピー
cp env/local.example .env.local
# 必要に応じて .env.local を編集
```

### 3. Dockerコンテナの起動

```bash
# ビルドと起動
docker compose build
docker compose up -d
```

### 4. アプリケーションへのアクセス

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
├── .storybook/                 # Storybook設定
│   ├── main.ts                 # メイン設定
│   └── preview.ts              # プレビュー設定
├── env/                        # 環境変数テンプレート
│   ├── local.example           # ローカル開発用
│   ├── production.example      # 本番環境用
│   └── test.example            # テスト用
├── public/                     # 静的ファイル
│   ├── favicon.ico
│   ├── logo192.png
│   ├── logo512.png
│   ├── manifest.json
│   ├── mockServiceWorker.js
│   └── robots.txt
├── src/                        # ソースコード
│   ├── app/                    # App Router (Next.js 13+)
│   │   ├── anime/              # アニメ関連
│   │   │   └── [id]/          # 動的ルート（アニメ詳細）
│   │   ├── auth/               # 認証関連
│   │   │   └── signin/        # サインインページ
│   │   ├── profile/           # プロフィールページ
│   │   ├── watchlist/         # ウォッチリストページ
│   │   ├── globals.css        # グローバルスタイル
│   │   ├── layout.tsx         # ルートレイアウト
│   │   └── page.tsx           # ホームページ
│   ├── assets/                # アセット（画像、アイコンなど）
│   │   └── icons/             # アイコン
│   ├── features/              # 機能別コンポーネント
│   │   └── anime-trend/       # トレンドアニメ機能
│   │       ├── components/    # コンポーネント
│   │       │   ├── anime-trend-card/     # アニメカード
│   │       │   └── anime-trend-card-list/ # アニメカードリスト
│   │       ├── fetchers/      # データフェッチャー
│   │       │   └── get-trending-anime/   # トレンドアニメ取得
│   │       └── mocks/         # モックデータ
│   │           └── get-trending-anime-mock/
│   ├── lib/                   # ライブラリとユーティリティ
│   │   ├── api/               # API関連
│   │   │   ├── fetchers/      # データフェッチャー
│   │   │   │   ├── anime/     # アニメ関連
│   │   │   │   ├── authentication/ # 認証関連
│   │   │   │   ├── profile/   # プロフィール関連
│   │   │   │   └── watchlist/ # ウォッチリスト関連
│   │   │   ├── generated/     # 自動生成されたクライアント
│   │   │   └── schemas/       # APIスキーマ
│   │   ├── mocks/             # モック設定
│   │   ├── client-generate/   # クライアント生成設定
│   │   └── utils.ts           # ユーティリティ関数
│   ├── stores/                # 状態管理
│   ├── testing/               # テスト設定
│   │   └── mocks/             # テスト用モック
│   ├── types/                 # TypeScript型定義
│   │   └── result/            # 結果型定義
│   └── utils/                 # ユーティリティ
├── .dockerignore              # Docker無視設定
├── .env.local                 # ローカル環境変数
├── .eslintrc.json             # ESLint設定
├── .gitignore                 # Git無視設定
├── .prettierrc.json           # Prettier設定
├── Dockerfile                 # Docker設定
├── next.config.js             # Next.js設定
├── orval.config.js            # API生成設定
├── package-lock.json          # 依存関係ロック
├── package.json               # プロジェクト設定
├── postcss.config.js          # PostCSS設定
├── tailwind.config.js         # Tailwind CSS設定
└── tsconfig.json              # TypeScript設定
```

## その他の設定

### 環境変数

Next.jsでは `NEXT_PUBLIC_` で始まる変数がクライアントサイドで使用できます。
必要な場合、`.env.local`、`.env.development`、`.env.production` を作成してください。

### APIクライアント生成

OpenAPIスキーマからTypeScriptクライアントを自動生成：

```bash
# コンテナ内で実行
docker compose exec frontend npm run generate:api

# またはローカルで実行
npm run generate:api
```

生成されたコードは `src/lib/api/generated/` に出力されます。

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

## デプロイ

### Cloudflare Pages

#### ビルド設定

**Staging環境:**
- ビルドコマンド: `npm run build:staging`
- ビルド出力ディレクトリ: `out`
- 環境変数: `NEXT_PUBLIC_API_URL=https://myanimelogs-staging.fly.dev`

**Production環境:**
- ビルドコマンド: `npm run build:production`
- ビルド出力ディレクトリ: `out`
- 環境変数: `NEXT_PUBLIC_API_URL=https://myanimelogs-production.fly.dev`

#### セキュリティヘッダー

`_headers.staging` と `_headers.production` ファイルには以下が含まれています：
- Content Security Policy (CSP)
- X-Frame-Options
- X-Content-Type-Options
- Referrer-Policy

ビルド時に適切な `_headers` ファイルが `out/` ディレクトリにコピーされます。

#### 画像ソース

CSPで以下の画像ソースが許可されています：
- `https://s4.anilist.co` - AniList CDN（メイン）
- `https://*.anilist.co` - AniList CDN（全サブドメイン）
- `https://cdn.myanimelist.net` - MyAnimeList CDN
- `https://pbs.twimg.com` - Twitter画像
- `https://*.twimg.com` - Twitter画像（全サブドメイン）
