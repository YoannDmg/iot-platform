# Telemetry Collector

> Microservice de collecte de télémétrie IoT via MQTT avec stockage TimescaleDB

[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go&logoColor=white)](https://golang.org)
[![gRPC](https://img.shields.io/badge/gRPC-HTTP%2F2-4285F4)](https://grpc.io)
[![MQTT](https://img.shields.io/badge/MQTT-3.1.1-660066?logo=mqtt&logoColor=white)](https://mqtt.org)
[![TimescaleDB](https://img.shields.io/badge/TimescaleDB-FDB515?logo=timescale&logoColor=black)](https://timescale.com)

## Table des matières

- [Vue d'ensemble](#vue-densemble)
- [Architecture](#architecture)
- [Démarrage rapide](#démarrage-rapide)
- [Configuration](#configuration)
- [MQTT](#mqtt)
- [API gRPC](#api-grpc)
- [Base de données](#base-de-données)

## Vue d'ensemble

Le Telemetry Collector ingère les données de télémétrie des appareils IoT via MQTT et les stocke dans TimescaleDB. Il expose une API gRPC pour interroger les données brutes et agrégées.

### Fonctionnalités

- **Ingestion MQTT** — Souscription aux topics des devices, parsing JSON
- **Stockage time-series** — TimescaleDB avec hypertables optimisées
- **Agrégations** — Moyennes, min, max par intervalles configurables
- **Cache** — Table de cache pour les dernières valeurs
- **Batch insert** — Insertion par lots pour les hauts débits

### Technologies

| Composant | Technologie |
|-----------|-------------|
| Langage | Go 1.24 |
| Protocol | gRPC (HTTP/2) |
| Messaging | MQTT (Paho) |
| Database | PostgreSQL + TimescaleDB |
| Driver | pgx/v5 avec connection pooling |

## Architecture

```
┌─────────────┐    MQTT     ┌──────────────┐
│ IoT Devices │────────────►│ MQTT Broker  │
└─────────────┘             │ (Mosquitto)  │
                            └──────┬───────┘
                                   │ subscribe
                                   ▼
┌─────────────────────────────────────────────┐
│           Telemetry Collector               │
│              Port 8083                      │
├─────────────────────────────────────────────┤
│  MQTT Client          │    gRPC Server      │
│  - Subscribe          │    - GetTelemetry   │
│  - Parse JSON         │    - GetAggregated  │
│  - Extract device_id  │    - GetLatest      │
└───────────┬───────────┴─────────────────────┘
            │
            ▼
┌─────────────────────────────────────────────┐
│              TimescaleDB                    │
│  ┌─────────────────┐  ┌──────────────────┐  │
│  │device_telemetry │  │device_telemetry_ │  │
│  │  (hypertable)   │  │     latest       │  │
│  └─────────────────┘  └──────────────────┘  │
└─────────────────────────────────────────────┘
```

### Structure du projet

```
telemetry-collector/
├── main.go              # Point d'entrée, serveur gRPC
├── mqtt/
│   └── client.go        # Client MQTT, parsing messages
├── storage/
│   ├── storage.go       # Interface Storage
│   └── timescale.go     # Implémentation TimescaleDB
├── Dockerfile
└── go.mod
```

## Démarrage rapide

### Prérequis

- Go 1.24+
- Docker (pour MQTT et TimescaleDB)
- Broker MQTT (Mosquitto)

### Lancement

```bash
# Depuis la racine du projet
make infra          # Démarre Mosquitto + PostgreSQL
make db-migrate     # Crée les tables
make dev-telemetry  # Lance le service

# Ou directement
cd services/telemetry-collector
go run main.go
```

Le service :
- Écoute sur `localhost:8083` (gRPC)
- Se connecte au broker MQTT sur `localhost:1883`
- Souscrit au topic `devices/+/telemetry`

## Configuration

### Variables d'environnement

| Variable | Description | Défaut |
|----------|-------------|--------|
| `TELEMETRY_GRPC_PORT` | Port gRPC | `8083` |
| `MQTT_BROKER` | URL du broker | `tcp://localhost:1883` |
| `MQTT_CLIENT_ID` | ID client MQTT | `telemetry-collector` |
| `MQTT_TOPIC` | Topic de souscription | `devices/+/telemetry` |
| `DB_HOST` | Hôte PostgreSQL | `localhost` |
| `DB_PORT` | Port PostgreSQL | `5432` |
| `DB_NAME` | Nom de la base | `iot_platform` |
| `DB_USER` | Utilisateur | `iot_user` |
| `DB_PASSWORD` | Mot de passe | `iot_password` |
| `DB_SSLMODE` | Mode SSL | `disable` |

## MQTT

### Format des messages

Les devices publient sur `devices/{device_id}/telemetry` :

```json
{
  "device_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2026-01-18T12:00:00Z",
  "metrics": [
    {
      "name": "temperature",
      "value": 23.5,
      "unit": "°C",
      "metadata": {
        "location": "room-1",
        "sensor": "DHT22"
      }
    },
    {
      "name": "humidity",
      "value": 45.2,
      "unit": "%"
    }
  ]
}
```

| Champ | Requis | Description |
|-------|--------|-------------|
| `device_id` | Oui | UUID du device (ou extrait du topic) |
| `timestamp` | Non | ISO 8601 (défaut: maintenant) |
| `metrics` | Oui | Liste des métriques |
| `metrics[].name` | Oui | Nom de la métrique |
| `metrics[].value` | Oui | Valeur numérique |
| `metrics[].unit` | Non | Unité de mesure |
| `metrics[].metadata` | Non | Métadonnées additionnelles |

### Test avec mosquitto_pub

```bash
mosquitto_pub -t "devices/device-001/telemetry" -m '{
  "metrics": [
    {"name": "temperature", "value": 22.5, "unit": "°C"}
  ]
}'
```

## API gRPC

### Service Definition

```protobuf
service TelemetryService {
  rpc GetTelemetry(GetTelemetryRequest) returns (GetTelemetryResponse);
  rpc GetTelemetryAggregated(GetTelemetryAggregatedRequest) returns (GetTelemetryAggregatedResponse);
  rpc GetLatestMetric(GetLatestMetricRequest) returns (GetLatestMetricResponse);
  rpc GetDeviceMetrics(GetDeviceMetricsRequest) returns (GetDeviceMetricsResponse);
}
```

### Exemples avec grpcurl

> Les commandes suivantes doivent être exécutées depuis la **racine du projet**.

**Récupérer les données brutes :**
```bash
grpcurl -plaintext \
  -import-path shared/proto \
  -proto telemetry/telemetry.proto \
  -d '{
    "device_id": "device-001",
    "metric_name": "temperature",
    "start_time": 1705579200,
    "end_time": 1705665600,
    "limit": 100
  }' localhost:8083 telemetry.TelemetryService/GetTelemetry
```

**Récupérer les agrégations :**
```bash
grpcurl -plaintext \
  -import-path shared/proto \
  -proto telemetry/telemetry.proto \
  -d '{
    "device_id": "device-001",
    "metric_name": "temperature",
    "start_time": 1705579200,
    "end_time": 1705665600,
    "interval": "1 hour"
  }' localhost:8083 telemetry.TelemetryService/GetTelemetryAggregated
```

**Dernière valeur :**
```bash
grpcurl -plaintext \
  -import-path shared/proto \
  -proto telemetry/telemetry.proto \
  -d '{
    "device_id": "device-001",
    "metric_name": "temperature"
  }' localhost:8083 telemetry.TelemetryService/GetLatestMetric
```

**Métriques disponibles :**
```bash
grpcurl -plaintext \
  -import-path shared/proto \
  -proto telemetry/telemetry.proto \
  -d '{
    "device_id": "device-001"
  }' localhost:8083 telemetry.TelemetryService/GetDeviceMetrics
```

### Intervalles d'agrégation supportés

- `1 minute`, `5 minutes`, `15 minutes`, `30 minutes`
- `1 hour`, `6 hours`, `12 hours`
- `1 day`, `1 week`

## Base de données

### Schéma TimescaleDB

```sql
-- Table principale (hypertable)
CREATE TABLE device_telemetry (
    time        TIMESTAMPTZ NOT NULL,
    device_id   UUID NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    value       DOUBLE PRECISION NOT NULL,
    unit        VARCHAR(50),
    metadata    JSONB
);

SELECT create_hypertable('device_telemetry', 'time');

-- Cache des dernières valeurs
CREATE TABLE device_telemetry_latest (
    device_id   UUID NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    time        TIMESTAMPTZ NOT NULL,
    value       DOUBLE PRECISION NOT NULL,
    unit        VARCHAR(50),
    PRIMARY KEY (device_id, metric_name)
);
```

### Index

```sql
CREATE INDEX idx_telemetry_device_metric ON device_telemetry(device_id, metric_name, time DESC);
CREATE INDEX idx_telemetry_time ON device_telemetry(time DESC);
```

### Connection Pool

| Paramètre | Valeur |
|-----------|--------|
| Min connections | 5 |
| Max connections | 20 |
| Max lifetime | 1 heure |
| Max idle time | 30 minutes |

## Simulation

Pour tester avec des données simulées :

```bash
# Depuis la racine
make simulate           # 5 devices, intervalle 3s
make simulate-heavy     # 50 devices, stress test
```

## License

MIT