# Portal DB API Makefile

# Color definitions
CYAN := \033[0;36m
BOLD := \033[1m
RESET := \033[0m

# ============================================================================
# HELP
# ============================================================================

.PHONY: help
.DEFAULT_GOAL := help
help: ## Show all available targets
	@echo ""
	@echo "$(BOLD)$(CYAN)🗄️ Portal DB Makefile Targets$(RESET)"
	@echo ""
	@echo "$(BOLD)=== 🚀 Service Management ===$(RESET)"
	@grep -h -E '^(postgrest-up|postgrest-down|postgrest-logs|portal-db-up|portal-db-down):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-30s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)=== 📝 API Generation ===$(RESET)"
	@grep -h -E '^(generate-openapi|generate-sdks|generate-all):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-30s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)=== 🔐 Authentication & Testing ===$(RESET)"
	@grep -h -E '^(test-auth|test-portal-app|gen-jwt|reset-dev-db):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-30s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)=== 💧 Data Hydration ===$(RESET)"
	@grep -h -E '^(hydrate-testdata|hydrate-services|hydrate-applications|hydrate-gateways|hydrate-prod):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-30s$(RESET) %s\n", $$1, $$2}'
	@echo ""

# ============================================================================
# SERVICE MANAGEMENT
# ============================================================================

.PHONY: postgrest-up
postgrest-up: ## Start PostgREST API service and PostgreSQL DB
	@echo "🚀 Starting Portal DB API services..."
	docker compose up -d
	@echo "✅ Services started!"
	@echo "   PostgreSQL DB: localhost:5435"
	@echo "   PostgREST API: http://localhost:3000"

.PHONY: postgrest-down
postgrest-down: ## Stop PostgREST API service and PostgreSQL DB
	@echo "🛑 Stopping Portal DB API services..."
	docker compose down
	@echo "✅ Services stopped!"

.PHONY: postgrest-logs
postgrest-logs: ## Show logs from PostgREST API service and PostgreSQL DB
	docker compose logs -f

.PHONY: portal-db-up
portal-db-up: ## Start Portal DB PostgreSQL container only
	@echo "🚀 Starting Portal DB PostgreSQL container..."
	docker compose up -d portal-db
	@echo "✅ PostgreSQL container started!"
	@echo "   Database: localhost:5435"
	@echo "   User: postgres"
	@echo "   Password: portal_password"

.PHONY: portal-db-down
portal-db-down: ## Stop Portal DB PostgreSQL container only
	@echo "🛑 Stopping Portal DB PostgreSQL container..."
	docker compose down portal-db
	@echo "✅ PostgreSQL container stopped!"

# ============================================================================
# API GENERATION
# ============================================================================

.PHONY: generate-openapi
generate-openapi: ## Generate OpenAPI specification from PostgREST
	@echo "📝 Generating OpenAPI specification..."
	cd api/codegen && ./generate-openapi.sh

.PHONY: generate-sdks
generate-sdks: ## Generate Go SDK from OpenAPI spec
	@echo "🔧 Generating Go SDK..."
	@echo "   Note: Requires Node.js and oapi-codegen"
	cd api/codegen && ./generate-sdks.sh

.PHONY: generate-all
generate-all: generate-openapi generate-sdks ## Generate OpenAPI spec and Go SDK
	@echo "✨ All generation tasks completed!"

# ============================================================================
# AUTHENTICATION & TESTING
# ============================================================================

.PHONY: test-auth
test-auth: ## Test JWT authentication flow
	@echo "🔐 Testing JWT authentication..."
	@if ! docker ps | grep -q portal-db-api; then \
		echo "❌ PostgREST API is not running. Please start it first:"; \
		echo "   make postgrest-up"; \
		exit 1; \
	fi
	cd api/scripts && ./test-auth.sh

.PHONY: test-portal-app
test-portal-app: ## Test portal application creation and retrieval
	@echo "🧪 Testing portal application creation..."
	@if ! docker ps | grep -q portal-db-api; then \
		echo "❌ PostgREST API is not running. Please start it first:"; \
		echo "   make postgrest-up"; \
		exit 1; \
	fi
	cd api/scripts && ./test-portal-app-creation.sh

.PHONY: gen-jwt
gen-jwt: ## Generate JWT token
	@echo "🔑 Generating JWT token..."
	@if ! docker ps | grep -q portal-db-api; then \
		echo "❌ PostgREST API is not running. Please start it first:"; \
		echo "   make postgrest-up"; \
		exit 1; \
	fi
	cd api/scripts && ./gen-jwt.sh

.PHONY: reset-dev-db
reset-dev-db: ## Reset the development database
	@echo "🔄 Resetting database..."
	rm -rf ./tmp/portal_db_data

# ============================================================================
# DATA HYDRATION
# ============================================================================
# Database connection for all hydrate scripts (matches docker-compose.yml)
DB_CONNECTION_STRING := postgresql://postgres:portal_password@localhost:5435/portal_db

.PHONY: hydrate-testdata
hydrate-testdata: ## Populate database with test data for development
	@echo "🧪 Hydrating Portal DB with test data..."
	@if ! docker ps | grep -q portal-db; then \
		echo "❌ Portal DB is not running. Please start it first:"; \
		echo "   make postgrest-up"; \
		exit 1; \
	fi
	DB_CONNECTION_STRING="$(DB_CONNECTION_STRING)" ./scripts/hydrate-testdata.sh

.PHONY: hydrate-services
hydrate-services: ## Hydrate services table from blockchain data
	@echo "🌐 Hydrating services table..."
	@if [ -z "$(NODE)" ] || [ -z "$(NETWORK)" ] || [ -z "$(SERVICE_IDS)" ]; then \
		echo "❌ Required parameters missing. Usage:"; \
		echo "   make hydrate-services NODE=<node_url> NETWORK=<network_id> SERVICE_IDS=<service_ids>"; \
		echo "   Example: make hydrate-services NODE=https://node.example.com NETWORK=pocket SERVICE_IDS=ethereum-mainnet,polygon-mainnet"; \
		exit 1; \
	fi
	@if ! docker ps | grep -q portal-db; then \
		echo "❌ Portal DB is not running. Please start it first:"; \
		echo "   make postgrest-up"; \
		exit 1; \
	fi
	DB_CONNECTION_STRING="$(DB_CONNECTION_STRING)" NODE="$(NODE)" NETWORK="$(NETWORK)" SERVICE_IDS="$(SERVICE_IDS)" ./scripts/hydrate-services.sh

.PHONY: hydrate-applications
hydrate-applications: ## Hydrate applications table from blockchain data
	@echo "📱 Hydrating applications table..."
	@if [ -z "$(NODE)" ] || [ -z "$(NETWORK)" ] || [ -z "$(APPLICATION_ADDRESSES)" ]; then \
		echo "❌ Required parameters missing. Usage:"; \
		echo "   make hydrate-applications NODE=<node_url> NETWORK=<network_id> APPLICATION_ADDRESSES=<addresses>"; \
		echo "   Example: make hydrate-applications NODE=https://node.example.com NETWORK=pocket APPLICATION_ADDRESSES=addr1,addr2"; \
		exit 1; \
	fi
	@if ! docker ps | grep -q portal-db; then \
		echo "❌ Portal DB is not running. Please start it first:"; \
		echo "   make postgrest-up"; \
		exit 1; \
	fi
	DB_CONNECTION_STRING="$(DB_CONNECTION_STRING)" NODE="$(NODE)" NETWORK="$(NETWORK)" APPLICATION_ADDRESSES="$(APPLICATION_ADDRESSES)" ./scripts/hydrate-applications.sh

.PHONY: hydrate-gateways
hydrate-gateways: ## Hydrate gateways table from blockchain data
	@echo "🔗 Hydrating gateways table..."
	@if [ -z "$(NODE)" ] || [ -z "$(NETWORK)" ] || [ -z "$(GATEWAY_ADDRESSES)" ]; then \
		echo "❌ Required parameters missing. Usage:"; \
		echo "   make hydrate-gateways NODE=<node_url> NETWORK=<network_id> GATEWAY_ADDRESSES=<addresses>"; \
		echo "   Example: make hydrate-gateways NODE=https://node.example.com NETWORK=pocket GATEWAY_ADDRESSES=addr1,addr2"; \
		exit 1; \
	fi
	@if ! docker ps | grep -q portal-db; then \
		echo "❌ Portal DB is not running. Please start it first:"; \
		echo "   make postgrest-up"; \
		exit 1; \
	fi
	DB_CONNECTION_STRING="$(DB_CONNECTION_STRING)" NODE="$(NODE)" NETWORK="$(NETWORK)" GATEWAY_ADDRESSES="$(GATEWAY_ADDRESSES)" ./scripts/hydrate-gateways.sh

.PHONY: hydrate-prod
hydrate-prod: ## Hydrate local database with production data
	@echo "🏭 Hydrating local database with production data..."
	@echo "   Note: Requires .env file with production database credentials"
	@if ! docker ps | grep -q portal-db; then \
		echo "❌ Portal DB is not running. Please start it first:"; \
		echo "   make postgrest-up"; \
		exit 1; \
	fi
	./scripts/hydrate-prod.sh
