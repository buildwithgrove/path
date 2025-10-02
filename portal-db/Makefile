# Portal DB API Makefile

# Color definitions
CYAN := \033[0;36m
BOLD := \033[1m
RESET := \033[0m
YELLOW := \033[0;33m
GREEN := \033[0;32m

# ============================================================================
# HELP
# ============================================================================

.PHONY: help
.DEFAULT_GOAL := help
help: ## Show all available targets
	@echo ""
	@echo "$(BOLD)$(CYAN)🗄️ Portal DB Makefile Targets$(RESET)"
	@echo ""
	@echo "$(BOLD)=== 🚀 Quick Start ===$(RESET)"
	@grep -h -E '^(quickstart):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-30s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)=== 🛠️ Service Management ===$(RESET)"
	@grep -h -E '^(portal-db-up|portal-db-down|portal-db-logs):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-30s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)=== 📝 API Generation ===$(RESET)"
	@grep -h -E '^(postgrest-generate-openapi|postgrest-generate-sdks|postgrest-generate-all):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-30s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)=== 🔐 Authentication & Testing ===$(RESET)"
	@grep -h -E '^(test-auth|test-portal-app|gen-jwt|reset-dev-db):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-30s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)=== 💧 Data Hydration ===$(RESET)"
	@grep -h -E '^(hydrate-testdata|hydrate-services|hydrate-applications|hydrate-gateways|hydrate-prod):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-30s$(RESET) %s\n", $$1, $$2}'
	@echo ""

# ============================================================================
# QUICK START
# ============================================================================

.PHONY: quickstart
quickstart: ## Quick start guide for Portal DB (starts services, hydrates data, tests endpoints)
	@echo ""
	@echo "$(BOLD)$(CYAN)🗄️ Portal DB Quick Start$(RESET)"
	@echo ""
	@echo "$(BOLD)Step 1: Start Portal DB services...$(RESET)"
	@echo "$(CYAN)make portal-db-up$(RESET)"
	@echo ""
	@echo "$(BOLD)Step 2: Hydrate test data...$(RESET)"
	@echo "$(CYAN)make hydrate-testdata$(RESET)"
	@echo ""
	@echo "$(BOLD)Step 3: Test public endpoint (networks)...$(RESET)"
	@echo "$(CYAN)curl -s http://localhost:3000/networks | jq$(RESET)"
	@echo ""
	@echo "$(BOLD)Step 4: Generate JWT token...$(RESET)"
	@echo "$(CYAN)make gen-jwt$(RESET)"
	@echo ""
	@echo "$(BOLD)Step 5: Set your JWT token:$(RESET)"
	@echo "$(YELLOW)export TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYXV0aGVudGljYXRlZCIsImVtYWlsIjoiam9obkBkb2UuY29tIiwiZXhwIjoxNzU4MjEzNjM5fQ.i1_Mrj86xsdgsxDqLmJz8FDd9dd-sJhlS0vBQXGIHuU$(RESET)"
	@echo ""
	@echo "$(BOLD)Step 6: Test authenticated endpoints...$(RESET)"
	@echo "$(CYAN)curl -s http://localhost:3000/portal_accounts -H \"Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYXV0aGVudGljYXRlZCIsImVtYWlsIjoiam9obkBkb2UuY29tIiwiZXhwIjoxNzU4MjEzNjM5fQ.i1_Mrj86xsdgsxDqLmJz8FDd9dd-sJhlS0vBQXGIHuU\" | jq$(RESET)"
	@echo ""
	@echo "$(BOLD)Step 7: Test rpc/me...$(RESET)"
	@echo "$(CYAN)curl -s http://localhost:3000/rpc/me -H \"Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYXV0aGVudGljYXRlZCIsImVtYWlsIjoiam9obkBkb2UuY29tIiwiZXhwIjoxNzU4MjEzNjM5fQ.i1_Mrj86xsdgsxDqLmJz8FDd9dd-sJhlS0vBQXGIHuU\" -H \"Content-Type: application/json\" | jq$(RESET)"
	@echo ""
	@echo "$(BOLD)Step 8: Test portal app creation...$(RESET)"
	@echo "$(CYAN)make test-portal-app$(RESET)"
	@echo ""
	@echo "$(GREEN)$(BOLD)✅ Quick start complete!$(RESET)"
	@echo ""

# ============================================================================
# SERVICE MANAGEMENT
# ============================================================================

.PHONY: portal-db-up
portal-db-up: ## Start all Portal DB services
	@echo "🚀 Starting all Portal DB services..."
	docker compose up -d
	@echo "✅ Services started!"
	@echo "   PostgreSQL DB: $(CYAN)localhost:5435$(RESET)"
	@echo "   PostgREST API: $(CYAN)http://localhost:3000$(RESET)"
	@echo ""
	@echo "🔧 You can connect to the database using:"
	@echo "   $(CYAN)psql 'postgresql://postgres:portal_password@localhost:5435/portal_db'$(RESET)"

.PHONY: portal-db-down
portal-db-down: ## Stop all Portal DB services
	@echo "🛑 Stopping all Portal DB services..."
	docker compose down
	@echo "✅ Services stopped!"

.PHONY: portal-db-logs
portal-db-logs: ## Show logs from all Portal DB services
	docker compose logs -f

# ============================================================================
# API GENERATION
# ============================================================================

.PHONY: postgrest-generate-openapi
postgrest-generate-openapi: ## Generate OpenAPI specification from PostgREST
	@echo "📝 Generating OpenAPI specification..."
	cd api/codegen && ./generate-openapi.sh

.PHONY: postgrest-generate-sdks
postgrest-generate-sdks: ## Generate Go SDK from OpenAPI spec
	@echo "🔧 Generating Go SDK..."
	@echo "   Note: Requires Node.js and oapi-codegen"
	cd api/codegen && ./generate-sdks.sh

.PHONY: postgrest-generate-all
postgrest-generate-all: postgrest-generate-openapi postgrest-generate-sdks ## Generate OpenAPI spec and Go SDK
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
