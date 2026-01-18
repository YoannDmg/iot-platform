.PHONY: help start dev infra services test build clean docs docs-build

# Variables
SERVICES := device-manager api-gateway user-service telemetry-collector
PROTO_DIR := shared/proto
BIN_DIR := bin
DASHBOARD_DIR := frontends/dashboard
SCRIPTS_DIR := scripts
MIGRATIONS_DIR := infrastructure/database/migrations

# Infrastructure services (Docker)
INFRA_SERVICES := postgres redis mosquitto prometheus grafana

# Database migrations directory
# Migrations are auto-discovered from files matching [0-9]*.sql (excluding 000_)

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
	@echo "🚀 DÉMARRAGE RAPIDE"
	@echo "  \033[36mmake dev\033[0m              Développement (infra Docker + services locaux)"
	@echo "  \033[36mmake start\033[0m            Tout en Docker (infra + services)"
	@echo "  \033[36mmake down\033[0m             Arrête tout (conserve les données)"
	@echo ""
	@echo "🐳 INFRASTRUCTURE (Postgres, Redis, MQTT, Prometheus, Grafana)"
	@echo "  \033[36minfra\033[0m                 Démarre l'infrastructure"
	@echo "  \033[36minfra-down\033[0m            Arrête (conserve les données)"
	@echo "  \033[36minfra-destroy\033[0m         Arrête et SUPPRIME les données"
	@echo "  \033[36minfra-logs\033[0m            Logs de l'infrastructure"
	@echo "  \033[36minfra-status\033[0m          Statut des containers"
	@echo ""
	@echo "🗄️  BASE DE DONNÉES"
	@echo "  \033[36mdb-migrate\033[0m            Applique les migrations (avec suivi)"
	@echo "  \033[36mdb-migrations\033[0m         Affiche l'état des migrations"
	@echo "  \033[36mdb-reset\033[0m              Réinitialise (SUPPRIME les données)"
	@echo "  \033[36mdb-status\033[0m             Affiche l'état des tables"
	@echo ""
	@echo "💻 DÉVELOPPEMENT (services en local, infra Docker)"
	@echo "  \033[36mdev\033[0m                   Infra + migrations + services Go"
	@echo "  \033[36mdev-dashboard\033[0m         Dev + dashboard React"
	@echo "  \033[36mdev-api\033[0m               API Gateway seul"
	@echo "  \033[36mdev-devices\033[0m           Device Manager seul"
	@echo "  \033[36mdev-users\033[0m             User Service seul"
	@echo "  \033[36mdev-telemetry\033[0m         Telemetry Collector seul"
	@echo ""
	@echo "📦 SERVICES DOCKER (api-gateway, device-manager, user-service, telemetry-collector)"
	@echo "  \033[36mservices\033[0m              Démarre les services (nécessite infra)"
	@echo "  \033[36mservices-down\033[0m         Arrête les services"
	@echo "  \033[36mservices-logs\033[0m         Logs des services"
	@echo "  \033[36mservices-rebuild\033[0m      Rebuild et relance"
	@echo ""
	@echo "🎮 SIMULATION"
	@echo "  \033[36msimulate\033[0m              5 devices, intervalle 3s"
	@echo "  \033[36msimulate-heavy\033[0m        50 devices, 60s (stress test)"
	@echo ""
	@echo "🧪 TESTS"
	@echo "  \033[36mtest\033[0m                  Tests unitaires"
	@echo "  \033[36mtest-integration\033[0m      Tests d'intégration (nécessite DB)"
	@echo "  \033[36mtest-e2e\033[0m              Tests end-to-end"
	@echo ""
	@echo "🔨 BUILD & SETUP"
	@echo "  \033[36msetup\033[0m                 Installe les outils"
	@echo "  \033[36mgenerate\033[0m              Génère proto + GraphQL"
	@echo "  \033[36mbuild\033[0m                 Compile les services"
	@echo "  \033[36mclean\033[0m                 Nettoie"
	@echo ""
	@echo "🌐 DASHBOARD"
	@echo "  \033[36mdashboard\033[0m             Mode dev"
	@echo "  \033[36mdashboard-build\033[0m       Build production"
	@echo "  \033[36mdashboard-lint\033[0m        Lint"
	@echo ""
	@echo "📚 DOCUMENTATION"
	@echo "  \033[36mdocs\033[0m                  Mode dev"
	@echo "  \033[36mdocs-build\033[0m            Build"
	@echo ""
	@echo "🛠️  UTILS"
	@echo "  \033[36mfmt\033[0m                   Formate le code Go"
	@echo "  \033[36mlint\033[0m                  Lint le code Go"
	@echo "  \033[36mdeps\033[0m                  Met à jour les dépendances"
	@echo ""

