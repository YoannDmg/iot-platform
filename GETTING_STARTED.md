# Getting Started Guide

Complete guide to set up and run the IoT platform locally.

## Prerequisites

### Required
- **Docker Desktop** - Container orchestration
- **Go 1.21+** - Backend services runtime
- **Protocol Buffers Compiler**
  ```bash
  brew install protobuf
  ```

### Optional
- Rust 1.75+ (Data Collector service)
- Node.js 20+ (Web frontend)
- Flutter 3.x (Mobile application)

## Architecture Overview

```
┌────────────────────┐
│  Web/Mobile Client │
└─────────┬──────────┘
          │ GraphQL
    ┌─────▼──────────┐
    │  API Gateway   │  :8080 (GraphQL)
    │     (Go)       │
    └─────────┬──────┘
              │ gRPC
    ┌─────────▼──────────┐
    │  Device Manager    │  :8081 (gRPC)
    │      (Go)          │
    └─────────┬──────────┘
              │
      ┌───────┴────────┐
      │   PostgreSQL   │  :5432
      │   Redis        │  :6379
      │   MQTT         │  :1883
      └────────────────┘
```

## Quick Start

### 1. Installation

```bash
cd iot-platform
make setup
```

This command will:
- Verify `protoc` installation
- Install gRPC plugins for Go
- Install `gqlgen` (GraphQL code generator)
- Download Go dependencies

### 2. Code Generation

```bash
make generate
```

Generates code from:
1. **Protocol Buffers** (`shared/proto/device.proto`)
   - Go structs
   - gRPC client/server code

2. **GraphQL schema** (`services/api-gateway/schema.graphql`)
   - Resolver interfaces
   - Type definitions

### 3. Start Infrastructure

```bash
make start
```

Verify services are running:
```bash
make status
```

**Infrastructure endpoints:**
- PostgreSQL: `localhost:5432` (user: `iot_user`, password: `iot_password`)
- Redis: `localhost:6379`
- MQTT: `localhost:1883`
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)

### 4. Start Application Services

**Terminal 1 - Device Manager (gRPC):**

```bash
make device-manager
```

Expected output:
```
Device Manager Service
Protocole: gRPC (HTTP/2)
Port: 8081
✅ Serveur démarré
```

**Terminal 2 - API Gateway (GraphQL):**
```bash
make api-gateway
```

Expected output:
```
🚀 API Gateway démarré sur le port 8080
📊 GraphQL Playground: http://localhost:8080/
```

## Testing the API

### GraphQL Playground

Open http://localhost:8080 in your browser for an interactive GraphQL interface.

### Create a Device

```graphql
mutation {
  createDevice(input: {
    name: "Temperature Sensor Living Room"
    type: "temperature_sensor"
    metadata: [
      { key: "location", value: "living_room" }
      { key: "floor", value: "1" }
    ]
  }) {
    id
    name
    type
    status
    createdAt
  }
}
```

Response:
```json
{
  "data": {
    "createDevice": {
      "id": "abc-123-def",
      "name": "Temperature Sensor Living Room",
      "type": "temperature_sensor",
      "status": "ONLINE",
      "createdAt": 1234567890
    }
  }
}
```

### List Devices

```graphql
query {
  devices(page: 1, pageSize: 10) {
    devices {
      id
      name
      type
      status
    }
    total
  }
}
```

### Query Statistics

```graphql
query {
  stats {
    totalDevices
    onlineDevices
    offlineDevices
  }
}
```

## Architecture Deep Dive

### Data Flow

```
Client (Browser/Mobile)
   │ GraphQL query (JSON/HTTP)
   ▼
API Gateway (:8080)
   │ Parse GraphQL request
   │ Call resolver
   │ gRPC call (Binary/HTTP2)
   ▼
Device Manager (:8081)
   │ Business logic
   │ Data persistence
   │ gRPC response (Protobuf)
   ▼
API Gateway
   │ Convert to JSON
   │ GraphQL response
   ▼
Client
```

### Technology Choices

**GraphQL (Public API)**
- Flexible data fetching
- Single endpoint
- Auto-generated documentation
- Type-safe queries

**gRPC (Internal)**
- High performance (binary, HTTP/2)
- Strict contracts (Protocol Buffers)
- Strong typing
- Built-in streaming support

## Common Commands

```bash
make help              # Show all available commands
make setup             # Initial setup (tools + dependencies)
make generate          # Generate code (proto + GraphQL)
make start             # Start infrastructure
make stop              # Stop infrastructure
make status            # View container status
make logs              # View Docker logs
make device-manager    # Run Device Manager
make api-gateway       # Run API Gateway
```

## Troubleshooting

### `protoc` not found
```bash
brew install protobuf
```

### Go plugins not in PATH
```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Port already in use
```bash
lsof -i :8080          # Find process using port
kill -9 <PID>          # Kill the process
```

### Reset and regenerate
```bash
make clean
make generate
make start
```

## Code Generation Explained

### Protocol Buffers
Define service contract in `.proto` files:
```protobuf
message Device {
  string id = 1;
  string name = 2;
}

service DeviceService {
  rpc CreateDevice(CreateDeviceRequest) returns (CreateDeviceResponse);
}
```

Generated Go code provides:
- Typed structs
- gRPC server/client interfaces
- Serialization/deserialization

### GraphQL
Define schema in `.graphql` files:
```graphql
type Device {
  id: ID!
  name: String!
}

type Mutation {
  createDevice(input: CreateDeviceInput!): Device
}
```

Generated Go code provides:
- Resolver interfaces
- Type definitions
- Query validation

## Next Steps

1. **Database Integration** - Connect Device Manager to PostgreSQL
2. **Authentication** - Implement JWT authentication in API Gateway
3. **Data Collector** - Build Rust service for MQTT data ingestion
4. **Web Dashboard** - Create React frontend
5. **Mobile App** - Build Flutter application

## FAQ

**Why use code generation?**
Ensures type safety, reduces boilerplate, maintains strict contracts between services.

**When to regenerate?**
Only when modifying `.proto` or `.graphql` files.

**Testing gRPC directly?**
Install `grpcurl` to call Device Manager directly:
```bash
grpcurl -plaintext localhost:8081 list
```

**Production readiness?**
Current implementation uses in-memory storage. See TODO comments in code for production considerations.
