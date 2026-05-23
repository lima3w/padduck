# syntax=docker/dockerfile:1
FROM golang:1.26.3-alpine AS builder
ARG APP_VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown
WORKDIR /app
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -mod=vendor -trimpath \
    -ldflags="-s -w -X padduck/version.Version=${APP_VERSION} -X padduck/version.Commit=${GIT_COMMIT} -X padduck/version.BuildDate=${BUILD_DATE}" \
    -o app .

FROM alpine:3.22
RUN addgroup -S app && adduser -S app -G app && \
    apk add --no-cache su-exec postgresql17-client
WORKDIR /app
COPY --from=builder /app/app .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/docs ./docs
RUN chown app:app /app && mkdir -p /app/data
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
# No USER directive — entrypoint runs as root to fix bind-mount ownership,
# then drops to the unprivileged app user via su-exec before exec'ing the binary.
HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 CMD wget -qO- "http://127.0.0.1:${SERVER_PORT:-8080}/health" | grep -q '"status":"ok"' || exit 1
ENTRYPOINT ["/entrypoint.sh"]
CMD ["./app"]
