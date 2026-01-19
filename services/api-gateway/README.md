# API Gateway

> Point d'entrée GraphQL de la plateforme IoT avec authentification JWT et streaming temps réel

[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go&logoColor=white)](https://golang.org)
[![GraphQL](https://img.shields.io/badge/GraphQL-E10098?logo=graphql&logoColor=white)](https://graphql.org)
[![gRPC](https://img.shields.io/badge/gRPC-HTTP%2F2-4285F4)](https://grpc.io)
[![WebSocket](https://img.shields.io/badge/WebSocket-Subscriptions-blue)](https://developer.mozilla.org/en-US/docs/Web/API/WebSocket)

## Table des matières

- [Vue d'ensemble](#vue-densemble)
- [Architecture](#architecture)
- [Démarrage rapide](#démarrage-rapide)
- [Configuration](#configuration)
- [Authentification](#authentification)
- [API GraphQL](#api-graphql)
- [Subscriptions temps réel](#subscriptions-temps-réel)
- [Développement](#développement)

## Vue d'ensemble

L'API Gateway est le point d'entrée unique de la plateforme. Il expose une API GraphQL pour les clients (Dashboard, applications mobiles) et communique avec les microservices backend via gRPC. Il supporte également les subscriptions GraphQL via WebSocket pour le streaming de données en temps réel.

### Fonctionnalités

- **API GraphQL** — Schéma typé, playground intégré
- **Authentification JWT** — Tokens HS256, expiration 24h
- **Autorisation par rôles** — admin, user, device
- **Clients gRPC** — Connexion aux 3 microservices
- **CORS** — Support cross-origin pour le frontend
- **WebSocket** — Subscriptions GraphQL temps réel
- **Redis Pub/Sub** — Réception des événements télémétrie

### Technologies

| Composant | Technologie |
|-----------|-------------|
| Langage | Go 1.24 |
| API | GraphQL (gqlgen) |
| Auth | JWT (HS256) |
| Backend | gRPC clients |
| Temps réel | WebSocket + Redis Pub/Sub |

## Architecture

```
┌─────────────┐  GraphQL   ┌──────────────────────────────────┐
│  Dashboard  │◄──────────►│          API Gateway             │
│  React+Vite │  HTTP/WS   │           Port 8080              │
└─────────────┘            ├──────────────────────────────────┤
                           │  CORS → JWT Middleware → GraphQL │
                           │  + Auth Extension + WebSocket    │
                           └──────────┬───────────────────────┘
                                      │ gRPC            ▲
                    ┌─────────────────┼─────────────────┤
                    │                 │                 │ Redis
                    ▼                 ▼                 │ Pub/Sub
          ┌──────────────┐  ┌──────────────┐  ┌────────┴─────┐
          │Device Manager│  │ User Service │  │    Redis     │
          │  Port 8081   │  │  Port 8082   │  │  Port 6379   │
          └──────────────┘  └──────────────┘  └──────────────┘
                    │                                   ▲
                    │              gRPC                 │ Pub
                    │       ┌──────────────┐            │
                    └──────►│Data Collector├────────────┘
                            │  Port 8083   │
                            └──────────────┘
```

### Structure du projet

```
api-gateway/
├── main.go                 # Point d'entrée, configuration, WebSocket
├── schema.graphql          # Schéma GraphQL (queries, mutations, subscriptions)
├── gqlgen.yml              # Configuration gqlgen
├── auth/
│   ├── jwt.go              # Génération et validation JWT
│   ├── middleware.go       # Middleware HTTP + WebSocket auth
│   └── graphql_auth.go     # Extension GraphQL d'autorisation
├── grpc/
│   └── client.go           # Clients gRPC (Device, User, Telemetry)
├── pubsub/
│   ├── broker.go           # Broker in-memory pour subscriptions
│   └── redis.go            # Subscriber Redis (écoute iot:telemetry:*)
├── graph/
│   ├── resolver.go         # Injection des dépendances (+ Broker)
│   ├── schema.resolvers.go # Resolvers (queries, mutations, subscriptions)
│   ├── generated/          # Code généré (ne pas modifier)
│   └── model/              # Modèles GraphQL générés
└── Dockerfile
```

## Démarrage rapide

### Prérequis

- Go 1.24+
- Redis (pour les subscriptions temps réel)
- Services backend actifs (Device Manager, User Service, Data Collector)

### Lancement

```bash
# Depuis la racine du projet
make dev-api

# Ou directement
cd services/api-gateway
go run main.go
```

Le serveur démarre sur `http://localhost:8080`.

### Endpoints

| Endpoint | Protocole | Description |
|----------|-----------|-------------|
| `/` | HTTP | GraphQL Playground |
| `/query` | HTTP | API GraphQL (queries, mutations) |
| `/query` | WebSocket | Subscriptions GraphQL |
| `/health` | HTTP | Health check |

## Configuration

### Variables d'environnement

| Variable | Description | Défaut |
|----------|-------------|--------|
| `PORT` | Port HTTP/WebSocket | `8080` |
| `DEVICE_MANAGER_ADDR` | Adresse Device Manager | `localhost:8081` |
| `USER_SERVICE_ADDR` | Adresse User Service | `localhost:8082` |
| `TELEMETRY_SERVICE_ADDR` | Adresse Data Collector | `localhost:8083` |
| `JWT_SECRET` | Clé secrète JWT | `dev-jwt-secret-...` |
| `REDIS_HOST` | Hôte Redis | `localhost` |
| `REDIS_PORT` | Port Redis | `6379` |

## Authentification

### Flux JWT

1. L'utilisateur s'inscrit via `register` ou se connecte via `login`
2. Le serveur retourne un token JWT (valide 24h)
3. Le client inclut le token dans le header : `Authorization: Bearer <token>`
4. Le middleware valide le token et injecte les claims dans le contexte

### Claims JWT

```go
type Claims struct {
    UserID string
    Email  string
    Name   string
    Role   string  // admin, user, device
}
```

### Opérations publiques

Ces opérations ne nécessitent pas d'authentification :

- `register` — Inscription
- `login` — Connexion
- `telemetryReceived` — Subscription (auth via WebSocket connectionParams)
- Introspection GraphQL

Toutes les autres opérations requièrent un token valide.

### Authentification WebSocket

Pour les subscriptions, le token JWT est passé via `connectionParams` :

```javascript
const wsLink = new GraphQLWsLink(createClient({
  url: 'ws://localhost:8080/query',
  connectionParams: {
    authorization: 'Bearer <token>',
  },
}));
```

### Rôles

| Rôle | Permissions |
|------|-------------|
| `admin` | Accès complet, gestion des utilisateurs |
| `user` | CRUD devices, lecture télémétrie |
| `device` | Envoi de télémétrie uniquement |

## API GraphQL

### Queries

```graphql
# Utilisateur courant
me: User

# Liste des utilisateurs (admin)
users(page: Int, pageSize: Int, role: String): UsersResponse

# Devices
device(id: ID!): Device
devices(page: Int, pageSize: Int, type: String, status: String): DevicesResponse
stats: Stats

# Télémétrie
deviceTelemetry(deviceId: ID!, metricName: String!, startTime: Int!, endTime: Int!, limit: Int): TelemetrySeries
deviceTelemetryAggregated(deviceId: ID!, metricName: String!, startTime: Int!, endTime: Int!, interval: String!): [TelemetryAggregation!]!
deviceLatestMetric(deviceId: ID!, metricName: String!): TelemetryPoint
deviceMetrics(deviceId: ID!): [String!]!
```

### Mutations

```graphql
# Authentification
register(input: RegisterInput!): AuthPayload!
login(input: LoginInput!): AuthPayload!

# Devices
createDevice(input: CreateDeviceInput!): Device!
updateDevice(input: UpdateDeviceInput!): Device!
deleteDevice(id: ID!): DeleteResult!
```

### Exemples

**Inscription :**
```graphql
mutation {
  register(input: {
    email: "admin@example.com"
    password: "password123"
    name: "Admin"
  }) {
    token
    user {
      id
      email
      role
    }
  }
}
```

**Connexion :**
```graphql
mutation {
  login(input: {
    email: "admin@example.com"
    password: "password123"
  }) {
    token
  }
}
```

**Créer un device :**
```graphql
mutation {
  createDevice(input: {
    name: "Capteur Température"
    type: "temperature_sensor"
    metadata: [
      { key: "location", value: "salon" }
    ]
  }) {
    id
    name
    status
  }
}
```

**Lister les devices :**
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

**Télémétrie agrégée :**
```graphql
query {
  deviceTelemetryAggregated(
    deviceId: "device-001"
    metricName: "temperature"
    startTime: 1705579200
    endTime: 1705665600
    interval: "1 hour"
  ) {
    bucket
    avg
    min
    max
    count
  }
}
```

## Subscriptions temps réel

L'API Gateway supporte les subscriptions GraphQL via WebSocket pour recevoir des données en temps réel.

### Architecture

```
Data Collector ──► Redis (iot:telemetry:*) ──► API Gateway ──► Client WebSocket
                        Pub/Sub                   Broker
```

1. Le **Data Collector** publie chaque mesure sur Redis après insertion en DB
2. Le **RedisSubscriber** écoute le pattern `iot:telemetry:*`
3. Le **Broker** dispatch les messages aux clients connectés
4. Les clients reçoivent les données via leur subscription WebSocket

### Subscriptions disponibles

```graphql
type Subscription {
  # Télémétrie temps réel d'un device
  telemetryReceived(deviceId: ID!): TelemetryPoint!

  # Updates de devices (à implémenter)
  deviceUpdated: Device!
}
```

### Exemple d'utilisation

**Dans le GraphQL Playground :**

```graphql
subscription {
  telemetryReceived(deviceId: "123e4567-e89b-12d3-a456-426614174000") {
    time
    value
    unit
  }
}
```

**Avec Apollo Client (React) :**

```typescript
import { useSubscription, gql } from '@apollo/client';

const TELEMETRY_SUB = gql`
  subscription TelemetryStream($deviceId: ID!) {
    telemetryReceived(deviceId: $deviceId) {
      time
      value
      unit
    }
  }
`;

function DeviceTelemetry({ deviceId }) {
  const { data, loading } = useSubscription(TELEMETRY_SUB, {
    variables: { deviceId },
  });

  if (loading) return <p>En attente de données...</p>;

  return (
    <div>
      <p>Valeur: {data.telemetryReceived.value}</p>
      <p>Unité: {data.telemetryReceived.unit}</p>
    </div>
  );
}
```

### Format des données

Chaque événement `telemetryReceived` contient :

```json
{
  "time": 1705579200,
  "value": 23.5,
  "unit": "°C"
}
```

## Développement

### Modifier le schéma GraphQL

1. Éditer `schema.graphql`
2. Régénérer le code :
   ```bash
   go run github.com/99designs/gqlgen generate
   ```
3. Implémenter les nouveaux resolvers dans `graph/`

### Tests

```bash
go test -v ./...
```

### Health Check

```bash
curl http://localhost:8080/health
```

## License

MIT
