# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy the Go files
COPY go.mod go.sum* ./
RUN go mod download

COPY . .

# Build the binary with optimizations (omit debug info/symbols for smaller size)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o auth-service ./cmd/api/main.go

# Production stage
FROM alpine:3.18

# Install standard certificates and timezone database
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy compiled binary from builder
COPY --from=builder /app/auth-service .

# Expose default port
EXPOSE 4001

# Run service
CMD ["./auth-service"]
