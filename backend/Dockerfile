
FROM golang:1.26-alpine
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o app
CMD ["./app"]
