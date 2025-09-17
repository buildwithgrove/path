# Portal DB API Makefile
# Provides convenient commands for managing the PostgREST API

.PHONY: postgrest-up postgrest-down postgrest-logs
.PHONY: generate-openapi generate-sdks generate-all test-auth
.PHONY: hydrate-testdata hydrate-services hydrate-applications hydrate-gateways

# ============================================================================
# SERVICE MANAGEMENT
# ============================================================================

postgrest-up: ## Start PostgREST API service and PostgreSQL DB
	@echo "🚀 Starting Portal DB API services..."
	docker compose up -d
	@echo "✅ Services started!"
	@echo "   PostgreSQL DB: localhost:5435"
	@echo "   PostgREST API: http://localhost:3000"

postgrest-down: ## Stop PostgREST API service and PostgreSQL DB
	@echo "🛑 Stopping Portal DB API services..."
	docker compose down
	@echo "✅ Services stopped!"

postgrest-logs: ## Show logs from PostgREST API service and PostgreSQL DB
	docker compose logs -f

portal-db-up: ## Start Portal DB PostgreSQL container only
	@echo "🚀 Starting Portal DB PostgreSQL container..."
	docker compose up -d portal-db
	@echo "✅ PostgreSQL container started!"
	@echo "   Database: localhost:5435"
	@echo "   User: postgres"
	@echo "   Password: portal_password"

portal-db-down: ## Stop Portal DB PostgreSQL container only
	@echo "🛑 Stopping Portal DB PostgreSQL container..."
	docker compose down portal-db
	@echo "✅ PostgreSQL container stopped!"

# ============================================================================
# API GENERATION
# ============================================================================

generate-openapi: ## Generate OpenAPI specification from PostgREST
	@echo "📝 Generating OpenAPI specification..."
	cd api/codegen && ./generate-openapi.sh

generate-sdks: ## Generate Go SDK from OpenAPI spec
	@echo "🔧 Generating Go SDK..."
	@echo "   Note: Requires Node.js and oapi-codegen"
	cd api/codegen && ./generate-sdks.sh

generate-all: generate-openapi generate-sdks ## Generate OpenAPI spec and Go SDK
	@echo "✨ All generation tasks completed!"

# ============================================================================
# AUTHENTICATION & TESTING
# ============================================================================

test-auth: ## Test JWT authentication flow
	@echo "🔐 Testing JWT authentication..."
	@if ! docker ps | grep -q portal-db-api; then \
		echo "❌ PostgREST API is not running. Please start it first:"; \
		echo "   make postgrest-up"; \
		exit 1; \
	fi
	cd api/scripts && ./test-auth.sh

gen-jwt: ## Generate JWT token
	@echo "🔑 Generating JWT token..."
	@if ! docker ps | grep -q portal-db-api; then \
		echo "❌ PostgREST API is not running. Please start it first:"; \
		echo "   make postgrest-up"; \
		exit 1; \
	fi
	cd api/scripts && ./gen-jwt.sh

# ============================================================================
# DATA HYDRATION
# ============================================================================
# Database connection for all hydrate scripts (matches docker-compose.yml)
DB_CONNECTION_STRING := postgresql://postgres:portal_password@localhost:5435/portal_db

hydrate-testdata: ## Populate database with test data for development
	@echo "🧪 Hydrating Portal DB with test data..."
	@if ! docker ps | grep -q portal-db; then \
		echo "❌ Portal DB is not running. Please start it first:"; \
		echo "   make postgrest-up"; \
		exit 1; \
	fi
	DB_CONNECTION_STRING="$(DB_CONNECTION_STRING)" ./scripts/hydrate-testdata.sh

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
