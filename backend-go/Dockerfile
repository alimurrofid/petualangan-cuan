# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git jika diperlukan untuk dependencies
# RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/api/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# HAPUS COPY .env.example .env (Security Risk)
# .env harus di-inject saat runtime atau mount volume

COPY --from=builder /app/main .

# Change ownership
RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

CMD ["./main"]
