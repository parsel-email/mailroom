# Stage 1: Build the application
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates make

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application with security flags enabled
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o service ./cmd/api

# Stage 2: Create the minimal runtime image
FROM alpine:3.21

# Add CA certificates and timezone data
RUN apk add --no-cache ca-certificates tzdata && \
    update-ca-certificates

# Create a non-root user to run the application
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/service .

# Copy database migration files
COPY --from=builder /app/db/migrations ./db/migrations

# Set ownership of the application files to appuser
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose the port the service runs on
EXPOSE 8080

# Command to run the application
CMD ["./service"]