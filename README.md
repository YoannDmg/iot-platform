# IoT Platform

> Plateforme de gestion et monitoring d'appareils IoT — Architecture microservices

[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go&logoColor=white)](https://golang.org)
[![React](https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=black)](https://react.dev)
[![GraphQL](https://img.shields.io/badge/GraphQL-E10098?logo=graphql&logoColor=white)](https://graphql.org)
[![gRPC](https://img.shields.io/badge/gRPC-HTTP%2F2-4285F4)](https://grpc.io)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?logo=postgresql&logoColor=white)](https://postgresql.org)
[![MQTT](https://img.shields.io/badge/MQTT-3.1.1-660066?logo=mqtt&logoColor=white)](https://mqtt.org)

## Table des matières

- [Vue d'ensemble](#vue-densemble)
- [Architecture](#architecture)
- [Démarrage rapide](#démarrage-rapide)
- [Services](#services)
- [Configuration](#configuration)
- [Développement](#développement)
- [Tests](#tests)
- [Monitoring](#monitoring)

## Vue d'ensemble

Plateforme complète pour la gestion d'appareils IoT, conçue autour d'une architecture microservices. Elle permet l'enregistrement de devices, la collecte de télémétrie en temps réel via MQTT, et le monitoring via une interface React.

### Fonctionnalités

- **Gestion des devices** — CRUD complet, statuts, métadonnées flexibles (JSONB)
- **Collecte télémétrie** — Ingestion MQTT temps réel, stockage TimescaleDB
- **Authentification** — JWT avec gestion des rôles (admin/user)
- **API GraphQL** — Point d'entrée unique, typage strict, playground intégré
- **Dashboard** — Interface React pour le monitoring et la configuration
- **Observabilité** — Métriques Prometheus, dashboards Grafana

### Stack technique

| Couche | Technologies |
|--------|--------------|
| **Backend** | Go 1.24, gRPC, GraphQL (gqlgen), Protocol Buffers |
| **Frontend** | React 19, TypeScript, Vite, Apollo Client, TailwindCSS |
| **Base de données** | PostgreSQL 16, TimescaleDB |
| **Messaging** | MQTT (Mosquitto), Redis |
| **Monitoring** | Prometheus, Grafana |
| **Infrastructure** | Docker Compose |

## Architecture

```
┌─────────────┐             ┌──────────────┐
│ IoT Devices │────MQTT────►│ MQTT Broker  │─────────────────────┐
└─────────────┘             │ (Mosquitto)  │                     │
                            └──────────────┘                     │
                                                                 │
┌─────────────┐             ┌──────────────────┐                 │
│  Dashboard  │◄──GraphQL──►│   API Gateway    │                 │
│ React+Vite  │             │    Port 8080     │                 │
└─────────────┘             └────────┬─────────┘                 │
                                     │ gRPC                      │ MQTT
                   ┌─────────────────┼─────────────────┐         │
                   │                 │                 │         │
                   ▼                 ▼                 ▼         │
         ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
         │Device Manager│  │ User Service │  │Data Collector│◄───┘
         │  Port 8081   │  │  Port 8082   │  │  Port 8083   │
         └───────┬──────┘  └──────┬───────┘  └──────┬───────┘
                 │                │                 │
                 └────────────────┼─────────────────┘
                                  ▼
                       ┌────────────────────┐
                       │    PostgreSQL      │
                       │   + TimescaleDB    │
                       └────────────────────┘
```

### Communication

| Protocole | Usage |
|-----------|-------|
| **GraphQL** | API publique (clients web/mobile) |
| **gRPC** | Communication inter-services |
| **MQTT** | Communication devices IoT |
| **Protocol Buffers** | Contrats d'API typés |

## Démarrage rapide

### Prérequis

- Docker & Docker Compose
- Go 1.24+
- Node.js 20+
- Protocol Buffers : `brew install protobuf`

### Installation

```bash
# Cloner et configurer
git clone <repository-url>
cd iot-platform
cp .env.example .env

# Installer les outils Go et dépendances Node
make setup

# Générer le code (Protocol Buffers + GraphQL)
make generate
```

### Lancement

```bash
make start          # Tout en Docker
make dev            # Développement (services locaux + infra Docker)
make dev-dashboard  # Développement avec Dashboard React
```

Toutes les commandes disponibles :

```bash
make help
```

## Services

| Service | Port | Protocole | Description |
|---------|------|-----------|-------------|
| [API Gateway](services/api-gateway/) | 8080 | HTTP | Point d'entrée GraphQL, authentification JWT |
| [Device Manager](services/device-manager/) | 8081 | gRPC | Gestion du cycle de vie des devices IoT |
| [User Service](services/user-service/) | 8082 | gRPC | Authentification et gestion des utilisateurs |
| [Data Collector](services/data-collector/) | 8083 | gRPC + MQTT | Collecte des données IoT via MQTT |

## Configuration

```bash
cp .env.example .env
```

Le fichier `.env.example` contient toutes les variables documentées avec leurs valeurs par défaut. Voir les README des services pour les configurations spécifiques.

## Développement

### Structure du projet

```
iot-platform/
├── services/
│   ├── api-gateway/           # GraphQL + Auth
│   ├── device-manager/        # Gestion devices
│   ├── user-service/          # Authentification
│   └── data-collector/        # Ingestion MQTT
├── frontends/
│   └── dashboard/             # React + TypeScript
├── shared/
│   └── proto/                 # Protocol Buffers
├── infrastructure/
│   ├── database/              # Migrations SQL
│   └── docker/                # Config Prometheus, Grafana, Mosquitto
├── tests/
│   └── e2e/                   # Tests end-to-end
├── scripts/                   # Outils (simulateur)
└── docs/                      # Documentation Docusaurus
```

## Tests

```bash
make test               # Tests unitaires
make test-integration   # Tests d'intégration (nécessite DB)
make test-e2e           # Tests end-to-end (nécessite plateforme)
make simulate           # Simulateur de devices
```

## Documentation

```bash
make docs               # Serveur local (port 3001)
```

## License

MIT
