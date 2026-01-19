# Streaming Temps Réel - Documentation

## État Actuel

| Composant | État | Notes |
|-----------|------|-------|
| Redis Pub/Sub | ✅ Implémenté | Data Collector publie sur `iot:telemetry:{device_id}` |
| GraphQL Subscription | ✅ Implémenté | `telemetryReceived(deviceId)` fonctionne |
| WebSocket (gorilla) | ✅ Configuré | Transport WebSocket actif sur `/query` |
| Apollo Client | 🟡 À faire | Étape 5 - Frontend |
| Data Collector | ✅ Complet | MQTT → TimescaleDB → Redis |

---

## Architecture Implémentée

```
┌─────────────────┐     MQTT      ┌─────────────────────┐
│  IoT Devices    │──────────────▶│   Data Collector    │
└─────────────────┘               └──────────┬──────────┘
                                             │
                                   ┌─────────▼─────────┐
                                   │   TimescaleDB     │
                                   │   (stockage)      │
                                   └───────────────────┘
                                             │
                                   ┌─────────▼─────────┐
                                   │   Redis Pub/Sub   │
                                   │                   │
                                   │ Channel:          │
                                   │ iot:telemetry:*   │
                                   └─────────┬─────────┘
                                             │
                                   ┌─────────▼─────────┐
                                   │   API Gateway     │
                                   │   (WebSocket)     │
                                   │                   │
                                   │ • RedisSubscriber │
                                   │ • Broker          │
                                   └─────────┬─────────┘
                                             │
                                   ┌─────────▼─────────┐
                                   │   Clients         │
                                   │   (GraphQL WS)    │
                                   └───────────────────┘
```

---

## Composants Backend

### 1. Data Collector - Redis Publisher

**Fichiers :**
- `services/data-collector/publisher/redis.go`
- `services/data-collector/main.go`

**Fonctionnement :**
1. Reçoit les données MQTT des devices
2. Insère dans TimescaleDB
3. Publie sur Redis après insertion réussie

**Format du message Redis :**
```json
{
  "device_id": "uuid",
  "metric_name": "temperature",
  "value": 23.5,
  "unit": "°C",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Channel Redis :** `iot:telemetry:{device_id}`

---

### 2. API Gateway - WebSocket Transport

**Fichier :** `services/api-gateway/main.go`

**Configuration :**
```go
srv.AddTransport(&transport.Websocket{
    Upgrader: websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool { return true },
    },
    KeepAlivePingInterval: 10 * time.Second,
    InitFunc: func(ctx context.Context, initPayload transport.InitPayload) (context.Context, *transport.InitPayload, error) {
        // Auth JWT via connectionParams
        token := initPayload.Authorization()
        if token != "" {
            claims, err := jwtManager.ValidateToken(token)
            if err == nil {
                ctx = auth.WithUser(ctx, claims)
            }
        }
        return ctx, &initPayload, nil
    },
})
```

---

### 3. API Gateway - Broker & Redis Subscriber

**Fichiers :**
- `services/api-gateway/pubsub/broker.go` - Gestion des subscriptions in-memory
- `services/api-gateway/pubsub/redis.go` - Écoute Redis et dispatch

**Broker :**
```go
type Broker struct {
    subscribers map[string]map[chan *model.TelemetryPoint]struct{}
    mu          sync.RWMutex
}

