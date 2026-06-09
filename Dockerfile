# Stage 1 — Build
FROM golang:1.23-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always 2>/dev/null || echo dev)" \
    -trimpath \
    -o cairn \
    ./cmd/cairn

# Stage 2 — Image finale
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/cairn /cairn
COPY --from=builder /build/web /web

VOLUME ["/data"]
EXPOSE 8080

ENTRYPOINT ["/cairn"]
