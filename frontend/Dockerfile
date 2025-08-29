# DebianベースのNode.jsイメージを使用
FROM node:18-slim

# 作業ディレクトリの設定
WORKDIR /frontend

# 必要なパッケージをインストール
RUN apt-get update && apt-get install -y \
    python3 \
    make \
    g++ \
    curl \
    procps \
    && rm -rf /var/lib/apt/lists/*

# ソースコードのコピー
COPY . .

# パッケージのインストール
RUN npm install

# ポートの公開
EXPOSE 3000

# アプリケーションの起動
CMD ["npm", "start"]