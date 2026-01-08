# IoT Platform

Plateforme IoT complète pour la gestion et le monitoring d'appareils connectés.

> **🚀 Nouveau ?** Commence par le [Guide de démarrage](GETTING_STARTED.md) pour une introduction complète !

## 🎯 Architecture de communication

- **GraphQL** : API publique pour les clients Web/Mobile
- **gRPC** : Communication inter-services (haute performance)
- **MQTT** : Communication avec les devices IoT
- **Protocol Buffers** : Contrats d'API stricts et typés

## 🏗️ Architecture

### Services principaux

- **API Gateway** (Go) - Point d'entrée unique, authentification, rate limiting
- **Device Manager** (Go) - Gestion du cycle de vie des devices IoT
- **Data Collector** (Rust) - Collecte et traitement temps réel des données
- **Time Series DB** - Stockage des métriques (TimescaleDB)
- **Message Broker** - Communication MQTT pour les devices
- **Web Dashboard** (React) - Interface web de monitoring
- **Mobile App** (Flutter) - Application mobile

### Stack technique

#### Backend
- Go 1.21+ (API Gateway, Device Manager)
- Rust 1.75+ (Data Collector, Edge Processing)
- MQTT Broker (Mosquitto)
- Redis (Cache & Pub/Sub)
- PostgreSQL 16 + TimescaleDB

#### Frontend
- React 18 avec TypeScript
- Flutter 3.x

#### Infrastructure
- Docker & Docker Compose
- Kubernetes (EKS)
- Terraform (AWS)
- Prometheus & Grafana (Monitoring)
- GitHub Actions (CI/CD)

## 📁 Structure du projet

```
iot-platform/
├── services/
│   ├── api-gateway/          # Go - API Gateway
│   ├── device-manager/       # Go - Gestion des devices
│   ├── data-collector/       # Rust - Collecte temps réel
│   └── notification-service/ # Go - Alertes et notifications
├── frontends/
│   ├── web-dashboard/        # React - Dashboard web
│   └── mobile-app/           # Flutter - App mobile
├── infrastructure/
│   ├── terraform/            # IaC AWS
│   ├── kubernetes/           # Manifests K8s
│   └── docker/               # Dockerfiles & compose
├── shared/
│   ├── proto/                # Protocol Buffers
│   └── schemas/              # Schémas de données
└── docs/
    ├── architecture/         # Diagrammes d'architecture
    └── api/                  # Documentation API

```

## 🚀 Démarrage rapide

**Pour une explication détaillée, voir le [Guide de démarrage complet](GETTING_STARTED.md)**

### Prérequis minimaux

- Docker Desktop
- Go 1.21+
- Protocol Buffers Compiler : `brew install protobuf`

### Installation rapide

```bash
# 1. Installer les outils et dépendances
make setup

# 2. Générer le code (Protocol Buffers + GraphQL)
make generate

# 3. Démarrer l'infrastructure (PostgreSQL, Redis, MQTT, etc.)
make start

# 4. Lancer les services (dans des terminaux séparés)
make device-manager    # Terminal 1 - gRPC sur port 8081
make api-gateway       # Terminal 2 - GraphQL sur port 8080
```

### Tester l'API

Ouvre http://localhost:8080 dans ton navigateur pour accéder au **GraphQL Playground**.

Exemple de requête :
```graphql
mutation {
  createDevice(input: {
    name: "Capteur Température"
    type: "temperature_sensor"
  }) {
    id
    name
    status
  }
}
```

### Déploiement

```bash
# Infrastructure AWS
cd infrastructure/terraform
terraform init
terraform plan
terraform apply

# Déploiement Kubernetes
kubectl apply -f infrastructure/kubernetes/
```

## 📊 Monitoring

- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000
- API Gateway: http://localhost:8080
- Web Dashboard: http://localhost:3001

## 🔒 Sécurité

- Authentification JWT
- TLS/SSL pour toutes les communications
- Secrets gérés via AWS Secrets Manager
- Rate limiting sur l'API Gateway
- RBAC sur Kubernetes

## 📝 License

MIT