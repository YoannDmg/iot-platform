# Device Manager Service

> Microservice gRPC de gestion des devices IoT avec support PostgreSQL

[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org)
[![gRPC](https://img.shields.io/badge/gRPC-HTTP%2F2-green.svg)](https://grpc.io)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-blue.svg)](https://postgresql.org)

## 📋 Table des matières

- [Vue d'ensemble](#-vue-densemble)
- [Architecture](#-architecture)
- [Démarrage rapide](#-démarrage-rapide)
- [Configuration](#-configuration)
- [API gRPC](#-api-grpc)
- [Base de données](#-base-de-données)
- [Tests](#-tests)
- [Développement](#-développement)

## 🎯 Vue d'ensemble

Le Device Manager est un microservice gRPC responsable de la gestion du cycle de vie complet des devices IoT dans la plateforme. Il offre une API performante et type-safe pour les opérations CRUD, avec support de deux backends de stockage.

### Fonctionnalités

- ✅ **CRUD complet** - Création, lecture, mise à jour, suppression de devices
- ✅ **Gestion des statuts** - ONLINE, OFFLINE, ERROR, MAINTENANCE
- ✅ **Métadonnées flexibles** - Stockage JSONB pour données personnalisées
- ✅ **Dual storage** - Support PostgreSQL (production) et In-Memory (dev/tests)
- ✅ **Type-safe** - Génération de code avec sqlc et Protocol Buffers
- ✅ **Performances** - Driver pgx hautes performances
- ✅ **Pagination** - Listing paginé des devices
- ✅ **Thread-safe** - Accès concurrent sécurisé

### Technologies

- **Langage**: Go 1.24
- **Protocol**: gRPC (HTTP/2)
- **Database**: PostgreSQL 15+ avec TimescaleDB
- **ORM**: sqlc (SQL-first, type-safe)
- **Driver**: pgx/v5 (3-5x plus rapide que lib/pq)
- **Schema**: Protocol Buffers (proto3)

## 🏗️ Architecture

### Architecture globale

```
┌─────────────────┐
│  API Gateway    │ ← GraphQL API (port 8080)
│   (GraphQL)     │
└────────┬────────┘
         │ gRPC
    ┌────▼──────────────────┐
    │  Device Manager       │ ← gRPC Service (port 8081)
    │  Storage Interface    │
    └────────┬──────────────┘
             │
      ┌──────┴───────┐
      │              │
  ┌───▼────┐    ┌───▼──────┐
  │ Memory │    │PostgreSQL│
  │Storage │    │ Storage  │
  └────────┘    └──────────┘
```

### Architecture interne

```
┌─────────────────────────────────────────┐
│         DeviceServer (gRPC)             │
├─────────────────────────────────────────┤
│         Storage Interface               │
│  - CreateDevice(...)                    │
│  - GetDevice(...)                       │
│  - ListDevices(...)                     │
│  - UpdateDevice(...)                    │
│  - DeleteDevice(...)                    │
└──────────┬──────────────────────────────┘
           │
    ┌──────┴──────┐
    │             │
┌───▼─────────┐ ┌─▼──────────────┐
│MemoryStorage│ │PostgresStorage │
│             │ │ ┌────────────┐ │
│ map[string] │ │ │ sqlc       │ │
│ *Device     │ │ │ generated  │ │
│             │ │ │ queries    │ │
│ RWMutex     │ │ └─────┬──────┘ │
└─────────────┘ │       │        │
                │  ┌────▼─────┐  │
                │  │ pgxpool  │  │
                │  └────┬─────┘  │
                └───────┼────────┘
                        │
                  ┌─────▼──────┐
                  │ PostgreSQL │
                  │ + TimescaleDB
                  └────────────┘
```

### Schéma de base de données

```sql
devices
├─ id          UUID PRIMARY KEY
├─ name        VARCHAR(255) NOT NULL
├─ type        VARCHAR(100) NOT NULL
├─ status      device_status (ENUM)
├─ created_at  TIMESTAMPTZ
├─ last_seen   TIMESTAMPTZ
└─ metadata    JSONB

Indexes:
- idx_devices_type (type)
- idx_devices_status (status)
- idx_devices_metadata (GIN on metadata)
```

## 🚀 Démarrage rapide

### Prérequis

- Go 1.24+
- Docker & Docker Compose (pour PostgreSQL)
- Protocol Buffers compiler (`protoc`)
- sqlc CLI tool

### Installation des outils

```bash
# Protocol Buffers
brew install protobuf

# Go tools
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# sqlc
brew install sqlc
```

### Démarrage avec Memory Storage (développement)

```bash
# Depuis la racine du projet
cd services/device-manager

# Générer le code Protocol Buffers
cd ../../shared/proto && ./generate.sh && cd -

# Lancer le service
go run main.go
```

Le service démarre sur `localhost:8081` avec le backend **in-memory**.

### Démarrage avec PostgreSQL (production-like)

```bash
# 1. Démarrer l'infrastructure Docker
make up

# 2. Lancer les migrations
make db-migrate

# 3. Lancer le service avec PostgreSQL
cd services/device-manager
STORAGE_TYPE=postgres go run main.go
```

### Via Makefile (recommandé)

```bash
# Depuis la racine du projet

# Démarrer toute l'infrastructure + services
make dev

# Ou services individuels
make device-manager      # Memory storage
make db-migrate         # Migrations PostgreSQL
```

## ⚙️ Configuration

### Variables d'environnement

| Variable | Description | Défaut | Production |
|----------|-------------|--------|------------|
| `STORAGE_TYPE` | Backend de stockage (`memory` ou `postgres`) | `memory` | `postgres` |
| `DB_HOST` | Hôte PostgreSQL | `localhost` | Variable |
| `DB_PORT` | Port PostgreSQL | `5432` | `5432` |
| `DB_NAME` | Nom de la base de données | `iot_platform` | Variable |
| `DB_USER` | Utilisateur PostgreSQL | `iot_user` | Variable |
| `DB_PASSWORD` | Mot de passe PostgreSQL | `iot_password` | **SECRET** |
| `DB_SSLMODE` | Mode SSL PostgreSQL | `disable` | `require` |

### Exemple de configuration

**Développement (Memory):**
```bash
# Aucune configuration nécessaire
go run main.go
```

**Développement (PostgreSQL local):**
```bash
export STORAGE_TYPE=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=iot_platform
export DB_USER=iot_user
export DB_PASSWORD=iot_password
export DB_SSLMODE=disable
go run main.go
```

**Production:**
```bash
export STORAGE_TYPE=postgres
export DB_HOST=postgres.production.example.com
export DB_PORT=5432
export DB_NAME=iot_platform_prod
export DB_USER=iot_app
export DB_PASSWORD="${DB_PASSWORD_SECRET}"  # Depuis secret manager
export DB_SSLMODE=require
./device-manager
```

## 📡 API gRPC

### Service Definition (Protocol Buffers)

```protobuf
service DeviceService {
  rpc CreateDevice(CreateDeviceRequest) returns (CreateDeviceResponse);
  rpc GetDevice(GetDeviceRequest) returns (GetDeviceResponse);
  rpc ListDevices(ListDevicesRequest) returns (ListDevicesResponse);
  rpc UpdateDevice(UpdateDeviceRequest) returns (UpdateDeviceResponse);
  rpc DeleteDevice(DeleteDeviceRequest) returns (DeleteDeviceResponse);
}
```

### Exemples d'utilisation

#### Avec grpcurl

**Installer grpcurl:**
```bash
brew install grpcurl
```

**Lister les services:**
```bash
grpcurl -plaintext localhost:8081 list
```

**Créer un device:**
```bash
grpcurl -plaintext -d '{
  "name": "Temperature Sensor",
  "type": "sensor",
  "metadata": {
    "location": "room-101",
    "model": "DHT22"
  }
}' localhost:8081 device.DeviceService/CreateDevice
```

**Récupérer un device:**
```bash
grpcurl -plaintext -d '{
  "id": "550e8400-e29b-41d4-a716-446655440000"
}' localhost:8081 device.DeviceService/GetDevice
```

**Lister les devices (paginé):**
```bash
grpcurl -plaintext -d '{
  "page": 1,
  "page_size": 10
}' localhost:8081 device.DeviceService/ListDevices
```

**Mettre à jour un device:**
```bash
grpcurl -plaintext -d '{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Updated Sensor",
  "status": "OFFLINE",
  "metadata": {
    "location": "room-102",
    "version": "2.0"
  }
}' localhost:8081 device.DeviceService/UpdateDevice
```

**Supprimer un device:**
```bash
grpcurl -plaintext -d '{
  "id": "550e8400-e29b-41d4-a716-446655440000"
}' localhost:8081 device.DeviceService/DeleteDevice
```

#### Depuis Go (client)

```go
package main

import (
    "context"
    "log"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    pb "github.com/yourusername/iot-platform/shared/proto"
)

func main() {
    conn, err := grpc.Dial(
        "localhost:8081",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    client := pb.NewDeviceServiceClient(conn)

    // Créer un device
    resp, err := client.CreateDevice(context.Background(), &pb.CreateDeviceRequest{
        Name: "My Sensor",
        Type: "temperature",
        Metadata: map[string]string{
            "location": "office",
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Device created: %s", resp.Device.Id)
}
```

## 🗄️ Base de données

### Schema PostgreSQL

Le schéma complet se trouve dans [`db/migrations/001_init.sql`](db/migrations/001_init.sql).

**Table `devices`:**
- **UUID** pour les IDs (uuid-ossp extension)
- **ENUM** pour le statut (type-safe)
- **JSONB** pour les métadonnées (flexible, indexé avec GIN)
- **TIMESTAMPTZ** pour les timestamps (timezone-aware)
- **Indexes** optimisés pour les requêtes courantes

### Migrations

**Appliquer les migrations:**
```bash
make db-migrate
```

**Réinitialiser la base:**
```bash
make db-reset
```

**Vérifier le statut:**
```bash
make db-status
```

**Accès direct PostgreSQL:**
```bash
docker-compose exec postgres psql -U iot_user -d iot_platform
```

### sqlc - SQL Type-Safe

Ce projet utilise [sqlc](https://sqlc.dev/) pour générer du code Go type-safe à partir de SQL.

**Queries SQL:** [`db/queries/devices.sql`](db/queries/devices.sql)

**Régénérer le code:**
```bash
make sqlc-generate
# ou
cd services/device-manager && sqlc generate
```

**Avantages:**
- ✅ Type-safety au compile-time
- ✅ Pas de reflection
- ✅ Performances optimales
- ✅ Intellisense/autocomplétion
- ✅ Détection d'erreurs SQL au build

### pgx - Driver PostgreSQL

[pgx](https://github.com/jackc/pgx) est utilisé comme driver PostgreSQL:
- **3-5x plus rapide** que lib/pq
- Support natif des types PostgreSQL
- Connection pooling intégré
- Préparation automatique des statements

## 🧪 Tests

### Tests unitaires (Memory Storage)

Tests rapides sans dépendances externes:

```bash
# Depuis la racine
make test-device

# Depuis le service
cd services/device-manager
go test ./... -v
```

**Coverage:** 6 suites de tests
- ✅ CreateDevice (4 cas)
- ✅ GetDevice (2 cas)
- ✅ ListDevices (2 cas)
- ✅ UpdateDevice (5 cas)
- ✅ DeleteDevice (3 cas)
- ✅ ConcurrentOperations (2 cas)

### Tests d'intégration (PostgreSQL)

Tests avec base de données réelle:

```bash
# Démarrer PostgreSQL
make up
make db-migrate

# Lancer les tests d'intégration
make test-device-integration
```

**Tests inclus:**
- ✅ CRUD complet avec persistance
- ✅ Pagination et listing
- ✅ Cohérence transactionnelle
- ✅ Gestion des timestamps
- ✅ Validation des contraintes

### Linter

```bash
make lint
```

## 🔧 Développement

### Structure du projet

```
services/device-manager/
├── db/
│   ├── migrations/          # SQL migration files
│   │   └── 001_init.sql
│   ├── queries/             # sqlc queries
│   │   └── devices.sql
│   └── sqlc/                # Generated code (ne pas éditer)
│       ├── db.go
│       ├── devices.sql.go
│       ├── models.go
│       └── querier.go
├── storage/
│   ├── storage.go           # Interface Storage
│   ├── memory.go            # In-memory implementation
│   └── postgres.go          # PostgreSQL implementation
├── main.go                  # Entry point
├── main_test.go             # Unit tests
├── integration_test.go      # Integration tests
├── sqlc.yaml                # sqlc configuration
└── README.md
```

### Ajouter une nouvelle query SQL

1. **Éditer** `db/queries/devices.sql`:
```sql
-- name: GetDevicesByStatus :many
SELECT * FROM devices WHERE status = $1 ORDER BY created_at DESC;
```

2. **Régénérer** le code:
```bash
make sqlc-generate
```

3. **Utiliser** dans `storage/postgres.go`:
```go
func (s *PostgresStorage) GetDevicesByStatus(ctx context.Context, status string) ([]*pb.Device, error) {
    dbDevices, err := s.queries.GetDevicesByStatus(ctx, status)
    // ...
}
```

### Modifier le schéma PostgreSQL

1. **Créer** une nouvelle migration `db/migrations/002_add_field.sql`
2. **Mettre à jour** `db/queries/devices.sql` si nécessaire
3. **Régénérer** sqlc: `make sqlc-generate`
4. **Appliquer** la migration: `make db-migrate`

### Workflows de développement

**Mode rapide (Memory):**
```bash
go run main.go
# Tests rapides, pas de setup
```

**Mode réaliste (PostgreSQL):**
```bash
make up && make db-migrate
STORAGE_TYPE=postgres go run main.go
# Tests avec vraie DB
```

**Tests complets:**
```bash
make test-device                # Unit tests
make test-device-integration    # Integration tests
make lint                       # Code quality
```

## 📚 Ressources

### Documentation

- [Protocol Buffers](https://protobuf.dev/)
- [gRPC Go](https://grpc.io/docs/languages/go/)
- [sqlc](https://docs.sqlc.dev/)
- [pgx](https://github.com/jackc/pgx)
- [PostgreSQL](https://www.postgresql.org/docs/)

### Outils

- [grpcurl](https://github.com/fullstorydev/grpcurl) - CLI pour tester gRPC
- [Evans](https://github.com/ktr0731/evans) - gRPC client interactif
- [Postman](https://www.postman.com/) - Support gRPC depuis v8.0

## 🐛 Troubleshooting

### Le service ne démarre pas

**Erreur:** `Failed to connect to PostgreSQL`
```bash
# Vérifier que PostgreSQL est lancé
docker-compose ps

# Vérifier les logs
docker-compose logs postgres

# Relancer
make up
```

### Les migrations échouent

**Erreur:** `relation "devices" already exists`
```bash
# Réinitialiser la base
make db-reset
```

### Les tests d'intégration échouent

```bash
# S'assurer que PostgreSQL est up
make up

# Appliquer les migrations
make db-migrate

# Relancer les tests
make test-device-integration
```

### Erreurs de compilation sqlc

```bash
# Régénérer le code
make sqlc-generate

# Si ça persiste, vérifier sqlc.yaml et les queries SQL
```

## 📝 TODO

- [ ] Stream temps réel avec WatchDevices
- [ ] Authentification gRPC (mTLS)
- [ ] Métriques Prometheus
- [ ] Tracing distribué (OpenTelemetry)
- [ ] Health checks (liveness/readiness)
- [ ] Graceful shutdown
- [ ] Rate limiting
- [ ] Circuit breaker
- [ ] Bulk operations

## 📄 License

Propriétaire - Tous droits réservés
