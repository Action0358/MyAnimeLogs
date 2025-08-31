# Goがインストール済みのDebianベースイメージを使用
FROM golang:1.22-bullseye

# 開発ツールに必要なパッケージをインストール
RUN apt-get update && apt-get install -y procps curl && rm -rf /var/lib/apt/lists/*

# 作業ディレクトリを設定
WORKDIR /app

# Goモジュールのファイルをコピー
COPY go.mod go.sum ./

# ソースコード全体をコピー
COPY . .

# ソースコードに基づいて依存関係を自動で整理・更新
RUN go mod tidy

# 依存関係をダウンロード
RUN go mod download

# ポートを公開
EXPOSE 8080

# 開発用に `go run` でサーバーを起動
CMD ["go", "run", "./cmd/server"]