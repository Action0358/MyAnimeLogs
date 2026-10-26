# Node.js 22 LTSを使用（開発環境最適化）
FROM node:22-slim

# 作業ディレクトリの設定
WORKDIR /frontend

# 必要なパッケージをインストール
RUN apt-get update && apt-get install -y \
    python3 \
    make \
    g++ \
    curl \
    procps \
    git \
    && rm -rf /var/lib/apt/lists/*

# package.jsonとpackage-lock.jsonを先にコピー（キャッシュ最適化）
COPY package*.json ./

# 依存関係のインストール
RUN npm ci

# ソースコード全体をコピー
COPY . .

# Next.jsキャッシュディレクトリの作成
RUN mkdir -p .next

# 開発用環境変数設定
ENV NODE_ENV=development
ENV CHOKIDAR_USEPOLLING=true
ENV CHOKIDAR_INTERVAL=1000
ENV FAST_REFRESH=true

# ポートの公開
EXPOSE 3000

# 開発サーバーの起動（ホットリロード有効）
CMD ["npm", "run", "dev"]