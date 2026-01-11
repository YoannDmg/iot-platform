.PHONY: help setup generate clean build dev test docs-dev docs-build docs-serve docs-clean

# Variables
SERVICES := device-manager api-gateway user-service
PROTO_DIR := shared/proto
BIN_DIR := bin
DASHBOARD_DIR := web/dashboard
MAKEFILE := Makefile

# Load .env file if it exists
-include .env
export

#==================================================================================
# HELP
#==================================================================================

help: ## Affiche l'aide
	@echo ""
	@echo "╔════════════════════════════════════════════════════════════════╗"
	@echo "║                 IoT Platform - Commandes Make                  ║"
	@echo "╚════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "📦 SETUP & GÉNÉRATION"
	@grep -E '^(setup|generate|generate-proto|generate-graphql):.*?## .*$$' $(MAKEFILE) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "🔨 BUILD & CLEAN"
	@grep -E '^(build|clean):.*?## .*$$' $(MAKEFILE) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "🐳 DOCKER"
	@grep -E '^(up|down|logs|status|restart):.*?## .*$$' $(MAKEFILE) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "🚀 SERVICES (DEV MODE)"
	@grep -E '^(device-manager|api-gateway|user-service|dashboard|dev|dev-full):.*?## .*$$' $(MAKEFILE) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "🧪 TESTS"
	@grep -E '^(test|test-unit|test-integration|test-e2e|test-all|test-security|test-device|test-user|test-api):.*?## .*$$' $(MAKEFILE) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "🗄️  DATABASE"
	@grep -E '^(db-migrate|db-reset|db-status|sqlc-generate):.*?## .*$$' $(MAKEFILE) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "📚 DOCUMENTATION"
	@grep -E '^(docs-dev|docs-build|docs-serve|docs-clean):.*?## .*$$' $(MAKEFILE) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "🌐 WEB DASHBOARD"
	@grep -E '^(dashboard-install|dashboard-dev|dashboard-build|dashboard-preview|dashboard-clean|dashboard-lint):.*?## .*$$' $(MAKEFILE) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "🛠️  UTILS"
	@grep -E '^(deps|fmt|lint):.*?## .*$$' $(MAKEFILE) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""

#==================================================================================
# SETUP & GÉNÉRATION
#==================================================================================

setup: ## Installe tous les outils nécessaires (Go + Node)
	@echo "📦 Installation des outils..."
	@command -v protoc >/dev/null 2>&1 || (echo "❌ protoc non installé. Installez-le avec: brew install protobuf" && exit 1)
	@command -v node >/dev/null 2>&1 || (echo "❌ Node.js non installé. Installez-le avec: brew install node" && exit 1)
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/99designs/gqlgen@latest
	@echo "📦 Installation des dépendances du dashboard..."
	@cd $(DASHBOARD_DIR) && npm install
	@echo "✅ Setup terminé!"

generate: generate-proto generate-graphql ## Génère tout le code (proto + GraphQL)

generate-proto: ## Génère le code Protobuf
	@echo "🔨 Génération du code Protobuf..."
	@cd $(PROTO_DIR) && ./generate.sh
	@echo "✅ Proto généré!"

generate-graphql: ## Génère le code GraphQL
	@echo "🔨 Génération du code GraphQL..."
	@cd services/api-gateway && gqlgen generate
	@echo "✅ GraphQL généré!"

#==================================================================================
# BUILD & CLEAN
#==================================================================================

build: ## Compile tous les services
	@echo "🔨 Compilation de tous les services..."
	@mkdir -p $(BIN_DIR)
	@for service in $(SERVICES); do \
		echo "  → Building $$service..."; \
		cd services/$$service && go build -o ../../$(BIN_DIR)/$$service && cd ../..; \
	done
	@echo "✅ Build terminé! Binaires dans ./$(BIN_DIR)/"

clean: ## Supprime les binaires et fichiers temporaires
	@echo "🧹 Nettoyage..."
	@rm -rf $(BIN_DIR)/
	@rm -f services/device-manager/device-manager
	@rm -f services/api-gateway/api-gateway
	@$(MAKE) dashboard-clean
	@echo "✅ Nettoyage terminé!"

#==================================================================================
# DOCKER
#==================================================================================

up: ## Lance l'infrastructure Docker (Postgres, Redis, MQTT)
	@echo "🐳 Démarrage de l'infrastructure..."
	@docker-compose up -d
	@echo "✅ Infrastructure démarrée!"
	@echo "PostgreSQL: localhost:5432"
	@echo "Redis: localhost:6379"
	@echo "MQTT: localhost:1883"

down: ## Arrête l'infrastructure Docker
	@echo "🛑 Arrêt de l'infrastructure..."
	@docker-compose down
	@echo "✅ Infrastructure arrêtée!"

