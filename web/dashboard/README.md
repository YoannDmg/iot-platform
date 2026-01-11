# IoT Platform Dashboard

Web dashboard for monitoring IoT devices in real-time.

## Tech Stack

- React 19 + TypeScript
- Vite 7
- TailwindCSS 4
- TanStack Query (React Query)
- GraphQL (via graphql-request)
- React Router
- Recharts (for future metrics visualization)

## Features

- Real-time device monitoring
- Auto-refresh every 5-10 seconds
- Device list view
- Detailed device metrics view (CPU, RAM, Disk, Battery, Network, etc.)
- Responsive design with dark mode support

## Installation

```bash
npm install
```

**Note:** This project uses TailwindCSS 4 which requires `@tailwindcss/postcss` instead of the traditional config file.

## Configuration

Create a [.env](.env) file in the root directory:

```env
VITE_GRAPHQL_ENDPOINT=http://localhost:8080/query
```

## Development

```bash
npm run dev
```

The app will be available at `http://localhost:5173`

## Build

```bash
npm run build
```

## Project Structure

```
src/
├── components/       # Reusable UI components
│   ├── DeviceStatus.tsx
│   └── MetricCard.tsx
├── hooks/           # Custom React hooks
│   ├── useDevice.ts
│   └── useDevices.ts
├── lib/             # Utilities and configs
│   ├── graphql-client.ts
│   ├── queries.ts
│   └── utils.ts
├── pages/           # Page components
│   ├── DeviceDetail.tsx
│   └── DeviceList.tsx
├── types/           # TypeScript types
│   └── device.ts
└── App.tsx
```

## GraphQL Schema

The dashboard expects the following GraphQL schema:

```graphql
type Device {
  id: ID!
  name: String!
  type: String!
  status: DeviceStatus!
  metadata: [DeviceMetadata!]
  createdAt: String!
  updatedAt: String!
}

type DeviceMetadata {
  key: String!
  value: String!
}

enum DeviceStatus {
  ONLINE
  OFFLINE
  INACTIVE
}
```

## Metrics Format

Device metrics are stored in the `metadata` field with the following keys:

- `cpu_percent` - CPU usage percentage
- `memory_used_gb` / `memory_gb` - Memory usage in GB
- `disk_used_gb` / `disk_gb` - Disk usage in GB
- `network_up_mb` - Network upload in MB
- `network_down_mb` - Network download in MB
- `battery_level` - Battery level percentage
- `process_count` - Number of active processes
