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
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o device-manager .

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/device-manager .

# Expose gRPC port
EXPOSE 8081

# Run the service
CMD ["./device-manager"]