logs: ## Affiche les logs Docker
	@docker-compose logs -f

status: ## Affiche le status de l'infrastructure
	@docker-compose ps

restart: ## Redémarre l'infrastructure
	@docker-compose restart
	@echo "✅ Infrastructure redémarrée!"

#==================================================================================
# SERVICES (DEV MODE)
#==================================================================================

device-manager: ## Lance le Device Manager
	@echo "Démarrage du Device Manager..."
	@cd services/device-manager && go run main.go

api-gateway: ## Lance l'API Gateway
	@echo "Démarrage de l'API Gateway..."
	@cd services/api-gateway && go run main.go

user-service: ## Lance le User Service
	@echo "Démarrage du User Service..."
	@cd services/user-service && go run main.go

dashboard: ## Lance le Dashboard (dev mode avec HMR)
	@echo "🌐 Démarrage du Dashboard..."
	@echo "📍 URL: http://localhost:5173"
	@cd $(DASHBOARD_DIR) && npm run dev

dev: up ## Lance infra + services backend (sans dashboard)
	@echo "Démarrage complet de la plateforme..."
	@echo ""
	@echo "⏳ Attente de l'infrastructure Docker..."
	@sleep 5
	@echo "✅ Infrastructure prête!"
	@echo ""
	@echo "⚠️  Utilise Ctrl+C pour arrêter tous les services."
	@echo ""
	@trap 'echo "\n🛑 Arrêt des services..."; kill 0' INT; \
	$(MAKE) device-manager & \
	(sleep 2 && $(MAKE) user-service) & \
	(sleep 4 && $(MAKE) api-gateway) & \
	wait

dev-full: up ## Lance TOUT: infra + services + dashboard
	@echo "🚀 Démarrage COMPLET de la plateforme..."
	@echo ""
	@echo "⏳ Attente de l'infrastructure Docker..."
	@sleep 5
	@echo "✅ Infrastructure prête!"
	@echo ""
	@echo "📍 Services:"
	@echo "  - API Gateway: http://localhost:8080"
	@echo "  - Dashboard:   http://localhost:5173"
	@echo ""
	@echo "⚠️  Utilise Ctrl+C pour arrêter tous les services."
	@echo ""
	@trap 'echo "\n🛑 Arrêt des services..."; kill 0' INT; \
	$(MAKE) device-manager & \
	(sleep 2 && $(MAKE) user-service) & \
	(sleep 4 && $(MAKE) api-gateway) & \
	(sleep 6 && $(MAKE) dashboard) & \
	wait

#==================================================================================
# TESTS
#==================================================================================

.PHONY: test test-unit test-integration test-e2e test-all \
        test-device test-user test-api test-security

# ------------------------------------------------------------------------------
# UNIT TESTS (~5s)
# ------------------------------------------------------------------------------

test-unit: ## Tests unitaires (parallèle, rapide)
	@echo "🧪 Unit tests..."
	@$(MAKE) -j3 test-device-unit test-user-unit test-api-unit

test-device-unit: ## Tests unitaires Device Manager
	@cd services/device-manager && go test -tags=unit ./... -v

test-user-unit: ## Tests unitaires User Service
	@cd services/user-service && go test -tags=unit ./... -v

test-api-unit: ## Tests unitaires API Gateway
	@cd services/api-gateway && go test -tags=unit ./... -v

# ------------------------------------------------------------------------------
# INTEGRATION TESTS (~30s, nécessite DB)
# ------------------------------------------------------------------------------

test-integration: ## Tests d'intégration PostgreSQL
	@echo "🗄️  Integration tests..."
	@$(MAKE) test-device-db test-user-db

test-device-db: ## Tests DB Device Manager
	@cd services/device-manager && go test -tags=integration ./storage/... -v

test-user-db: ## Tests DB User Service
	@cd services/user-service && go test -tags=integration ./storage/... -v

# ------------------------------------------------------------------------------
# E2E TESTS (~1-2min, full system)
# ------------------------------------------------------------------------------

test-e2e: ## Tests end-to-end (tous les services)
	@echo "🎯 E2E tests..."
	@echo "⚠️  Assurez-vous que PostgreSQL tourne: make up && make db-migrate"
	@cd tests/e2e && go test -tags=e2e -v -timeout=5m ./...

# ------------------------------------------------------------------------------
# SECURITY TESTS
# ------------------------------------------------------------------------------

test-security: ## Tests de sécurité JWT
	@echo "🛡️  Security tests..."
	@cd services/api-gateway && go test -tags=unit ./auth/... -v -run "Security"

# ------------------------------------------------------------------------------
# CI/CD
# ------------------------------------------------------------------------------

