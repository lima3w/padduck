# syntax=docker/dockerfile:1
FROM golang:1.26.3-alpine AS builder
WORKDIR /build
COPY go.mod main.go ./
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags="-s -w" -o ipam-agent .

FROM alpine:3.22
RUN apk add --no-cache iputils
COPY --from=builder /build/ipam-agent /usr/local/bin/ipam-agent
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD pidof ipam-agent >/dev/null || exit 1
ENTRYPOINT ["/usr/local/bin/ipam-agent"]
