# Stage 1: Build
FROM golang:1.26-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o main cmd/api/main.go

# Stage 2: Final
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .
# Copy configs and other necessary files
COPY --from=builder /app/configs ./configs
# Copy migrations if they are needed by the app at runtime
COPY --from=builder /app/migrations ./migrations

# Expose port
EXPOSE 8000

# Command to run
CMD ["./main"]
