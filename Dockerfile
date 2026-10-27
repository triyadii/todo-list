# ==============================================================================
# Multi-stage Dockerfile for TodoList Go Fiber Backend
# ==============================================================================

# --- Stage 1: Build binary ---
FROM golang:alpine AS builder

WORKDIR /app

# Install required build utilities
RUN apk add --no-cache git ca-certificates tzdata

# Download dependencies (leveraging Docker layer cache)
COPY go.mod go.sum ./
RUN go mod download

# Copy application source code
COPY . .

# Build statically linked binary with optimizations (-w -s to reduce binary size)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/todolist-api ./cmd/api

# --- Stage 2: Lightweight runtime image ---
FROM alpine:3.20

WORKDIR /app

# Install tzdata (crucial for timezones like Asia/Jakarta) and CA certificates
RUN apk add --no-cache ca-certificates tzdata

# Set default timezone
ENV TZ=Asia/Jakarta

# Copy executable binary and swagger docs from builder stage
COPY --from=builder /app/bin/todolist-api /app/todolist-api
COPY --from=builder /app/docs /app/docs

# Expose API application port
EXPOSE 8080

# Run the backend API
ENTRYPOINT ["/app/todolist-api"]
