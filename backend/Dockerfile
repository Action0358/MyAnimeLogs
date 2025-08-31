# Goがインストール済みのDebianベースイメージを使用
FROM golang:1.22-bullseye

# 開発ツールに必要なパッケージをインストール
RUN apt-get update && apt-get install -y procps curl && rm -rf /var/lib/apt/lists/*

# 作業ディレクトリを設定
WORKDIR /backend

# Goモジュールのファイルをコピー
COPY go.mod go.sum ./

# 依存関係をダウンロード
RUN go mod download

# Air（ホットリロードツール）をインストール（Go 1.22対応バージョン）
RUN go install github.com/cosmtrek/air@v1.49.0

# ソースコード全体をコピー
COPY . .

# ソースコードに基づいて依存関係を自動で整理・更新
RUN go mod tidy

# tmpディレクトリを作成（Airが使用）
RUN mkdir -p tmp

# ポートを公開
EXPOSE 8080

# 開発用にAirでホットリロード起動
CMD ["air", "-c", ".air.toml"]