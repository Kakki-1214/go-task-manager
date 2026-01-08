# 1. ベースイメージの指定（軽量なAlpine版）
FROM golang:1.25-alpine

# 2. 必要なツールのインストール
# git: go mod downloadに必要
# air: ホットリロード用ツール
RUN apk update && apk add --no-cache git && \
    go install github.com/air-verse/air@v1.52.3

# 3. 作業ディレクトリの設定
WORKDIR /app

# 4. go.mod と go.sum をコピーして依存関係をダウンロード
# 先にこれやることで、コード変更時のビルド時間を短縮（キャッシュ利用）
COPY go.mod go.sum* ./
RUN go mod download

# 5. ソースコードをコピー
COPY . .

# 6. Airを使ってアプリを起動（ホットリロード有効化）
CMD ["air"]