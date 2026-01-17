# Plan d'Évolution - Plateforme IoT

## État Actuel (Baseline)

### ✅ Ce qui est implémenté

| Composant | Status | Détails |
|-----------|--------|---------|
| **API Gateway** | ✅ Fonctionnel | GraphQL, JWT auth, CORS |
| **Device Manager** | ✅ Fonctionnel | CRUD gRPC, PostgreSQL/Memory |
| **User Service** | ✅ Fonctionnel | Auth, bcrypt, roles |
| **Frontend** | ✅ Basique | React, devices list/detail |
| **Infrastructure** | ✅ Configurée | Docker Compose complet |

### ❌ Ce qui manque pour une vraie plateforme IoT

- Pas d'ingestion de télémétrie (données capteurs)
- MQTT broker configuré mais non connecté aux services
- TimescaleDB présent mais inutilisé
- Pas de temps réel (subscriptions non implémentées)
- Pas de système d'alertes

---

## Phase 1 : Télémétrie & MQTT (Cœur IoT)

> **Objectif** : Permettre aux devices d'envoyer des données et les stocker

### 1.1 Schema Base de Données - Télémétrie

**Fichier** : `infrastructure/database/migrations/003_create_telemetry_tables.sql`

```sql
-- Activer TimescaleDB
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Table principale de télémétrie (hypertable)
CREATE TABLE device_telemetry (
    time        TIMESTAMPTZ NOT NULL,
    device_id   UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    metric_name VARCHAR(100) NOT NULL,
    value       DOUBLE PRECISION NOT NULL,
    unit        VARCHAR(50),
    metadata    JSONB DEFAULT '{}'::jsonb
);

-- Convertir en hypertable TimescaleDB (partitionnement automatique par temps)
SELECT create_hypertable('device_telemetry', 'time');

-- Index pour requêtes fréquentes
CREATE INDEX idx_telemetry_device_time ON device_telemetry(device_id, time DESC);
CREATE INDEX idx_telemetry_metric ON device_telemetry(metric_name, time DESC);

-- Politique de rétention : supprimer données > 90 jours
SELECT add_retention_policy('device_telemetry', INTERVAL '90 days');

-- Agrégations continues (moyennes par heure)
CREATE MATERIALIZED VIEW telemetry_hourly
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', time) AS bucket,
    device_id,
    metric_name,
    AVG(value) as avg_value,
    MIN(value) as min_value,
    MAX(value) as max_value,
    COUNT(*) as sample_count
FROM device_telemetry
GROUP BY bucket, device_id, metric_name;

-- Rafraîchir automatiquement les agrégations
SELECT add_continuous_aggregate_policy('telemetry_hourly',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour');
```

### 1.2 Nouveau Service : Telemetry Collector

**Structure** : `services/telemetry-collector/`

```
services/telemetry-collector/
├── main.go                 # Point d'entrée, connexion MQTT
├── mqtt/
│   ├── client.go          # Client MQTT (paho)
│   └── handlers.go        # Handlers par topic
├── storage/
│   ├── storage.go         # Interface
│   └── timescale.go       # Implémentation TimescaleDB
├── proto/
│   └── telemetry.proto    # Messages gRPC (pour API Gateway)
└── go.mod
```

**Fonctionnalités** :
- [ ] Connexion au broker Mosquitto
- [ ] Souscription aux topics : `devices/+/telemetry`
- [ ] Parsing des messages JSON
- [ ] Validation des données
- [ ] Insertion batch dans TimescaleDB
- [ ] Exposition gRPC pour l'API Gateway

**Format message MQTT attendu** :
```json
{
  "device_id": "uuid",
  "timestamp": "2024-01-15T10:30:00Z",
  "metrics": [
    {"name": "temperature", "value": 23.5, "unit": "°C"},
    {"name": "humidity", "value": 65.2, "unit": "%"}
  ]
}
```

### 1.3 Topics MQTT Convention

```
devices/{device_id}/telemetry    # Données capteurs (device → platform)
devices/{device_id}/status       # Changement d'état (device → platform)
devices/{device_id}/commands     # Commandes (platform → device)
devices/{device_id}/config       # Configuration (platform → device)
```

### 1.4 API GraphQL - Queries Télémétrie

**Ajouts au schema** :

