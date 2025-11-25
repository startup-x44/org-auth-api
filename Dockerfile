# Build stage
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Install git (required for go mod download)
RUN apk add --no-cache git

# Copy go mod and sum files (if they exist)
COPY go.mod* go.sum* ./

# Download dependencies and tidy
RUN go mod tidy && go mod download

# Copy source code
COPY . .

# Tidy dependencies again after copying source code
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Final stage
FROM alpine:latest

# Install ca-certificates and create non-root user
RUN apk --no-cache add ca-certificates && \
    adduser -D -g '' appuser

WORKDIR /root/
# Change to /app for non-root user
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Set ownership
RUN chown appuser:appuser /app/main

# Expose port
EXPOSE 8080

# Switch to non-root user
USER appuser

# Command to run
CMD ["./main"]