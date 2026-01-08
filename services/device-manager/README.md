# Device Manager Service

Service gRPC de gestion des devices IoT.

## 🎯 Responsabilités

- Créer, lire, mettre à jour, supprimer des devices (CRUD)
- Gérer le statut des devices (online, offline, error)
- Stocker les métadonnées des devices
- Exposer une API gRPC pour les autres services

## 🏗️ Architecture

```
┌─────────────────┐
│  API Gateway    │
│   (GraphQL)     │
└────────┬────────┘
         │ gRPC
    ┌────▼────────────────┐
    │  Device Manager     │
    │  Port: 8081         │
    │  Protocol: gRPC     │
    └────────┬────────────┘
             │
        ┌────▼────┐
        │   DB    │
        └─────────┘
```

## 🚀 Démarrage

### Installer les dépendances

```bash
cd services/device-manager
go mod download
```

### Générer le code Protocol Buffers (IMPORTANT)

```bash
cd ../../shared/proto
./generate.sh
```

### Lancer le service

```bash
go run main.go
```

Le service écoute sur le port **8081** en gRPC.

## 📡 API gRPC

### Méthodes disponibles

1. **CreateDevice** - Créer un nouveau device
2. **GetDevice** - Récupérer un device par ID
3. **ListDevices** - Lister tous les devices
4. **UpdateDevice** - Mettre à jour un device
5. **DeleteDevice** - Supprimer un device
6. **WatchDevices** - Stream temps réel des changements

## 🧪 Tester le service

### Avec grpcurl (outil CLI pour gRPC)

**Installation:**
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
  "name": "Capteur Température",
  "type": "temperature_sensor",
  "metadata": {"location": "salon"}
}' localhost:8081 device.DeviceService/CreateDevice
```

## 📝 TODO

- [ ] Connexion à PostgreSQL
- [ ] Gestion de la persistence
- [ ] Authentification gRPC
- [ ] Métriques Prometheus
- [ ] Tests unitaires
