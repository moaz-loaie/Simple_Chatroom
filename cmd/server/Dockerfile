FROM golang:1.25 AS builder
WORKDIR /src

# Cache Go modules
COPY go.mod ./
RUN go env -w GOPROXY=https://proxy.golang.org || true
RUN go mod download || true

# Copy rest of the source
COPY . .

# Build static linux binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/server ./cmd/server

FROM alpine:3.18
RUN apk add --no-cache ca-certificates

# unprivileged user
RUN addgroup -g 1000 app && adduser -D -G app -u 1000 app

COPY --from=builder /app/server /usr/local/bin/server
USER app
WORKDIR /home/app

EXPOSE 8080/tcp

ENTRYPOINT ["/usr/local/bin/server"]
CMD ["-port","8080"]
