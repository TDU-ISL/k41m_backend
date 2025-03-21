# # ベースイメージ
# FROM golang:1.23.0 as builder

# # 作業ディレクトリを設定
# WORKDIR /app

# # モジュールをキャッシュ
# COPY go.mod go.sum ./
# RUN go mod tidy

# # ソースコードをコピー
# COPY . .

# # 静的リンクでビルド
# RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# # 実行イメージ(skaffold dev で使うため、slimイメージを使用)
# FROM debian:bullseye-slim 
# WORKDIR /app
# COPY --from=builder /app/main .
# EXPOSE 8080
# CMD ["./main"]

# 開発中はホットリロードをするためマルチステージビルドを使わない
# ベースイメージ
FROM public.ecr.aws/docker/library/golang:1.23.0

# 作業ディレクトリを設定
WORKDIR /app

# Air をインストール
RUN go install github.com/cosmtrek/air@v1.40.4

# 必要なパッケージをインストール（Air が利用する場合もある inotify-tools）
RUN apt-get update && \
    apt-get install -y inotify-tools && \
    apt-get install -y wget && \
    wget -qO trivy.deb https://github.com/aquasecurity/trivy/releases/download/v0.58.0/trivy_0.58.0_Linux-64bit.deb && \
    dpkg -i trivy.deb && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Go モジュールをコピーして依存関係を解決
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# Air 設定ファイルを含める（air.toml が存在する場合）
COPY air.toml .

# Air をデフォルトコマンドとして起動
CMD ["air"]
