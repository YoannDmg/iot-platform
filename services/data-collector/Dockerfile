# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git for fetching dependencies
RUN apk add --no-cache git

# Copy go.mod + go.sum from repo root
COPY services/data-collector/go.mod services/data-collector/go.sum ./

# Copy shared proto
COPY shared/proto /shared/proto

# Update go.mod replace
RUN go mod edit -replace github.com/yourusername/iot-platform/shared/proto=/shared/proto

# Download dependencies
RUN go mod download

# Copy service source
COPY services/data-collector ./

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o data-collector .

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install certs
RUN apk add --no-cache ca-certificates

# Copy binary
COPY --from=builder /app/data-collector .

# Expose gRPC port
EXPOSE 8083

# Run
CMD ["./data-collector"]
