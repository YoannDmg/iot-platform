# Tests End-to-End (E2E)

Tests de bout en bout qui valident le fonctionnement complet de la plateforme IoT avec tous les services démarrés.

## Architecture

```
tests/e2e/
├── setup_test.go          # Setup: démarre tous les services
├── auth_flow_test.go      # Tests: register → login → auth
├── device_flow_test.go    # Tests: CRUD complet devices
├── permissions_test.go    # Tests: RBAC admin vs user
└── README.md
```

## Prérequis

Les tests E2E nécessitent:

1. **PostgreSQL** (via docker-compose)
2. **Tous les services compilés** (fait automatiquement)
3. **Database migrée** (voir ci-dessous)

## Lancement

### 1. Démarrer l'infrastructure

```bash
# Terminal 1: Démarre PostgreSQL
make up

# Vérifie que PostgreSQL tourne
make status

# Lance les migrations
make db-migrate
```

### 2. Lancer les tests E2E

```bash
# Lance tous les tests E2E
make test-e2e

# Ou directement
cd tests/e2e
go test -tags=e2e -v -timeout=5m ./...
```

### 3. Tests avec logs détaillés

```bash
# Mode verbose
go test -tags=e2e -v ./...

# Logs de tous les services (même si tests passent)
go test -tags=e2e -v ./... -args -test.v
```

## Ce que testent les E2E

### ✅ `auth_flow_test.go` - Authentification

- [x] Register nouveau user
- [x] Login avec credentials valides
- [x] Login avec password invalide → FAIL
- [x] Accès ressource protégée avec token → SUCCESS
- [x] Accès ressource protégée sans token → FAIL
- [x] Duplicate registration → FAIL

### ✅ `device_flow_test.go` - Lifecycle Devices

- [x] Create device avec auth
- [x] Get device par ID
- [x] List devices paginé
- [x] Update device (name, status, metadata)
- [x] Delete device
- [x] Vérification device supprimé
- [x] Operations sans auth → FAIL

### ✅ `permissions_test.go` - RBAC

- [x] User normal crée device
- [x] User voit ses propres devices
- [x] User update ses devices
- [x] Admin voit TOUS les devices
- [x] Admin update n'importe quel device
- [x] Admin delete n'importe quel device
- [x] Admin liste tous les users

## Comment ça marche

### 1. Setup Automatique

Le setup (`SetupE2EEnvironment`) fait automatiquement:

```go
1. Clean la database
2. Build tous les binaires (bin/device-manager, bin/user-service, bin/api-gateway)
3. Démarre Device Manager sur port 18081
4. Démarre User Service sur port 18083
5. Démarre API Gateway sur port 18080
6. Attend que tous soient ready (health checks)
7. Lance les tests
8. Cleanup: kill tous les processus
```

### 2. Isolation

Chaque test:
- Utilise des ports dédiés (180xx au lieu de 80xx)
- Clean la DB avant de commencer
- Est indépendant des autres tests
- Cleanup automatique en cas d'échec

### 3. Debugging

En cas d'échec, les logs des 3 services sont affichés:

```bash
=== Device Manager Logs ===
[device-manager] Server started on :18081...

=== User Service Logs ===
[user-service] Server started on :18083...

=== API Gateway Logs ===
[api-gateway] Server started on :18080...
```

## Troubleshooting

### Erreur: "Failed to connect to PostgreSQL"

```bash
# Vérifier que PostgreSQL tourne
make status

# Redémarrer si besoin
make down
make up
make db-migrate
```

### Erreur: "Port already in use"

Les tests utilisent les ports 18080, 18081, 18083. Si occupés:

```bash
# Trouver le processus
lsof -i :18080

# Kill si nécessaire
kill -9 <PID>
```

### Erreur: "Build failed"

```bash
# Rebuild manuellement
make build

# Ou clean puis rebuild
make clean
make build
```

### Tests timeout

```bash
# Augmenter le timeout
go test -tags=e2e -v -timeout=10m ./...
```

## Ajouter de nouveaux tests E2E

### 1. Créer un nouveau fichier

```go
// +build e2e

package e2e

import (
    "testing"
)

func TestE2E_MonNouveauScenario(t *testing.T) {
    env := SetupE2EEnvironment(t)

    client := &http.Client{}
    gatewayURL := "http://" + env.APIGatewayAddr + "/query"

    // Votre test ici
}
```

### 2. Utiliser les helpers

```go
// Envoyer une requête GraphQL
resp := graphqlRequest(t, client, gatewayURL, query, token)

// Vérifier les données
data := resp["data"].(map[string]interface{})
```

### 3. Lancer

```bash
go test -tags=e2e -v ./... -run TestE2E_MonNouveauScenario
```

## CI/CD

Pour intégrer dans CI/CD (GitHub Actions, GitLab CI, etc.):

```yaml
- name: Start Infrastructure
  run: make up && make db-migrate

- name: Run E2E Tests
  run: make test-e2e
  timeout-minutes: 10

- name: Cleanup
  if: always()
  run: make down
```

## Performance

Les tests E2E sont **lents** (~30s-1min) car ils:
- Démarrent 3 services réels
- Utilisent une vraie DB PostgreSQL
- Font de vraies requêtes HTTP/gRPC

**Recommandation**: Lancer en CI uniquement sur `main` branch ou avant merge.

## Prochaines étapes

Tests à ajouter:
- [ ] Token expiration & refresh
- [ ] Pagination avancée
- [ ] Filtres et recherche
- [ ] Upload de fichiers
- [ ] WebSocket/SSE pour real-time
- [ ] Rate limiting
- [ ] Concurrent users
