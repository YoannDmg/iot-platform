# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git for fetching dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum* ./

# Copy shared proto (needed for local replace directive)
COPY ../../shared/proto /shared/proto

# Update go.mod to use absolute path for shared proto
RUN go mod edit -replace github.com/yourusername/iot-platform/shared/proto=/shared/proto

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-gateway .

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS and wget for health checks
RUN apk --no-cache add ca-certificates wget

# Copy binary from builder
COPY --from=builder /app/api-gateway .

# Expose HTTP port
EXPOSE 8080

# Run the service
CMD ["./api-gateway"]
