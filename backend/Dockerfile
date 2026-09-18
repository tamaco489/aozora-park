# ビルドコンテキストは backend/ (docker build -f backend/Dockerfile backend)

# Go のバージョンは go.mod と揃える、OS まで含めてタグを固定する
FROM golang:1.27.1-trixie AS build

# worker が増えても Dockerfile を増やさず、cmd/<SERVICE> を切り替えて使い回す
ARG SERVICE=api

WORKDIR /src

# go.mod と go.sum だけを先にコピーし、ソースの変更で依存のダウンロードをやり直さない
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 実行用のイメージに libc が無いため静的リンクにする、-trimpath はビルド環境のパスをバイナリに残さない
# GOTOOLCHAIN=local はビルド中に別の Go をダウンロードさせず、イメージの Go で失敗させるため
RUN CGO_ENABLED=0 GOTOOLCHAIN=local go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/${SERVICE}

# シェルもパッケージマネージャも持たない最小のイメージで、root 以外のユーザーで動かす
FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=build /out/server /server

# Cloud Run は PORT を渡すが、ローカルで起動するときの既定と揃える
EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/server"]
