# Notes d'apprentissage - Plateforme IoT

> Document évolutif - Notes de cours sur la construction d'une plateforme IoT microservices

**Dernière mise à jour:** 2026-01-09
**Commit:** `cae20ba`

---

## Table des matières

1. [Architecture générale](#1-architecture-générale)
2. [Protocol Buffers et gRPC](#2-protocol-buffers-et-grpc)
3. [GraphQL](#3-graphql)
4. [Génération de code](#4-génération-de-code)
5. [Synchronisation et thread-safety](#5-synchronisation-et-thread-safety)
6. [Docker et orchestration](#6-docker-et-orchestration)
7. [Patterns et bonnes pratiques](#7-patterns-et-bonnes-pratiques)
8. [Commandes utiles](#8-commandes-utiles)

---

## 1. Architecture générale

### 1.1 Vue d'ensemble

L'architecture microservices sépare les responsabilités en services indépendants qui communiquent entre eux.

```
┌────────────────────┐
│  Web/Mobile Client │  ← Utilisateurs finaux
└─────────┬──────────┘
          │ GraphQL (HTTP/JSON)
    ┌─────▼──────────┐
    │  API Gateway   │  ← Point d'entrée public
    │     (Go)       │     Port 8080
    └─────────┬──────┘
              │ gRPC (HTTP/2/Protobuf)
    ┌─────────▼──────────┐
    │  Device Manager    │  ← Logique métier
    │      (Go)          │     Port 8081
    └─────────┬──────────┘
              │
      ┌───────┴────────┐
      │   PostgreSQL   │  ← Persistance
      │   Redis        │  ← Cache
      │   MQTT         │  ← IoT devices
      └────────────────┘
```

### 1.2 Pourquoi cette architecture ?

**GraphQL pour l'externe:**
- Les clients demandent exactement ce dont ils ont besoin
- Un seul endpoint au lieu de dizaines de routes REST
- Documentation auto-générée
- Typage fort des requêtes

**gRPC pour l'interne:**
- Performance: protocole binaire (Protobuf) vs JSON
- HTTP/2: multiplexing, streaming
- Contrat strict: le fichier `.proto` définit l'API
- Génération automatique du code client/serveur

### 1.3 Flux de données complet

```
1. Client envoie une mutation GraphQL
   POST /query
   { "query": "mutation { createDevice(...) { id } }" }

2. API Gateway reçoit et parse la requête
   - Valide la syntaxe GraphQL
   - Appelle le resolver correspondant

3. Resolver appelle Device Manager via gRPC
   req := &pb.CreateDeviceRequest{Name: "sensor", Type: "temp"}
   resp, err := grpcClient.CreateDevice(ctx, req)

4. Device Manager traite la logique métier
   - Valide les données
   - Génère un UUID
   - Stocke en mémoire (ou DB en production)
   - Retourne le Device en Protobuf

5. API Gateway convertit Protobuf → JSON
   Le resolver transforme *pb.Device en type GraphQL

6. Client reçoit la réponse JSON
   { "data": { "createDevice": { "id": "uuid..." } } }
```

---

## 2. Protocol Buffers et gRPC

### 2.1 C'est quoi Protocol Buffers ?

Protocol Buffers (protobuf) est un **langage de définition d'interface** (IDL) développé par Google.

**Analogie:** C'est comme un contrat entre services. Si tu changes le contrat, tout le monde le sait immédiatement.

### 2.2 Anatomie d'un fichier .proto

```protobuf
// shared/proto/device.proto

syntax = "proto3";  // Version du protocole

package device;     // Namespace

// Import d'autres protos si besoin
import "google/protobuf/timestamp.proto";

// Définition d'un message (= struct en Go)
message Device {
  string id = 1;        // Le numéro = tag unique (ne JAMAIS changer)
  string name = 2;      // Types: string, int32, int64, bool, bytes...
  string type = 3;
  DeviceStatus status = 4;
  int64 created_at = 5;
  int64 last_seen = 6;
  repeated KeyValue metadata = 7;  // repeated = array
}

// Enum pour les statuts
enum DeviceStatus {
  UNKNOWN = 0;      // 0 est obligatoire comme valeur par défaut
  ONLINE = 1;
  OFFLINE = 2;
  ERROR = 3;
  MAINTENANCE = 4;
}

// Message pour key-value
message KeyValue {
  string key = 1;
  string value = 2;
}

// Définition du service gRPC
service DeviceService {
  // RPC = Remote Procedure Call
  rpc CreateDevice(CreateDeviceRequest) returns (CreateDeviceResponse);
  rpc GetDevice(GetDeviceRequest) returns (GetDeviceResponse);
  rpc ListDevices(ListDevicesRequest) returns (ListDevicesResponse);

  // Streaming: le serveur envoie plusieurs réponses
  rpc WatchDevices(Empty) returns (stream Device);
}

// Messages pour les requêtes/réponses
message CreateDeviceRequest {
  string name = 1;
  string type = 2;
  repeated KeyValue metadata = 3;
}

message CreateDeviceResponse {
  Device device = 1;
}
```

### 2.3 Génération du code Go

Quand tu exécutes `make generate-proto`, voici ce qui se passe:

```bash
protoc \
  --go_out=. \                    # Génère les structs Go
  --go_opt=paths=source_relative \
  --go-grpc_out=. \               # Génère le code gRPC
  --go-grpc_opt=paths=source_relative \
  device.proto
```

**Fichiers générés:**

1. `device.pb.go` - Les structs
```go
type Device struct {
    Id        string        `protobuf:"bytes,1,opt,name=id,proto3"`
    Name      string        `protobuf:"bytes,2,opt,name=name,proto3"`
    Type      string        `protobuf:"bytes,3,opt,name=type,proto3"`
    Status    DeviceStatus  `protobuf:"varint,4,opt,name=status,proto3,enum=device.DeviceStatus"`
    // ...
}
```

2. `device_grpc.pb.go` - Les interfaces et clients
```go
// Interface que ton serveur doit implémenter
type DeviceServiceServer interface {
    CreateDevice(context.Context, *CreateDeviceRequest) (*CreateDeviceResponse, error)
    GetDevice(context.Context, *GetDeviceRequest) (*GetDeviceResponse, error)
    // ...
}

// Client généré pour appeler le service
type DeviceServiceClient interface {
    CreateDevice(ctx context.Context, in *CreateDeviceRequest, opts ...grpc.CallOption) (*CreateDeviceResponse, error)
    // ...
}
```

### 2.4 Implémentation du serveur gRPC

```go
// services/device-manager/main.go

// Structure qui implémente l'interface DeviceServiceServer
type DeviceServer struct {
    pb.UnimplementedDeviceServiceServer  // Embed pour forward compatibility
    mu      sync.RWMutex                 // Pour thread-safety
    devices map[string]*pb.Device        // Stockage en mémoire
}

// Implémentation d'une méthode RPC
func (s *DeviceServer) CreateDevice(
    ctx context.Context,
    req *pb.CreateDeviceRequest,
) (*pb.CreateDeviceResponse, error) {
    // 1. Validation
    if req.Name == "" {
        return nil, status.Error(codes.InvalidArgument, "name required")
    }

    // 2. Création
    device := &pb.Device{
        Id:        uuid.New().String(),
        Name:      req.Name,
        Type:      req.Type,
        Status:    pb.DeviceStatus_ONLINE,
        CreatedAt: time.Now().Unix(),
    }

    // 3. Stockage (thread-safe)
    s.mu.Lock()
    s.devices[device.Id] = device
    s.mu.Unlock()

    // 4. Retour
    return &pb.CreateDeviceResponse{Device: device}, nil
}

// Démarrage du serveur
func main() {
    listener, _ := net.Listen("tcp", ":8081")
    grpcServer := grpc.NewServer()

    // Enregistrement du service
    pb.RegisterDeviceServiceServer(grpcServer, NewDeviceServer())

    grpcServer.Serve(listener)  // Bloquant
}
```

### 2.5 Appel du service depuis un client

```go
// Dans l'API Gateway (futur)
conn, err := grpc.Dial("localhost:8081", grpc.WithInsecure())
client := pb.NewDeviceServiceClient(conn)

resp, err := client.CreateDevice(context.Background(), &pb.CreateDeviceRequest{
    Name: "Sensor 1",
    Type: "temperature",
})

fmt.Println(resp.Device.Id)  // UUID généré
```

### 2.6 Avantages de gRPC

| Aspect | REST/JSON | gRPC/Protobuf |
|--------|-----------|---------------|
| **Taille** | ~1KB | ~300 bytes |
| **Parse** | Lent (JSON parsing) | Rapide (binaire) |
| **Typage** | Faible (JSON Schema optionnel) | Fort (`.proto` obligatoire) |
| **HTTP** | HTTP/1.1 | HTTP/2 (multiplexing) |
| **Streaming** | Compliqué (SSE, WebSocket) | Natif |
| **Génération** | Manuelle ou Swagger | Automatique (protoc) |

---

## 3. GraphQL

### 3.1 Pourquoi GraphQL ?

**Problème avec REST:**
```
GET /devices              → Liste tous les devices (trop de data)
GET /devices/123          → Device complet (trop de data)
GET /devices/123/name     → Faut créer une route custom
GET /devices?fields=id,name  → Non standard
```

**Solution GraphQL:**
```graphql
query {
  device(id: "123") {
    id
    name
    # Le client demande exactement ce qu'il veut
  }
}
```

### 3.2 Schéma GraphQL

```graphql
# services/api-gateway/schema.graphql

# Types de base
type Device {
  id: ID!           # ! = obligatoire
  name: String!
  type: String!
  status: DeviceStatus!
  createdAt: Int!
  lastSeen: Int!
  metadata: [KeyValue!]!
}

enum DeviceStatus {
  UNKNOWN
  ONLINE
  OFFLINE
  ERROR
  MAINTENANCE
}

type KeyValue {
  key: String!
  value: String!
}

# Input types (pour les mutations)
input CreateDeviceInput {
  name: String!
  type: String!
  metadata: [KeyValueInput!]
}

input KeyValueInput {
  key: String!
  value: String!
}

# Queries (lecture)
type Query {
  device(id: ID!): Device
  devices(page: Int!, pageSize: Int!): DeviceConnection!
  stats: DeviceStats!
}

# Mutations (écriture)
type Mutation {
  createDevice(input: CreateDeviceInput!): Device!
  updateDevice(id: ID!, input: UpdateDeviceInput!): Device!
  deleteDevice(id: ID!): Boolean!
}

# Subscriptions (temps réel)
type Subscription {
  deviceUpdated: Device!
}

# Types pour pagination
type DeviceConnection {
  devices: [Device!]!
  total: Int!
  page: Int!
  pageSize: Int!
}
```

### 3.3 Génération du code GraphQL

```bash
# Dans services/api-gateway/
gqlgen generate
```

**Fichiers générés:**

1. `graph/generated/generated.go` - Le serveur GraphQL
2. `graph/model/models_gen.go` - Les types Go
3. `graph/schema.resolvers.go` - Les fonctions à implémenter

### 3.4 Resolvers (à implémenter)

```go
// graph/schema.resolvers.go

func (r *mutationResolver) CreateDevice(
    ctx context.Context,
    input model.CreateDeviceInput,
) (*model.Device, error) {
    // TODO: Appeler le Device Manager via gRPC

    // 1. Connexion gRPC
    conn, err := grpc.Dial("localhost:8081", grpc.WithInsecure())
    defer conn.Close()

    client := pb.NewDeviceServiceClient(conn)

    // 2. Conversion GraphQL input → Protobuf request
    req := &pb.CreateDeviceRequest{
        Name: input.Name,
        Type: input.Type,
        Metadata: convertMetadata(input.Metadata),
    }

    // 3. Appel gRPC
    resp, err := client.CreateDevice(ctx, req)
    if err != nil {
        return nil, err
    }

    // 4. Conversion Protobuf response → GraphQL type
    return &model.Device{
        ID:        resp.Device.Id,
        Name:      resp.Device.Name,
        Type:      resp.Device.Type,
        Status:    convertStatus(resp.Device.Status),
        CreatedAt: int(resp.Device.CreatedAt),
    }, nil
}
```

### 3.5 Utilisation du GraphQL Playground

```
http://localhost:8080
```

**Exemple de mutation:**
```graphql
mutation CreateSensor {
  createDevice(input: {
    name: "Temperature Sensor"
    type: "temperature_sensor"
    metadata: [
      { key: "location", value: "living_room" }
      { key: "floor", value: "1" }
    ]
  }) {
    id
    name
    status
    createdAt
  }
}
```

**Exemple de query:**
```graphql
query GetDevices {
  devices(page: 1, pageSize: 10) {
    devices {
      id
      name
      status
    }
    total
  }
}
```

---

## 4. Génération de code

### 4.1 Pourquoi générer du code ?

**Sans génération:**
- Écrire manuellement les structs
- Écrire le code de serialization
- Écrire les clients/serveurs
- Risque d'erreurs
- Pas de contrat strict

**Avec génération:**
- Le contrat (`.proto`, `.graphql`) est la source de vérité
- Code généré automatiquement
- Typage fort garanti
- Si le contrat change, la compilation échoue

### 4.2 Pipeline de génération

```
┌─────────────────┐
│ device.proto    │  Définition du service
└────────┬────────┘
         │
         │ make generate-proto
         ▼
  ┌──────────────┐
  │    protoc    │  Compilateur Protocol Buffers
  └──────┬───────┘
         │
         ├──→ device.pb.go        (structs)
         └──→ device_grpc.pb.go   (client/serveur)
```

```
┌─────────────────┐
│ schema.graphql  │  Schéma GraphQL
└────────┬────────┘
         │
         │ make generate-graphql
         ▼
  ┌──────────────┐
  │   gqlgen     │  Générateur GraphQL
  └──────┬───────┘
         │
         ├──→ generated.go        (serveur)
         ├──→ models_gen.go       (types)
         └──→ schema.resolvers.go (à implémenter)
```

### 4.3 Commandes Make

```makefile
# Makefile

generate-proto:
	cd shared/proto && ./generate.sh

generate-graphql:
	cd services/api-gateway && gqlgen generate

generate: generate-proto generate-graphql
```

**Workflow:**
1. Modifier `device.proto` ou `schema.graphql`
2. `make generate`
3. Implémenter les nouvelles fonctions
4. Compiler

---

## 5. Synchronisation et thread-safety

### 5.1 Le problème

En Go, les goroutines s'exécutent en parallèle. Si plusieurs requêtes modifient la même map simultanément, c'est le **race condition** → crash.

```go
// ❌ DANGEREUX
type Server struct {
    devices map[string]*Device
}

func (s *Server) CreateDevice(...) {
    s.devices[id] = device  // Plusieurs goroutines peuvent écrire ici
}
```

### 5.2 Solution: sync.RWMutex

```go
// ✅ SAFE
type DeviceServer struct {
    mu      sync.RWMutex              // Read-Write Mutex
    devices map[string]*pb.Device
}

// Écriture (exclusif)
func (s *DeviceServer) CreateDevice(...) {
    s.mu.Lock()              // Bloque tout (lecture ET écriture)
    s.devices[id] = device
    s.mu.Unlock()
}

// Lecture (partagé)
func (s *DeviceServer) GetDevice(...) {
    s.mu.RLock()             // Permet plusieurs lecteurs simultanés
    device := s.devices[id]
    s.mu.RUnlock()
}
```

### 5.3 RWMutex vs Mutex

| Type | Lock | RLock | Utilisation |
|------|------|-------|-------------|
| `sync.Mutex` | Exclusif | N/A | Toujours exclusif |
| `sync.RWMutex` | Exclusif | Partagé | Optimisé pour lectures fréquentes |

**Quand utiliser RWMutex ?**
- Beaucoup de lectures (`GetDevice`, `ListDevices`)
- Peu d'écritures (`CreateDevice`, `UpdateDevice`)

### 5.4 Deadlock et bonnes pratiques

```go
// ❌ Deadlock
func (s *DeviceServer) BadMethod() {
    s.mu.Lock()
    s.AnotherMethod()  // Essaie de Lock() à nouveau → deadlock
    s.mu.Unlock()
}

// ✅ Bon
func (s *DeviceServer) GoodMethod() {
    s.mu.Lock()
    defer s.mu.Unlock()  // Garantit Unlock même si panic

    // Code ici
}
```

**Règles:**
1. Lock/Unlock toujours dans la même fonction
2. Utiliser `defer` pour garantir Unlock
3. Ne jamais appeler une fonction qui Lock quand on a déjà Lock
4. Scope minimal: Lock le plus tard possible, Unlock le plus tôt possible

---

## 6. Docker et orchestration

### 6.1 Docker Compose

```yaml
# docker-compose.yml

version: '3.9'

services:
  # Base de données
  postgres:
    image: timescale/timescaledb:latest-pg16
    container_name: iot-postgres
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: iot_user
      POSTGRES_PASSWORD: iot_password
      POSTGRES_DB: iot_platform
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U iot_user"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Cache
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  # MQTT Broker
  mosquitto:
    image: eclipse-mosquitto:2
    ports:
      - "1883:1883"   # MQTT
      - "9001:9001"   # WebSocket
    volumes:
      - ./infrastructure/docker/mosquitto/mosquitto.conf:/mosquitto/config/mosquitto.conf

volumes:
  postgres_data:
  redis_data:

networks:
  default:
    name: iot-platform-network
```

### 6.2 Commandes Docker

```bash
# Démarrer tout
docker-compose up -d

# Voir les logs
docker-compose logs -f postgres

# Arrêter tout
docker-compose down

# Arrêter et supprimer les volumes
docker-compose down -v

# Status
docker-compose ps
```

### 6.3 Volumes et persistance

```
postgres_data:/var/lib/postgresql/data
     ↑                    ↑
  volume             path dans container
```

**Sans volume:** Les données disparaissent quand le container s'arrête
**Avec volume:** Les données persistent sur le disque

---

## 7. Patterns et bonnes pratiques

### 7.1 Separation of concerns

```
services/
├── api-gateway/      → Exposition publique (GraphQL)
├── device-manager/   → Logique métier (gRPC)
└── data-collector/   → Traitement temps réel (Rust)

shared/
└── proto/           → Contrats partagés
```

**Principe:** Chaque service a UNE responsabilité.

### 7.2 Error handling en Go

```go
// ❌ Mauvais
func CreateDevice(...) (*Device, error) {
    device := &Device{}
    return device, nil  // Ignore les erreurs
}

// ✅ Bon
func CreateDevice(...) (*Device, error) {
    if req.Name == "" {
        return nil, status.Error(codes.InvalidArgument, "name required")
    }

    device, err := s.repository.Save(device)
    if err != nil {
        log.Printf("Failed to save device: %v", err)
        return nil, status.Error(codes.Internal, "database error")
    }

    return device, nil
}
```

### 7.3 Codes d'erreur gRPC

```go
codes.InvalidArgument  → Client error (400)
codes.NotFound         → Resource not found (404)
codes.Internal         → Server error (500)
codes.Unauthenticated  → Auth required (401)
codes.PermissionDenied → Forbidden (403)
```

### 7.4 Context en Go

```go
func (s *DeviceServer) CreateDevice(
    ctx context.Context,  // Toujours premier argument
    req *pb.CreateDeviceRequest,
) (*pb.CreateDeviceResponse, error) {
    // Vérifier si le client a annulé
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // Propager le context aux appels downstream
    result, err := s.database.Save(ctx, device)

    return result, err
}
```

**Le context sert à:**
- Timeout: `ctx, cancel := context.WithTimeout(parent, 5*time.Second)`
- Annulation: `cancel()` annule toutes les opérations
- Metadata: headers, trace ID, etc.

---

## 8. Commandes utiles

### 8.1 Make commands

```bash
make setup              # Install tools + dependencies
make generate           # Generate proto + GraphQL code
make start              # Start Docker infrastructure
make stop               # Stop infrastructure
make device-manager     # Run Device Manager
make api-gateway        # Run API Gateway
make status             # View Docker status
make logs               # View logs
```

### 8.2 Git

```bash
git status              # Voir les modifications
git add -A              # Ajouter tous les fichiers
git commit -m "message" # Créer un commit
git log --oneline       # Historique
git diff                # Voir les changements
```

### 8.3 Go

```bash
go mod init module-name     # Créer go.mod
go mod tidy                 # Nettoyer les dépendances
go mod download             # Télécharger les dépendances
go build                    # Compiler
go run main.go              # Compiler + exécuter
go test ./...               # Lancer les tests
```

### 8.4 Protocol Buffers

```bash
protoc --version            # Vérifier installation
protoc --go_out=. file.proto  # Générer code Go
```

### 8.5 Docker

```bash
docker ps                   # Containers running
docker logs container-name  # Voir les logs
docker exec -it container bash  # Se connecter au container
docker-compose up -d        # Démarrer en background
docker-compose down -v      # Tout supprimer (volumes inclus)
```

---

## 9. Prochaines étapes

### 9.1 TODO immédiat

- [ ] Implémenter les resolvers GraphQL dans l'API Gateway
- [ ] Connecter l'API Gateway au Device Manager via gRPC
- [ ] Remplacer le stockage en mémoire par PostgreSQL
- [ ] Ajouter des tests unitaires

### 9.2 TODO moyen terme

- [ ] Authentification JWT
- [ ] Rate limiting
- [ ] Health checks
- [ ] Graceful shutdown
- [ ] Logging structuré
- [ ] Metrics (Prometheus)

### 9.3 TODO long terme

- [ ] Service de collecte de données (Rust)
- [ ] Dashboard web (React)
- [ ] Application mobile (Flutter)
- [ ] Déploiement Kubernetes
- [ ] CI/CD (GitHub Actions)

---

## 10. Ressources

### Documentation officielle

- **gRPC:** https://grpc.io/docs/languages/go/
- **Protocol Buffers:** https://protobuf.dev/
- **GraphQL:** https://graphql.org/learn/
- **gqlgen:** https://gqlgen.com/
- **Docker Compose:** https://docs.docker.com/compose/

### Concepts avancés à étudier

1. **gRPC streaming:** Server-side, client-side, bidirectional
2. **GraphQL subscriptions:** WebSocket pour temps réel
3. **Distributed tracing:** OpenTelemetry
4. **Service mesh:** Istio, Linkerd
5. **Event sourcing:** Event-driven architecture

---

## Notes personnelles

_Ajouter tes notes ici au fur et à mesure de l'apprentissage..._

