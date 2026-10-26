# --- Airビルドステージ ---
FROM golang:1.26.10-alpine3.24 AS air-builder
WORKDIR /src
# Airをインストール
RUN go install github.com/cosmtrek/air@v1.49.0

# --- メインアプリケーションステージ ---
FROM golang:1.26.10-alpine3.24
# パッケージを更新して脆弱性を修正
RUN apk update && apk upgrade
# 開発ツールに必要なパッケージをインストール
RUN apk add --no-cache procps curl

# air-builderステージからAirバイナリをコピー
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