# Stage 1: ビルド
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/api main.go

FROM alpine:3.19

RUN apk --no-cache add ca-certificates && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

# CGO無効・最適化フラグ付きでビルド
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s' \
    -o /app/api main.go

# Stage 2: 実行（最小・安全なイメージ）
FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /app/api .

# nonroot ユーザーで実行（セキュリティ）
USER nonroot:nonroot

RUN chown appuser:appgroup /app/api

USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/api"]