test-all: test-unit test-integration test-e2e ## Tous les tests (CI)

# ------------------------------------------------------------------------------
# LEGACY (compatibility)
# ------------------------------------------------------------------------------

test-device: test-device-unit ## Alias: tests Device Manager
test-user: test-user-unit ## Alias: tests User Service
test-api: test-api-unit ## Alias: tests API Gateway

#==================================================================================
# DATABASE
#==================================================================================

db-migrate: ## Lance les migrations PostgreSQL
	@echo "🗄️  Lancement des migrations..."
	@docker-compose exec -T postgres psql -U iot_user -d iot_platform < infrastructure/database/migrations/001_create_devices_table.sql
	@docker-compose exec -T postgres psql -U iot_user -d iot_platform < infrastructure/database/migrations/002_create_users_table.sql
	@echo "✅ Migrations terminées!"

db-reset: ## Réinitialise la base de données
	@echo "🗑️  Réinitialisation de la base..."
	@docker-compose exec -T postgres psql -U iot_user -d iot_platform -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@$(MAKE) db-migrate
	@echo "✅ Base réinitialisée!"

db-status: ## Vérifie le statut de la base
	@echo "🔍 Statut de la base de données..."
	@docker-compose exec -T postgres psql -U iot_user -d iot_platform -c "\dt"

sqlc-generate: ## Génère le code sqlc
	@echo "🔨 Génération du code sqlc..."
	@cd services/device-manager && sqlc generate
	@cd services/user-service && sqlc generate
	@echo "✅ Code sqlc généré!"

#==================================================================================
# UTILS
#==================================================================================

deps: ## Met à jour les dépendances Go
	@echo "📦 Mise à jour des dépendances..."
	@for service in $(SERVICES); do \
		echo "  → $$service"; \
		(cd services/$$service && go mod tidy) || exit 1; \
	done
	@echo "✅ Dépendances à jour!"

fmt: ## Formate le code Go
	@echo "✨ Formatage du code..."
	@gofmt -w services/

lint: ## Lint le code (nécessite golangci-lint)
	@echo "🔍 Linting..."
	@for service in $(SERVICES); do \
		echo "  → $$service"; \
		(cd services/$$service && golangci-lint run) || exit 1; \
	done

#==================================================================================
# DOCUMENTATION
#==================================================================================

docs-dev: ## Lance le serveur de documentation en mode dev
	@echo "📚 Démarrage de la documentation..."
	@echo "🌐 Disponible sur: http://localhost:3001"
	@echo ""
	@cd docs && npm start -- --port 3001

docs-build: ## Build la documentation statique pour production
	@echo "🔨 Build de la documentation..."
	@cd docs && npm run build
	@echo "✅ Documentation buildée dans docs/build/"

docs-serve: docs-build ## Sert la documentation buildée (test avant deploy)
	@echo "📖 Serving documentation buildée..."
	@echo "🌐 Disponible sur: http://localhost:3001"
	@cd docs && npm run serve -- --port 3001 --no-open

docs-clean: ## Nettoie les fichiers de build de la documentation
	@echo "🧹 Nettoyage de la documentation..."
	@rm -rf docs/build docs/.docusaurus
	@echo "✅ Documentation nettoyée"

#==================================================================================
# WEB DASHBOARD
#==================================================================================

dashboard-install: ## Installe les dépendances du dashboard
	@echo "📦 Installation des dépendances du dashboard..."
	@cd $(DASHBOARD_DIR) && npm install
	@echo "✅ Dépendances installées!"

dashboard-dev: ## Lance le dashboard en mode développement
	@echo "🌐 Démarrage du dashboard..."
	@echo "📍 URL: http://localhost:5173"
	@cd $(DASHBOARD_DIR) && npm run dev

dashboard-build: ## Build le dashboard pour production
	@echo "🔨 Build du dashboard..."
	@cd $(DASHBOARD_DIR) && npm run build
	@echo "✅ Dashboard buildé dans $(DASHBOARD_DIR)/dist/"

dashboard-preview: dashboard-build ## Preview du build de production
	@echo "👀 Preview du dashboard..."
	@cd $(DASHBOARD_DIR) && npm run preview

dashboard-clean: ## Nettoie les fichiers de build du dashboard
	@echo "🧹 Nettoyage du dashboard..."
	@rm -rf $(DASHBOARD_DIR)/dist $(DASHBOARD_DIR)/node_modules
	@echo "✅ Dashboard nettoyé"

dashboard-lint: ## Lint le code du dashboard
	@echo "🔍 Linting du dashboard..."
	@cd $(DASHBOARD_DIR) && npm run lint
	@echo "✅ Lint terminé!"
