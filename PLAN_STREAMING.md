# Plan d'Attaque - Streaming Temps Réel

## État Actuel

| Composant | État | Notes |
|-----------|------|-------|
| Redis | ✅ Configuré | Dans docker-compose, mais non utilisé |
| GraphQL Subscription | 🟡 Déclaré | `deviceUpdated` existe mais `panic("not implemented")` |
| WebSocket (gorilla) | 🟡 Dépendance présente | Non utilisé dans le serveur |
| Apollo Client | 🟡 Partiel | Pas de WebSocketLink |
| Data Collector | ✅ Fonctionne | MQTT → TimescaleDB, pas de Redis |

---

## Architecture Cible

```
┌─────────────────┐     MQTT      ┌─────────────────────┐
│  IoT Devices    │──────────────▶│ Data Collector │
└─────────────────┘               └──────────┬──────────┘
                                             │
                                   ┌─────────▼─────────┐
                                   │   TimescaleDB     │
                                   └───────────────────┘
                                             │
                                   ┌─────────▼─────────┐
                                   │   Redis Pub/Sub   │
                                   │                   │
                                   │ Channels:         │
                                   │ • iot:telemetry:* │
                                   │ • iot:device:*    │
                                   └─────────┬─────────┘
                                             │
                                   ┌─────────▼─────────┐
                                   │   API Gateway     │
                                   │   (WebSocket)     │
                                   └─────────┬─────────┘
                                             │
                                   ┌─────────▼─────────┐
                                   │   Frontend        │
                                   │   (Apollo WS)     │
                                   └───────────────────┘
```

---

## Étapes d'Implémentation

### Étape 1 : Redis Pub/Sub dans Telemetry Collector

**Fichiers à modifier :**
- `services/Data-collector/main.go`
- `services/Data-collector/publisher/redis.go` (nouveau)

**Travail :**
1. Ajouter dépendance `github.com/redis/go-redis/v9`
2. Créer un publisher Redis
3. Après chaque insertion en BDD, publier sur Redis :
   - Channel : `iot:telemetry:{device_id}`
   - Payload : JSON avec device_id, metric_name, value, timestamp

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

---

### Étape 2 : WebSocket Transport dans API Gateway

**Fichiers à modifier :**
- `services/api-gateway/main.go`
- `services/api-gateway/gqlgen.yml` (si besoin)

**Travail :**
1. Configurer le transport WebSocket avec gqlgen
2. Utiliser `github.com/gorilla/websocket` (déjà présent)
3. Ajouter le handler WebSocket sur `/query` (même endpoint)
4. Configurer le protocole `graphql-transport-ws`

**Code principal :**
```go
// main.go
import "github.com/99designs/gqlgen/graphql/handler/transport"

srv := handler.NewDefaultServer(generated.NewExecutableSchema(cfg))

// Ajouter WebSocket transport
srv.AddTransport(&transport.Websocket{
    Upgrader: websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool { return true },
    },
    KeepAlivePingInterval: 10 * time.Second,
})
```

---

### Étape 3 : Subscriber Redis dans API Gateway

**Fichiers à créer :**
- `services/api-gateway/pubsub/redis.go`
- `services/api-gateway/pubsub/broker.go`

**Travail :**
1. Créer un broker qui s'abonne aux channels Redis
2. Maintenir une map de subscribers (par device_id)
3. Quand un message arrive sur Redis → dispatcher aux subscribers GraphQL

**Architecture interne :**
```go
type Broker struct {
    redis       *redis.Client
    subscribers map[string][]chan *TelemetryPoint  // device_id -> channels
    mu          sync.RWMutex
}

func (b *Broker) Subscribe(deviceID string) <-chan *TelemetryPoint
func (b *Broker) Unsubscribe(deviceID string, ch <-chan *TelemetryPoint)
```

---

### Étape 4 : Implémenter les Resolvers de Subscription

**Fichiers à modifier :**
- `services/api-gateway/graph/schema.resolvers.go`

**Subscriptions à implémenter :**

```graphql
type Subscription {
  # Déjà déclaré - à implémenter
  deviceUpdated: Device!

  # À ajouter au schema
  telemetryReceived(deviceId: ID!): TelemetryPoint!
}
```

**Code resolver :**
```go
func (r *subscriptionResolver) TelemetryReceived(ctx context.Context, deviceID string) (<-chan *model.TelemetryPoint, error) {
    ch := r.broker.Subscribe(deviceID)

    go func() {
        <-ctx.Done()
        r.broker.Unsubscribe(deviceID, ch)
    }()

    return ch, nil
}
```

---

### Étape 5 : Configurer Apollo Client (Frontend)

**Fichiers à modifier :**
- `frontends/dashboard/src/lib/apollo-client.ts`

**Travail :**
1. Installer `graphql-ws` : `npm install graphql-ws`
2. Créer un WebSocketLink
3. Split : HTTP pour queries/mutations, WS pour subscriptions

**Code :**
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

---

### Étape 6 : Ajouter les Queries GraphQL (Frontend)

**Fichiers à modifier :**
- `frontends/dashboard/src/graphql/queries.ts`

**Subscriptions à ajouter :**
```graphql
subscription TelemetryStream($deviceId: ID!) {
  telemetryReceived(deviceId: $deviceId) {
    time
    value
    unit
  }
}
```

---

### Étape 7 : Tests E2E

**Scénario de test :**
1. Démarrer tous les services (`docker-compose up`)
2. Ouvrir le frontend, se connecter
3. Souscrire à un device
4. Publier un message MQTT simulé
5. Vérifier que le frontend reçoit les données en temps réel

**Script de test MQTT :**
```bash
mosquitto_pub -h localhost -t "devices/<device_id>/telemetry" -m '{
  "device_id": "<device_id>",
  "timestamp": "2024-01-15T10:30:00Z",
  "metrics": [{"name": "temperature", "value": 25.5, "unit": "°C"}]
}'
```

---

## Checklist des Livrables

### Backend
- [ ] Publisher Redis dans telemetry-collector
- [ ] WebSocket transport dans API Gateway
- [ ] Subscriber Redis dans API Gateway
- [ ] Broker de subscriptions (in-memory)
- [ ] Resolver `telemetryReceived`
- [ ] Resolver `deviceUpdated` (optionnel, Phase 2.3)

### Frontend
- [ ] Installer `graphql-ws`
- [ ] Configurer WebSocketLink dans Apollo
- [ ] Hook `useTelemetryStream(deviceId)`

### Tests
- [ ] Test unitaire du broker
- [ ] Test E2E MQTT → Frontend

---

## Ordre d'Exécution

```
1. Telemetry Collector + Redis Publisher     ←── Commencer ici
2. API Gateway + WebSocket Transport
3. API Gateway + Redis Subscriber + Broker
4. API Gateway + Subscription Resolvers
5. Frontend + Apollo WebSocket
6. Tests E2E
```

---

## Estimation de Complexité

| Étape | Fichiers | Complexité |
|-------|----------|------------|
| 1. Redis Publisher | 2 | 🟢 Faible |
| 2. WebSocket Transport | 1 | 🟢 Faible |
| 3. Redis Subscriber | 2 | 🟡 Moyenne |
| 4. Subscription Resolvers | 2 | 🟡 Moyenne |
| 5. Frontend Apollo WS | 2 | 🟢 Faible |
| 6. Tests | 1-2 | 🟢 Faible |

**Total : ~10 fichiers à créer/modifier**

---

## Prochaine Action

Commencer par **Étape 1** : Ajouter le publisher Redis dans le telemetry-collector.
