# syntax=docker/dockerfile:1
FROM golang:1.26.5-alpine AS builder
WORKDIR /build
COPY go.mod main.go ./
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags="-s -w" -o padduck-agent .

FROM alpine:3.22
# libcap/setcap: grant ping cap_net_raw explicitly so it works for the
# non-root user (Docker's default capability bounding set includes NET_RAW,
# so no extra runtime flags are needed).
RUN apk add --no-cache iputils libcap libssl3=3.5.7-r0 libcrypto3=3.5.7-r0 \
    && setcap cap_net_raw+ep "$(command -v ping)" \
    && addgroup -S padduck && adduser -S -G padduck padduck
COPY --from=builder /build/padduck-agent /usr/local/bin/padduck-agent
USER padduck
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD pidof padduck-agent >/dev/null || exit 1
ENTRYPOINT ["/usr/local/bin/padduck-agent"]
