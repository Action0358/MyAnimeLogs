# --- Builder Stage for Air ---
FROM golang:1.22-bullseye AS air-builder
WORKDIR /src
# Install Air
RUN go install github.com/cosmtrek/air@v1.49.0

# --- Main Application Stage ---
FROM golang:1.22-bullseye
# 開発ツールに必要なパッケージをインストール
RUN apt-get update && apt-get install -y procps curl && rm -rf /var/lib/apt/lists/*

# Copy Air binary from the builder stage
COPY --from=air-builder /go/bin/air /usr/local/bin/

# 作業ディレクトリを設定
WORKDIR /backend

# Goモジュールのファイルをコピー
COPY go.mod go.sum ./

# 依存関係をダウンロード
RUN go mod download

# ソースコード全体をコピー
COPY . .

# tmpディレクトリを作成（Airが使用）
RUN mkdir -p tmp

# ポートを公開
EXPOSE 8080

# 開発用にAirでホットリロード起動
CMD ["air", "-c", ".air.toml"]