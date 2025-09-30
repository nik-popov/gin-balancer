# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.22.7-bullseye AS builder

WORKDIR /app

# Cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy the remaining sources and build the binary
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags "-s -w" \
    -o gin-balancer ./main.go

# Runtime stage
FROM debian:bookworm-slim AS runtime

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create a non-root user to run the service
RUN useradd --system --create-home --uid 10001 appuser

WORKDIR /home/appuser

COPY --from=builder /app/gin-balancer /usr/local/bin/gin-balancer

ENV PORT=8080
EXPOSE 8080

USER appuser

ENTRYPOINT ["gin-balancer"]
