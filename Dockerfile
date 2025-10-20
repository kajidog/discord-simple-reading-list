# Build stage
FROM golang:1.22-bullseye AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bot ./cmd/bot

# Runtime stage
FROM debian:bullseye-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Create a non-root user and set proper permissions
RUN useradd -m -u 1000 appuser && \
    chown -R appuser:appuser /app

COPY --from=builder /bot /usr/local/bin/bot

USER appuser

ENTRYPOINT ["/usr/local/bin/bot"]