#==================================================================================
# DÉMARRAGE RAPIDE
#==================================================================================

start: infra db-migrate services ## Tout en Docker (infra + services)
	@echo ""
	@echo "✅ Plateforme démarrée!"
	@echo ""
	@echo "📍 Services disponibles:"
	@echo "  API Gateway:         http://localhost:8080"
	@echo "  GraphQL Playground:  http://localhost:8080/"
	@echo "  Grafana:             http://localhost:3000"

down: ## Arrête tout (conserve les données)
	@echo "🛑 Arrêt de la plateforme..."
	@docker-compose stop
	@echo "✅ Plateforme arrêtée (données conservées)"

dev: infra db-migrate ## Développement (infra Docker + services locaux)
	@echo ""
	@echo "🚀 Démarrage des services en mode développement..."
	@echo ""
	@echo "📍 Services:"
	@echo "  Device Manager:      localhost:8081 (gRPC)"
	@echo "  User Service:        localhost:8082 (gRPC)"
	@echo "  Telemetry Collector: localhost:8083 (gRPC + MQTT)"
	@echo "  API Gateway:         http://localhost:8080 (GraphQL)"
	@echo ""
	@echo "⚠️  Ctrl+C pour arrêter"
	@echo ""
	@trap 'echo "\n🛑 Arrêt des services..."; kill 0' INT; \
	(cd services/device-manager && go run main.go) & \
	(sleep 2 && cd services/user-service && go run main.go) & \
	(sleep 3 && cd services/telemetry-collector && go run main.go) & \
	(sleep 5 && cd services/api-gateway && go run main.go) & \
	wait

dev-dashboard: infra db-migrate ## Dev + dashboard React
	@echo ""
	@echo "🚀 Démarrage complet (services + dashboard)..."
	@echo ""
	@echo "📍 Services:"
	@echo "  API Gateway:  http://localhost:8080"
	@echo "  Dashboard:    http://localhost:5173"
	@echo "  Grafana:      http://localhost:3000"
	@echo ""
	@echo "⚠️  Ctrl+C pour arrêter"
	@echo ""
	@trap 'echo "\n🛑 Arrêt des services..."; kill 0' INT; \
	(cd services/device-manager && go run main.go) & \
	(sleep 2 && cd services/user-service && go run main.go) & \
	(sleep 3 && cd services/telemetry-collector && go run main.go) & \
	(sleep 5 && cd services/api-gateway && go run main.go) & \
	(sleep 7 && cd $(DASHBOARD_DIR) && npm run dev) & \
	wait

# Services individuels (pour debug)
dev-api: ## API Gateway seul
	@cd services/api-gateway && go run main.go

dev-devices: ## Device Manager seul
	@cd services/device-manager && go run main.go

dev-users: ## User Service seul
	@cd services/user-service && go run main.go

dev-telemetry: ## Telemetry Collector seul
	@cd services/telemetry-collector && go run main.go

#==================================================================================
# INFRASTRUCTURE
#==================================================================================

infra: ## Démarre l'infrastructure
	@echo "🐳 Démarrage de l'infrastructure..."
	@docker-compose up -d $(INFRA_SERVICES)
	@echo "⏳ Attente que PostgreSQL soit prêt..."
	@until docker-compose exec -T postgres pg_isready -U iot_user -d iot_platform >/dev/null 2>&1; do sleep 1; done
	@echo "✅ Infrastructure prête!"
	@echo ""
	@echo "📍 Services:"
	@echo "  PostgreSQL:  localhost:5432"
	@echo "  Redis:       localhost:6379"
	@echo "  MQTT:        localhost:1883"
	@echo "  Prometheus:  http://localhost:9090"
	@echo "  Grafana:     http://localhost:3000"

infra-down: ## Arrête l'infrastructure (conserve les données)
	@echo "🛑 Arrêt de l'infrastructure..."
	@docker-compose stop $(INFRA_SERVICES)
	@echo "✅ Infrastructure arrêtée (données conservées)"