```graphql
# Types
type TelemetryPoint {
  time: String!
  value: Float!
  unit: String
}

type TelemetrySeries {
  metricName: String!
  points: [TelemetryPoint!]!
}

type TelemetryAggregation {
  bucket: String!
  avg: Float!
  min: Float!
  max: Float!
  count: Int!
}

# Queries
extend type Query {
  # Données brutes d'un device
  deviceTelemetry(
    deviceId: ID!
    metricName: String!
    from: String!
    to: String!
    limit: Int = 1000
  ): TelemetrySeries!

  # Données agrégées (pour graphiques)
  deviceTelemetryAggregated(
    deviceId: ID!
    metricName: String!
    from: String!
    to: String!
    interval: String! # "1 minute", "1 hour", "1 day"
  ): [TelemetryAggregation!]!

  # Dernière valeur d'une métrique
  deviceLatestMetric(
    deviceId: ID!
    metricName: String!
  ): TelemetryPoint

  # Liste des métriques disponibles pour un device
  deviceMetrics(deviceId: ID!): [String!]!
}
```

### 1.5 Livrables Phase 1

- [ ] Migration SQL `003_create_telemetry_tables.sql`
- [ ] Service `telemetry-collector` complet
- [ ] Proto `telemetry.proto` + génération code
- [ ] Client gRPC dans API Gateway
- [ ] Resolvers GraphQL télémétrie
- [ ] Tests E2E ingestion MQTT → Query GraphQL
- [ ] Docker Compose mis à jour
- [ ] Script de simulation de devices (pour tests)

---

## Phase 2 : Temps Réel (Streaming)

> **Objectif** : Push des données vers les clients en temps réel

### 2.1 GraphQL Subscriptions

**Implémentation WebSocket** dans l'API Gateway :

```go
// Utiliser gorilla/websocket avec gqlgen
// Transport: graphql-ws protocol
```

**Nouveaux subscriptions** :

```graphql
extend type Subscription {
  # Mise à jour d'un device spécifique
  deviceUpdated(deviceId: ID): Device!

  # Flux de télémétrie en temps réel
  telemetryReceived(deviceId: ID!): TelemetryPoint!

  # Nouvelles alertes
  alertTriggered: Alert!
}
```

### 2.2 Pub/Sub avec Redis

**Architecture** :
```
[Telemetry Collector] --publish--> [Redis Pub/Sub] --subscribe--> [API Gateway]
                                                                       |
                                                                       v
                                                               [WebSocket clients]
```

**Channels Redis** :
```
iot:telemetry:{device_id}     # Données temps réel par device
iot:device:status             # Changements de statut
iot:alerts                    # Alertes système
```

### 2.3 gRPC Streaming

**Implémenter `WatchDevices`** dans device-manager :

```protobuf
service DeviceService {
  // ... existing ...

  // Stream de changements de devices
  rpc WatchDevices(WatchDevicesRequest) returns (stream DeviceEvent);
}

message WatchDevicesRequest {
  repeated string device_ids = 1; // vide = tous
}

message DeviceEvent {
  enum EventType {
    CREATED = 0;
    UPDATED = 1;
    DELETED = 2;
    STATUS_CHANGED = 3;
  }
  EventType type = 1;
  Device device = 2;
  int64 timestamp = 3;
}
```

### 2.4 Livrables Phase 2

- [ ] WebSocket transport dans API Gateway
- [ ] Subscription `deviceUpdated` fonctionnel
- [ ] Subscription `telemetryReceived` fonctionnel
- [ ] Redis Pub/Sub intégration
- [ ] gRPC `WatchDevices` stream
- [ ] Tests de charge WebSocket
- [ ] Documentation protocole temps réel

---

## Phase 3 : Alertes & Notifications

> **Objectif** : Réagir aux événements et notifier les utilisateurs

### 3.1 Schema Base de Données - Alertes

**Fichier** : `infrastructure/database/migrations/004_create_alerts_tables.sql`

