# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/api/main.go

# Final stage — Ubuntu 24.04 has GLIBC 2.39 (whisper-cli needs >= 2.38)
FROM ubuntu:24.04

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates tzdata ffmpeg \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN groupadd -r appgroup && useradd -r -g appgroup appuser

COPY --from=builder /app/main .

# Copy whisper-cli binary and model
COPY bin/whisper-cli /app/bin/whisper-cli
COPY bin/ggml-small.bin /app/bin/ggml-small.bin
RUN chmod +x /app/bin/whisper-cli

# Create uploads directory
RUN mkdir -p /app/uploads && chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

CMD ["./main"]
