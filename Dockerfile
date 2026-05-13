# Stage 1: Build
FROM golang:1.26-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Optimize module download speed
ENV GOPROXY=https://proxy.golang.org,direct

WORKDIR /app

# Copy go mod and sum files first (layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build static binary (no CGO dependency)
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/api/main.go

# Stage 2: Final (pinned version for reproducibility)
FROM alpine:3.21

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .
# Copy configs and other necessary files
COPY --from=builder /app/configs ./configs
# Copy migrations if they are needed by the app at runtime
COPY --from=builder /app/migrations ./migrations

# Set ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8000

# Command to run
CMD ["./main"]