```sql
-- Types d'alertes
CREATE TYPE alert_severity AS ENUM ('INFO', 'WARNING', 'CRITICAL');
CREATE TYPE alert_status AS ENUM ('ACTIVE', 'ACKNOWLEDGED', 'RESOLVED');

-- Règles d'alertes (définies par les utilisateurs)
CREATE TABLE alert_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    device_id UUID REFERENCES devices(id) ON DELETE CASCADE,
    device_type VARCHAR(100),  -- OU par type de device
    metric_name VARCHAR(100) NOT NULL,
    condition VARCHAR(20) NOT NULL,  -- 'gt', 'lt', 'eq', 'gte', 'lte'
    threshold DOUBLE PRECISION NOT NULL,
    severity alert_severity NOT NULL DEFAULT 'WARNING',
    enabled BOOLEAN NOT NULL DEFAULT true,
    cooldown_minutes INT NOT NULL DEFAULT 5,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Alertes déclenchées
CREATE TABLE alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID REFERENCES alert_rules(id) ON DELETE SET NULL,
    device_id UUID REFERENCES devices(id) ON DELETE CASCADE,
    severity alert_severity NOT NULL,
    status alert_status NOT NULL DEFAULT 'ACTIVE',
    title VARCHAR(255) NOT NULL,
    message TEXT,
    metric_name VARCHAR(100),
    metric_value DOUBLE PRECISION,
    threshold DOUBLE PRECISION,
    triggered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    acknowledged_at TIMESTAMPTZ,
    acknowledged_by UUID REFERENCES users(id),
    resolved_at TIMESTAMPTZ,
    resolved_by UUID REFERENCES users(id)
);

-- Index
CREATE INDEX idx_alerts_device ON alerts(device_id, triggered_at DESC);
CREATE INDEX idx_alerts_status ON alerts(status, triggered_at DESC);
CREATE INDEX idx_alert_rules_device ON alert_rules(device_id);
CREATE INDEX idx_alert_rules_enabled ON alert_rules(enabled);
```

### 3.2 Nouveau Service : Alert Manager

**Structure** : `services/alert-manager/`

```
services/alert-manager/
├── main.go
├── engine/
│   ├── evaluator.go       # Évaluation des règles
│   └── processor.go       # Traitement des alertes
├── notifier/
│   ├── notifier.go        # Interface
│   ├── email.go           # Notifications email
│   └── webhook.go         # Webhooks
├── storage/
│   └── postgres.go
├── proto/
│   └── alert.proto
└── go.mod
```

**Fonctionnement** :
1. S'abonne à Redis `iot:telemetry:*`
2. Évalue les règles en mémoire (chargées au démarrage)
3. Détecte les violations de seuils
4. Crée les alertes en BDD
5. Publie sur Redis `iot:alerts`
6. Envoie notifications (email, webhook)

### 3.3 API GraphQL - Alertes

```graphql
type Alert {
  id: ID!
  deviceId: ID!
  device: Device
  severity: AlertSeverity!
  status: AlertStatus!
  title: String!
  message: String
  metricName: String
  metricValue: Float
  threshold: Float
  triggeredAt: String!
  acknowledgedAt: String
  resolvedAt: String
}

type AlertRule {
  id: ID!
  name: String!
  description: String
  deviceId: ID
  deviceType: String
  metricName: String!
  condition: String!
  threshold: Float!
  severity: AlertSeverity!
  enabled: Boolean!
  cooldownMinutes: Int!
}

enum AlertSeverity {
  INFO
  WARNING
  CRITICAL
}

enum AlertStatus {
  ACTIVE
  ACKNOWLEDGED
  RESOLVED
}

extend type Query {
  # Alertes actives
  activeAlerts(deviceId: ID, severity: AlertSeverity): [Alert!]!

  # Historique des alertes
  alertHistory(
    deviceId: ID
    from: String!
    to: String!
    limit: Int = 100
  ): [Alert!]!

  # Règles d'alertes
  alertRules(deviceId: ID): [AlertRule!]!
}

extend type Mutation {
  # Créer une règle d'alerte
  createAlertRule(input: CreateAlertRuleInput!): AlertRule!

  # Modifier une règle
  updateAlertRule(input: UpdateAlertRuleInput!): AlertRule!

  # Supprimer une règle
  deleteAlertRule(id: ID!): DeleteResult!

  # Acquitter une alerte
  acknowledgeAlert(id: ID!): Alert!

  # Résoudre une alerte
  resolveAlert(id: ID!, message: String): Alert!
}
```

### 3.4 Livrables Phase 3

- [ ] Migration SQL `004_create_alerts_tables.sql`
- [ ] Service `alert-manager` complet
- [ ] Moteur d'évaluation des règles
- [ ] Intégration Redis Pub/Sub
- [ ] Proto `alert.proto` + génération
- [ ] Resolvers GraphQL alertes
- [ ] Notifier email (SMTP)
- [ ] Notifier webhook
- [ ] Tests E2E alerting
- [ ] Dashboard alertes (frontend - optionnel)

---

## Phase 4 : Commandes & Contrôle

> **Objectif** : Envoyer des commandes aux devices

### 4.1 Bidirectional MQTT

**Nouveau topic** :
```
devices/{device_id}/commands     # Platform → Device
devices/{device_id}/commands/ack # Device → Platform (accusé)
```

