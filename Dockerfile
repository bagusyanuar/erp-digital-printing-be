# Stage 1: Build
FROM golang:1.26-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git curl

# Optimize module download speed
ENV GOPROXY=https://proxy.golang.org,direct

WORKDIR /app

# Download golang-migrate binary
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.2/migrate.linux-amd64.tar.gz | tar xvz \
    && mv migrate ./migrate_cli

# Copy go mod and sum files first (layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build API binary
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/api/main.go

# Build Seeder binary
RUN CGO_ENABLED=0 GOOS=linux go build -o seed_bin cmd/seed/main.go

# Stage 2: Final (pinned version for reproducibility)
FROM alpine:3.21

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy API binary from builder
COPY --from=builder /app/main .
# Copy Seeder binary from builder
COPY --from=builder /app/seed_bin ./seed
# Copy migrate cli from builder
COPY --from=builder /app/migrate_cli ./migrate
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
