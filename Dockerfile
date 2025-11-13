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

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Expose port
EXPOSE 8080

# Command to run
CMD ["./main"]