FROM golang:1.26.2-alpine AS builder
WORKDIR /build
COPY go.mod ./
COPY main.go ./
RUN go build -ldflags="-s -w" -o ipam-agent .

FROM alpine:3.22
RUN apk add --no-cache iputils
COPY --from=builder /build/ipam-agent /usr/local/bin/ipam-agent
ENTRYPOINT ["/usr/local/bin/ipam-agent"]
