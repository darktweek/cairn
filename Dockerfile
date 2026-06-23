# Stage 1 — Build
FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN GONOSUMDB=* GOFLAGS=-mod=mod go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GONOSUMDB=* \
    go build \
    -mod=mod \
    -ldflags="-s -w -X main.version=$(git describe --tags --always 2>/dev/null || echo dev)" \
    -trimpath \
    -o cairn \
    ./cmd/cairn

# Stage 2 — Image finale
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/cairn /cairn

VOLUME ["/data"]
EXPOSE 8080

ENTRYPOINT ["/cairn"]
