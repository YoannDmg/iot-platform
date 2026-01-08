.PHONY: help start stop restart logs clean build test

help: ## Afficher l'aide
	@echo "Commandes disponibles :"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Infrastructure
start: ## Démarrer l'infrastructure locale (Docker)
	docker-compose up -d
	@echo "✅ Infrastructure démarrée"
	@echo "PostgreSQL: localhost:5432"
	@echo "Redis: localhost:6379"
	@echo "MQTT: localhost:1883"
	@echo "Prometheus: http://localhost:9090"
	@echo "Grafana: http://localhost:3000 (admin/admin)"

stop: ## Arrêter l'infrastructure
	docker-compose down
	@echo "✅ Infrastructure arrêtée"

restart: ## Redémarrer l'infrastructure
	docker-compose restart
	@echo "✅ Infrastructure redémarrée"

logs: ## Voir les logs de l'infrastructure
	docker-compose logs -f

clean: ## Nettoyer les volumes et containers
	docker-compose down -v
	@echo "✅ Volumes et containers supprimés"

# Services
api-gateway: ## Démarrer l'API Gateway
	cd services/api-gateway && go run main.go

device-manager: ## Démarrer le Device Manager
	cd services/device-manager && go run main.go

data-collector: ## Démarrer le Data Collector
	cd services/data-collector && cargo run

notification-service: ## Démarrer le Notification Service
	cd services/notification-service && go run main.go

# Frontend
web: ## Démarrer le dashboard web
	cd frontends/web-dashboard && npm run dev

mobile: ## Démarrer l'app mobile (Flutter)
	cd frontends/mobile-app && flutter run

# Développement
install-tools: ## Installer les outils nécessaires (protoc, gqlgen, etc.)
	@echo "📦 Installation des outils..."
	@command -v protoc >/dev/null 2>&1 || (echo "❌ protoc non installé. Run: brew install protobuf" && exit 1)
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/99designs/gqlgen@latest
	@echo "✅ Outils installés"

generate-proto: ## Générer le code Protocol Buffers
	@echo "🔧 Génération du code proto..."
	cd shared/proto && ./generate.sh
	@echo "✅ Code proto généré"

generate-graphql: ## Générer le code GraphQL
	@echo "🔧 Génération du code GraphQL..."
	cd services/api-gateway && go run github.com/99designs/gqlgen generate
	@echo "✅ Code GraphQL généré"

generate: generate-proto generate-graphql ## Générer tout le code (proto + GraphQL)

install-go-deps: ## Installer les dépendances Go
	@echo "📦 Installation des dépendances Go..."
	cd services/api-gateway && go mod download
	cd services/device-manager && go mod download
	@echo "✅ Dépendances Go installées"

install-rust-deps: ## Installer les dépendances Rust
	cd services/data-collector && cargo build

install-web-deps: ## Installer les dépendances web
	cd frontends/web-dashboard && npm install

setup: install-tools install-go-deps ## Configuration initiale (outils + dépendances)
	@echo "✅ Setup terminé"

init: setup start ## Initialiser le projet (première fois)
	@echo "🚀 Initialisation du projet..."
	@echo "⏳ Attente du démarrage de l'infrastructure..."
	@sleep 10
	@echo "✅ Projet initialisé"

# Tests
test: ## Lancer tous les tests
	@echo "Running tests..."
	cd services/api-gateway && go test ./...
	cd services/device-manager && go test ./...
	cd services/data-collector && cargo test
	cd frontends/web-dashboard && npm test

# Status
status: ## Voir le statut des services
	@docker-compose ps