infra-destroy: ## Arrête et SUPPRIME les données
	@echo "⚠️  Cela va SUPPRIMER toutes les données!"
	@read -p "   Continuer? [y/N] " confirm && [ "$$confirm" = "y" ] || (echo "Annulé." && exit 1)
	@docker-compose down -v
	@echo "✅ Infrastructure supprimée"

infra-logs: ## Logs de l'infrastructure
	@docker-compose logs -f $(INFRA_SERVICES)

infra-status: ## Statut des containers
	@docker-compose ps $(INFRA_SERVICES)

#==================================================================================
# SERVICES DOCKER
#==================================================================================

services: ## Démarre les services (nécessite infra)
	@echo "📦 Démarrage des services..."
	@docker-compose up -d --build api-gateway device-manager user-service telemetry-collector
	@echo "✅ Services démarrés!"

services-down: ## Arrête les services
	@echo "🛑 Arrêt des services..."
	@docker-compose stop api-gateway device-manager user-service telemetry-collector
	@echo "✅ Services arrêtés"

services-logs: ## Logs des services
	@docker-compose logs -f api-gateway device-manager user-service telemetry-collector

services-rebuild: ## Rebuild et relance
	@echo "🔨 Rebuild des services..."
	@docker-compose up -d --build api-gateway device-manager user-service telemetry-collector
	@echo "✅ Services reconstruits!"

#==================================================================================
# BASE DE DONNÉES
#==================================================================================

db-migrate: ## Applique les migrations (avec suivi)
	@echo "🗄️  Application des migrations..."
	@# Bootstrap: create schema_migrations table (idempotent)
	@docker-compose exec -T postgres psql -U iot_user -d iot_platform < $(MIGRATIONS_DIR)/000_create_schema_migrations.sql >/dev/null 2>&1
	@# Auto-discover and apply migrations (sorted by filename, excluding 000_)
	@for file in $$(ls $(MIGRATIONS_DIR)/[0-9]*.sql 2>/dev/null | grep -v '000_' | sort); do \
		migration=$$(basename "$$file" .sql); \
		if docker-compose exec -T postgres psql -U iot_user -d iot_platform -tAc \
			"SELECT 1 FROM schema_migrations WHERE version = '$$migration'" 2>/dev/null | grep -q 1; then \
			echo "  ⏭️  $$migration (déjà appliquée)"; \
		else \
			echo "  📦 $$migration..."; \
			docker-compose exec -T postgres psql -U iot_user -d iot_platform < "$$file" || exit 1; \
			docker-compose exec -T postgres psql -U iot_user -d iot_platform -c \
				"INSERT INTO schema_migrations (version) VALUES ('$$migration')" >/dev/null; \
		fi; \
	done
	@echo "✅ Migrations à jour!"

db-reset: ## Réinitialise (SUPPRIME les données)
	@echo "⚠️  Réinitialisation de la base de données..."
	@echo "   Cela va SUPPRIMER toutes les données!"
	@read -p "   Continuer? [y/N] " confirm && [ "$$confirm" = "y" ] || (echo "Annulé." && exit 1)
	@docker-compose exec -T postgres psql -U iot_user -d iot_platform -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@echo "📦 Ré-application des migrations..."
	@$(MAKE) db-migrate
	@echo "✅ Base réinitialisée!"

db-status: ## Affiche l'état des tables et données
	@echo "🔍 État de la base de données..."
	@docker-compose exec -T postgres psql -U iot_user -d iot_platform -c "\dt"
	@echo ""
	@docker-compose exec -T postgres psql -U iot_user -d iot_platform -c "SELECT 'devices' as table_name, COUNT(*) FROM devices UNION ALL SELECT 'users', COUNT(*) FROM users UNION ALL SELECT 'device_telemetry', COUNT(*) FROM device_telemetry;" 2>/dev/null || echo "Tables non créées - lancez 'make db-migrate'"

db-migrations: ## Affiche l'état des migrations
	@echo "🔍 État des migrations..."
	@echo ""
	@for file in $$(ls $(MIGRATIONS_DIR)/[0-9]*.sql 2>/dev/null | grep -v '000_' | sort); do \
		migration=$$(basename "$$file" .sql); \
		if docker-compose exec -T postgres psql -U iot_user -d iot_platform -tAc \
			"SELECT applied_at FROM schema_migrations WHERE version = '$$migration'" 2>/dev/null | grep -q .; then \
			applied=$$(docker-compose exec -T postgres psql -U iot_user -d iot_platform -tAc \
				"SELECT applied_at FROM schema_migrations WHERE version = '$$migration'" 2>/dev/null); \
			echo "  ✅ $$migration (appliquée: $$applied)"; \
		else \
			echo "  ⏳ $$migration (en attente)"; \
		fi; \
	done

