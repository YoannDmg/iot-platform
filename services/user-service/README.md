# User Service

> Microservice gRPC d'authentification et gestion des utilisateurs

[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go&logoColor=white)](https://golang.org)
[![gRPC](https://img.shields.io/badge/gRPC-HTTP%2F2-4285F4)](https://grpc.io)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-4169E1?logo=postgresql&logoColor=white)](https://postgresql.org)

## Table des matières

- [Vue d'ensemble](#vue-densemble)
- [Architecture](#architecture)
- [Démarrage rapide](#démarrage-rapide)
- [Configuration](#configuration)
- [API gRPC](#api-grpc)
- [Base de données](#base-de-données)
- [Tests](#tests)

## Vue d'ensemble

Le User Service gère l'authentification et le cycle de vie des utilisateurs de la plateforme IoT. Il fournit une API gRPC pour l'inscription, la connexion et la gestion des comptes avec support de rôles.

### Fonctionnalités

- **Authentification** — Inscription, connexion avec validation bcrypt
- **Gestion des rôles** — admin, user, device
- **CRUD utilisateurs** — Création, lecture, mise à jour, suppression
- **Dual storage** — PostgreSQL (production) et In-Memory (dev/tests)
- **Sécurité** — Hachage bcrypt, validation email, comptes désactivables

### Technologies

| Composant | Technologie |
|-----------|-------------|
| Langage | Go 1.24 |
| Protocol | gRPC (HTTP/2) |
| Database | PostgreSQL 15+ |
| ORM | sqlc (SQL-first, type-safe) |
| Driver | pgx/v5 |
| Hachage | bcrypt |

## Architecture

```
┌─────────────────┐
│   API Gateway   │
│    (GraphQL)    │
└────────┬────────┘
         │ gRPC
         ▼
┌─────────────────────────────┐
│       User Service          │
│         Port 8082           │
├─────────────────────────────┤
│     Storage Interface       │
│  - CreateUser               │
│  - Authenticate             │
│  - GetUser / GetByEmail     │
│  - ListUsers                │
│  - UpdateUser / Delete      │
└──────────┬──────────────────┘
           │
    ┌──────┴──────┐
    │             │
┌───▼─────┐  ┌───▼──────────┐
│ Memory  │  │  PostgreSQL  │
│ Storage │  │   Storage    │
└─────────┘  └──────────────┘
```

### Structure du projet

```
user-service/
├── main.go              # Point d'entrée, serveur gRPC
├── main_test.go         # Tests unitaires
├── storage/
│   ├── storage.go       # Interface Storage
│   ├── memory.go        # Implémentation in-memory
│   └── postgres.go      # Implémentation PostgreSQL
├── db/
│   ├── queries/         # Requêtes SQL (sqlc)
│   └── sqlc/            # Code généré
├── Dockerfile
└── sqlc.yaml
```

## Démarrage rapide

### Prérequis

- Go 1.24+
- Docker (pour PostgreSQL)
- Protocol Buffers compiler

### Lancement

```bash
# Depuis la racine du projet
make dev-users

# Ou directement
cd services/user-service
go run main.go
```

Le service démarre sur `localhost:8082`.

## Configuration

### Variables d'environnement

| Variable | Description | Défaut |
|----------|-------------|--------|
| `USER_SERVICE_PORT` | Port gRPC | `8082` |
| `STORAGE_TYPE` | Backend (`memory`, `postgres`) | `memory` |
| `DB_HOST` | Hôte PostgreSQL | `localhost` |
| `DB_PORT` | Port PostgreSQL | `5432` |
| `DB_NAME` | Nom de la base | `iot_platform` |
| `DB_USER` | Utilisateur | `iot_user` |
| `DB_PASSWORD` | Mot de passe | `iot_password` |
| `DB_SSLMODE` | Mode SSL | `disable` |

## API gRPC

### Service Definition

```protobuf
service UserService {
  rpc Register(RegisterRequest) returns (RegisterResponse);
  rpc Authenticate(AuthenticateRequest) returns (AuthenticateResponse);
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc GetUserByEmail(GetUserByEmailRequest) returns (GetUserByEmailResponse);
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
  rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse);
  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse);
}
```

### Exemples avec grpcurl

> Les commandes suivantes doivent être exécutées depuis la **racine du projet**.

**Créer un utilisateur :**
```bash
grpcurl -plaintext \
  -import-path shared/proto \
  -proto user/user.proto \
  -d '{
    "email": "admin@example.com",
    "password": "secret123",
    "name": "Admin User",
    "role": "admin"
  }' localhost:8082 user.UserService/Register
```

**Authentifier :**
```bash
grpcurl -plaintext \
  -import-path shared/proto \
  -proto user/user.proto \
  -d '{
    "email": "admin@example.com",
    "password": "secret123"
  }' localhost:8082 user.UserService/Authenticate
```

**Lister les utilisateurs :**
```bash
grpcurl -plaintext \
  -import-path shared/proto \
  -proto user/user.proto \
  -d '{
    "page": 1,
    "page_size": 10,
    "role": "admin"
  }' localhost:8082 user.UserService/ListUsers
```

**Récupérer un utilisateur :**
```bash
grpcurl -plaintext \
  -import-path shared/proto \
  -proto user/user.proto \
  -d '{
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }' localhost:8082 user.UserService/GetUser
```

### Modèle User

| Champ | Type | Description |
|-------|------|-------------|
| `id` | string | UUID |
| `email` | string | Email unique |
| `name` | string | Nom complet |
| `role` | string | admin, user, device |
| `created_at` | int64 | Timestamp création |
| `last_login` | int64 | Dernier login |
| `is_active` | bool | Compte actif |

## Base de données

### Schéma

```sql
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name          VARCHAR(255) NOT NULL,
    role          VARCHAR(50) NOT NULL CHECK (role IN ('admin', 'user', 'device')),
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    last_login    TIMESTAMPTZ,
    is_active     BOOLEAN DEFAULT true
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_active ON users(is_active);
```

### Requêtes sqlc

Les requêtes SQL sont définies dans `db/queries/users.sql` et le code Go est généré avec :

```bash
cd services/user-service && sqlc generate
```

## Tests

```bash
# Tests unitaires
cd services/user-service
go test -v ./...

# Tests d'intégration (nécessite PostgreSQL)
go test -tags=integration -v ./storage/...
```

## Sécurité

- Mots de passe hachés avec bcrypt (cost par défaut)
- Validation email au niveau base de données
- Comptes désactivables (`is_active`)
- Codes d'erreur gRPC appropriés (NotFound, AlreadyExists, InvalidArgument)

## License

MIT