# API Gateway

Point d'entrée unique de la plateforme IoT. Expose une API GraphQL pour les clients (Web, Mobile) et communique avec les microservices en gRPC.

## 🎯 Responsabilités

- Exposer une API GraphQL publique
- Authentification et autorisation (JWT)
- Rate limiting
- Routing vers les microservices gRPC
- Agrégation de données

## 🏗️ Architecture

```
┌──────────────────┐
│  Web Dashboard   │
│  Mobile App      │
└────────┬─────────┘
         │ GraphQL (HTTP)
    ┌────▼────────────────┐
    │   API Gateway       │
    │   Port: 8080        │
    │   Protocol: HTTP    │
    │   API: GraphQL      │
    └────────┬────────────┘
             │ gRPC (interne)
    ┌────────┼────────────┐
    ▼        ▼            ▼
┌────────┐ ┌─────┐  ┌────────┐
│ Device │ │Data │  │ Notif  │
│Manager │ │Coll.│  │Service │
└────────┘ └─────┘  └────────┘
```

## 🚀 Démarrage

### 1. Installer les dépendances

```bash
cd services/api-gateway
go mod download
```

### 2. Générer le code GraphQL

```bash
# Installer gqlgen (une seule fois)
go install github.com/99designs/gqlgen@latest

# Générer le code
go run github.com/99designs/gqlgen generate
```

Cela va créer :
- `graph/generated/` : Code généré automatiquement
- `graph/model/` : Modèles Go pour GraphQL
- `graph/*.resolvers.go` : Fonctions à implémenter

### 3. Lancer le serveur

```bash
go run main.go
```

Le serveur démarre sur le port **8080**.

## 🧪 Tester l'API

### GraphQL Playground

Ouvre ton navigateur sur : http://localhost:8080

C'est une interface interactive pour tester tes requêtes GraphQL !

### Exemples de requêtes

**Créer un device :**
```graphql
mutation {
  createDevice(input: {
    name: "Capteur Température Salon"
    type: "temperature_sensor"
    metadata: [
      { key: "location", value: "salon" }
      { key: "floor", value: "1" }
    ]
  }) {
    id
    name
    type
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

**Récupérer un device :**
```graphql
query {
  device(id: "123") {
    id
    name
    type
    status
    metadata {
      key
      value
    }
  }
}
```

**Statistiques :**
```graphql
query {
  stats {
    totalDevices
    onlineDevices
    offlineDevices
  }
}
```

### Health Check

```bash
curl http://localhost:8080/health
```

## 📝 Structure du code

```
api-gateway/
├── main.go              # Point d'entrée
├── schema.graphql       # Schéma GraphQL
├── gqlgen.yml          # Configuration gqlgen
├── graph/
│   ├── generated/      # Code généré (ne pas modifier)
│   ├── model/          # Modèles GraphQL
│   └── resolver.go     # Implémentation des resolvers
└── README.md
```

## 🔄 Workflow de développement

1. Modifier `schema.graphql`
2. Lancer `go run github.com/99designs/gqlgen generate`
3. Implémenter les resolvers dans `graph/*.resolvers.go`
4. Tester dans GraphQL Playground

## 📝 TODO

- [ ] Implémenter les resolvers
- [ ] Connexion gRPC au Device Manager
- [ ] Authentification JWT
- [ ] Rate limiting
- [ ] Métriques Prometheus
- [ ] Tests
