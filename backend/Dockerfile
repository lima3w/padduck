FROM golang:1.26.2-alpine AS builder
ARG APP_VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown
WORKDIR /app
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -trimpath -ldflags="-s -w -X ipam-next/version.Version=${APP_VERSION} -X ipam-next/version.Commit=${GIT_COMMIT} -X ipam-next/version.BuildDate=${BUILD_DATE}" -o app .

FROM alpine:3.21
RUN addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=builder /app/app .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/docs ./docs
USER app
HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 CMD wget -qO- "http://127.0.0.1:${SERVER_PORT:-8080}/health" | grep -q '"status":"ok"' || exit 1
CMD ["./app"]
