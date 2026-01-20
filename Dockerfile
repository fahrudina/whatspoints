# Use the official Golang image as the base image
FROM golang:1.24-alpine AS builder

# Install build dependencies (gcc and musl-dev needed for CGO with SQLite)
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    gcc \
    musl-dev \
    sqlite-dev

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the application with optimizations
# CGO_ENABLED=1 needed for SQLite driver, -ldflags for smaller size
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o whatspoints main.go

# Use a minimal base image for the final container
FROM alpine:latest

# Install runtime dependencies (SQLite libraries needed for CGO binary)
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    sqlite-libs

# Create non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/whatspoints .

# Copy the web directory for static files
COPY --from=builder /app/web ./web

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose the port the service will run on
EXPOSE 8080

# Command to run the application
CMD ["./whatspoints"]