#==================================================================================
# SIMULATION
#==================================================================================

simulate: ## 5 devices, intervalle 3s
	@echo "🎮 Démarrage du simulateur..."
	@echo "   (Ctrl+C pour arrêter)"
	@cd $(SCRIPTS_DIR) && go run simulate-devices.go -devices 5 -interval 3

simulate-heavy: ## 50 devices, 60s (stress test)
	@echo "🎮 Stress test (50 devices, 60s)..."
	@cd $(SCRIPTS_DIR) && go run simulate-devices.go -devices 50 -interval 1 -duration 60

#==================================================================================
# TESTS
#==================================================================================

test: ## Tests unitaires
	@echo "🧪 Tests unitaires..."
	@for service in $(SERVICES); do \
		echo "  → $$service"; \
		(cd services/$$service && go test -tags=unit ./... -v) || exit 1; \
	done

test-integration: ## Tests d'intégration (nécessite DB)
	@echo "🗄️  Tests d'intégration..."
	@cd services/device-manager && go test -tags=integration ./storage/... -v
	@cd services/user-service && go test -tags=integration ./storage/... -v

test-e2e: ## Tests end-to-end
	@echo "🎯 Tests E2E..."
	@echo "⚠️  Assurez-vous que la plateforme tourne: make start"
	@cd tests/e2e && go test -tags=e2e -v -timeout=5m ./...

#==================================================================================
# BUILD & SETUP
#==================================================================================

setup: ## Installe les outils
	@echo "📦 Installation des outils..."
	@command -v protoc >/dev/null 2>&1 || (echo "❌ protoc requis: brew install protobuf" && exit 1)
	@command -v node >/dev/null 2>&1 || (echo "❌ Node.js requis: brew install node" && exit 1)
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/99designs/gqlgen@latest
	@cd $(DASHBOARD_DIR) && npm install
	@echo "✅ Setup terminé!"

generate: ## Génère proto + GraphQL
	@echo "🔨 Génération du code..."
	@cd $(PROTO_DIR) && ./generate.sh
	@cd services/api-gateway && gqlgen generate
	@echo "✅ Code généré!"

build: ## Compile les services
	@echo "🔨 Compilation..."
	@mkdir -p $(BIN_DIR)
	@for service in $(SERVICES); do \
		echo "  → $$service"; \
		(cd services/$$service && go build -o ../../$(BIN_DIR)/$$service) || exit 1; \
	done
	@echo "✅ Binaires dans ./$(BIN_DIR)/"

clean: ## Nettoie
	@echo "🧹 Nettoyage..."
	@rm -rf $(BIN_DIR)/
	@rm -rf $(DASHBOARD_DIR)/dist $(DASHBOARD_DIR)/node_modules
	@echo "✅ Nettoyé!"

#==================================================================================
# DASHBOARD
#==================================================================================

dashboard: ## Mode dev
	@echo "🌐 Dashboard: http://localhost:5173"
	@cd $(DASHBOARD_DIR) && npm run dev

dashboard-build: ## Build production
	@cd $(DASHBOARD_DIR) && npm run build
	@echo "✅ Build dans $(DASHBOARD_DIR)/dist/"

dashboard-lint: ## Lint
	@cd $(DASHBOARD_DIR) && npm run lint

#==================================================================================
# DOCUMENTATION
#==================================================================================

docs: ## Mode dev
	@echo "📚 Documentation: http://localhost:3001"
	@cd docs && npm start -- --port 3001

docs-build: ## Build
	@cd docs && npm run build
	@echo "✅ Build dans docs/build/"

#==================================================================================
# UTILS
#==================================================================================

fmt: ## Formate le code Go
	@gofmt -w services/

lint: ## Lint le code Go
	@for service in $(SERVICES); do \
		(cd services/$$service && golangci-lint run) || exit 1; \
	done

deps: ## Met à jour les dépendances
	@for service in $(SERVICES); do \
		(cd services/$$service && go mod tidy) || exit 1; \
	done
	@echo "✅ Dépendances à jour!"