func (b *Broker) Subscribe(deviceID string) chan *model.TelemetryPoint
func (b *Broker) Unsubscribe(deviceID string, ch chan *model.TelemetryPoint)
func (b *Broker) Publish(deviceID string, point *model.TelemetryPoint)
```

**RedisSubscriber :**
- S'abonne au pattern `iot:telemetry:*`
- Parse les messages JSON
- Dispatch via le Broker

---

### 4. GraphQL Subscription Resolver

**Fichier :** `services/api-gateway/graph/schema.resolvers.go`

**Schema GraphQL :**
```graphql
type Subscription {
  deviceUpdated: Device!
  telemetryReceived(deviceId: ID!): TelemetryPoint!
}
```

**Resolver :**
```go
func (r *subscriptionResolver) TelemetryReceived(ctx context.Context, deviceID string) (<-chan *model.TelemetryPoint, error) {
    ch := r.Broker.Subscribe(deviceID)

    go func() {
        <-ctx.Done()
        r.Broker.Unsubscribe(deviceID, ch)
    }()

    return ch, nil
}
```

---

## Configuration

### Variables d'environnement

**Data Collector :**
```env
REDIS_HOST=redis      # ou localhost
REDIS_PORT=6379
REDIS_PASSWORD=       # optionnel
REDIS_DB=0            # optionnel
```

**API Gateway :**
```env
REDIS_HOST=redis      # ou localhost
REDIS_PORT=6379
```

### Docker Compose

Les deux services (`data-collector` et `api-gateway`) dépendent maintenant de Redis :
```yaml
depends_on:
  redis:
    condition: service_healthy
```

---

## Test du Streaming

### 1. Démarrer les services
```bash
docker-compose up -d
```

### 2. Lancer le simulateur
```bash
make simulate
```

### 3. Vérifier Redis (optionnel)
```bash
docker exec -it iot-redis redis-cli PSUBSCRIBE 'iot:telemetry:*'
```

### 4. Tester via GraphQL Playground

1. Ouvrir http://localhost:8080/
2. Créer un compte ou se connecter
3. Récupérer un device ID :
```graphql
query {
  devices {
    devices { id name }
  }
}
```

4. Lancer la subscription :
```graphql
subscription {
  telemetryReceived(deviceId: "<device_id>") {
    time
    value
    unit
  }
}
```

---

## Prochaines Étapes

### Étape 5 : Frontend Apollo WebSocket

**Fichiers à modifier :**
- `frontends/dashboard/src/lib/apollo-client.ts`

**Installation :**
```bash
npm install graphql-ws
```

**Configuration :**
```typescript
import { GraphQLWsLink } from '@apollo/client/link/subscriptions';
import { createClient } from 'graphql-ws';
import { split, HttpLink } from '@apollo/client';
import { getMainDefinition } from '@apollo/client/utilities';

const httpLink = new HttpLink({ uri: 'http://localhost:8080/query' });

const wsLink = new GraphQLWsLink(createClient({
  url: 'ws://localhost:8080/query',
  connectionParams: () => ({
    authorization: localStorage.getItem('token') || '',
  }),
}));

const splitLink = split(
  ({ query }) => {
    const def = getMainDefinition(query);
    return def.kind === 'OperationDefinition' && def.operation === 'subscription';
  },
  wsLink,
  httpLink,
);
```

### Étape 6 : Hook React pour le streaming

```typescript
// hooks/useTelemetryStream.ts
import { useSubscription, gql } from '@apollo/client';

const TELEMETRY_SUBSCRIPTION = gql`
  subscription TelemetryStream($deviceId: ID!) {
    telemetryReceived(deviceId: $deviceId) {
      time
      value
      unit
    }
  }
`;

export function useTelemetryStream(deviceId: string) {
  return useSubscription(TELEMETRY_SUBSCRIPTION, {
    variables: { deviceId },
  });
}
```

---

## Checklist

### Backend ✅
- [x] Publisher Redis dans data-collector
- [x] WebSocket transport dans API Gateway
- [x] Subscriber Redis dans API Gateway
- [x] Broker de subscriptions (in-memory)
- [x] Resolver `telemetryReceived`
- [ ] Resolver `deviceUpdated` (optionnel)

### Frontend 🟡
- [ ] Installer `graphql-ws`
- [ ] Configurer WebSocketLink dans Apollo
- [ ] Hook `useTelemetryStream(deviceId)`
- [ ] Composant de visualisation temps réel

### Tests
- [ ] Test unitaire du broker
- [ ] Test E2E MQTT → Frontend
