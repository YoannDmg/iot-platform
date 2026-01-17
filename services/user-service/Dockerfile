# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git for fetching dependencies
RUN apk add --no-cache git

# Copy go.mod + go.sum from repo root
COPY services/user-service/go.mod services/user-service/go.sum ./


# Copy shared proto
COPY shared/proto /shared/proto

# Update go.mod replace
RUN go mod edit -replace github.com/yourusername/iot-platform/shared/proto=/shared/proto

# Download dependencies
RUN go mod download

# Copy user-service source code
COPY services/user-service ./

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o user-service .

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/user-service .

# Expose gRPC port
EXPOSE 8082

# Run the service
CMD ["./user-service"]