**Format commande** :
```json
{
  "command_id": "uuid",
  "action": "set_config",
  "params": {
    "interval": 30
  },
  "timestamp": "2024-01-15T10:30:00Z",
  "expires_at": "2024-01-15T10:35:00Z"
}
```

### 4.2 API GraphQL - Commandes

```graphql
type Command {
  id: ID!
  deviceId: ID!
  action: String!
  params: JSON
  status: CommandStatus!
  sentAt: String!
  acknowledgedAt: String
  executedAt: String
  error: String
}

enum CommandStatus {
  PENDING
  SENT
  ACKNOWLEDGED
  EXECUTED
  FAILED
  EXPIRED
}

extend type Mutation {
  # Envoyer une commande à un device
  sendCommand(input: SendCommandInput!): Command!
}

extend type Subscription {
  # Suivi d'une commande
  commandStatusChanged(commandId: ID!): Command!
}
```

### 4.3 Livrables Phase 4

- [ ] Table `device_commands` en BDD
- [ ] Publication MQTT depuis API Gateway
- [ ] Tracking statut des commandes
- [ ] Timeout et expiration
- [ ] Tests avec device simulé

---

## Phase 5 : Production Ready

> **Objectif** : Préparer pour la production

### 5.1 Sécurité

- [ ] **TLS partout**
  - Certificats Let's Encrypt ou auto-signés
  - mTLS entre services gRPC
  - MQTT over TLS (port 8883)

- [ ] **Secrets management**
  - Variables d'environnement via Kubernetes Secrets
  - Rotation des JWT secrets
  - Pas de secrets en dur

- [ ] **MQTT Auth**
  - Authentification devices par certificat client
  - ACL par device (ne peut publier que sur ses topics)

- [ ] **Rate limiting**
  - API Gateway : 100 req/s par IP
  - MQTT : 10 msg/s par device

### 5.2 Observabilité

- [ ] **Métriques Prometheus**
  ```go
  // Exposer /metrics sur chaque service
  - http_requests_total
  - http_request_duration_seconds
  - grpc_server_handled_total
  - mqtt_messages_received_total
  - telemetry_points_ingested_total
  - alerts_triggered_total
  ```

- [ ] **Dashboards Grafana**
  - Santé des services
  - Throughput MQTT
  - Latence API
  - Alertes actives

- [ ] **Logging structuré**
  - Format JSON
  - Correlation IDs
  - Niveaux : DEBUG, INFO, WARN, ERROR

- [ ] **Tracing (optionnel)**
  - OpenTelemetry
  - Jaeger pour visualisation

### 5.3 Résilience

- [ ] **Health checks**
  - `/health/live` - Le process tourne
  - `/health/ready` - Prêt à servir (BDD connectée, etc.)

- [ ] **Graceful shutdown**
  - Drain des connexions
  - Flush des buffers

- [ ] **Circuit breakers**
  - Entre services gRPC
  - Vers la base de données

- [ ] **Retry policies**
  - Exponential backoff
  - Idempotency keys

### 5.4 Livrables Phase 5

- [ ] Configuration TLS tous services
- [ ] Authentification MQTT
- [ ] Métriques Prometheus exposées
- [ ] Dashboards Grafana
- [ ] Health checks endpoints
- [ ] Graceful shutdown
- [ ] Documentation déploiement

---

## Résumé des Phases

| Phase | Focus | Services impactés | Priorité |
|-------|-------|-------------------|----------|
| **1** | Télémétrie & MQTT | Nouveau: telemetry-collector | 🔴 Critique |
| **2** | Temps réel | API Gateway, tous | 🟠 Haute |
| **3** | Alertes | Nouveau: alert-manager | 🟡 Moyenne |
| **4** | Commandes | API Gateway, MQTT | 🟢 Normale |
| **5** | Production | Tous | 🔴 Critique (avant prod) |

---

## Ordre d'implémentation recommandé

```
Phase 1.1 → Migration télémétrie
Phase 1.2 → Telemetry collector (MQTT → DB)
Phase 1.3 → API GraphQL télémétrie
Phase 2.1 → Redis Pub/Sub
Phase 2.2 → GraphQL Subscriptions
Phase 3.1 → Migration alertes
Phase 3.2 → Alert manager
Phase 4   → Commandes (si besoin)
Phase 5   → Hardening production
```

---

## Prochaine étape

Commencer par **Phase 1.1** : créer la migration `003_create_telemetry_tables.sql` et valider le schema TimescaleDB.
