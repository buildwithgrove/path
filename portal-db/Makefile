# Portal DB API Makefile
# ============================================================================
# This Makefile provides developer workflows for the Portal DB:
# - Service management (Docker Compose)
# - API generation (OpenAPI, SDKs)
# - Data hydration (test data, blockchain data)
# - Authentication testing (JWT flows)
#
# Quick start: make quickstart
# Full help:   make help
# ============================================================================

# Color definitions
CYAN := \033[0;36m
BOLD := \033[1m
RESET := \033[0m
YELLOW := \033[0;33m
GREEN := \033[0;32m

JWT_ROLE ?= portal_db_admin
JWT_EMAIL ?= john@doe.com
JWT_EXPIRES ?= 1h

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
	@echo "$(BOLD)=== 💧 Data Hydration ===$(RESET)"
	@grep -h -E '^(postgrest-hydrate-testdata|hydrate-services|hydrate-applications|hydrate-gateways|hydrate-prod|dehydrate-reset-db):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-30s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)=== 🔐 Authentication & Testing ===$(RESET)"
	@grep -h -E '^(test-postgrest-auth|test-postgrest-portal-app-creation|postgrest-gen-jwt|grove-download-jwks|check-jwks):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-30s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)=== 📝 API Generation ===$(RESET)"
	@grep -h -E '^(postgrest-generate-openapi|postgrest-swagger-ui|postgrest-generate-sdks|postgrest-generate-all):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-30s$(RESET) %s\n", $$1, $$2}'
	@echo ""

# ============================================================================
# QUICK START
# ============================================================================

.PHONY: quickstart
quickstart: ## Quick start guide for Portal DB (shows commands to run)
	@echo ""
	@echo "$(BOLD)$(CYAN)🗄️ Portal DB Quick Start$(RESET)"
	@echo "Keep this in one shell and run the following commands in another shell"
	@echo ""
	@echo "$(BOLD)Step 1: Start Portal DB services$(RESET)"
	@echo "$(CYAN)  make portal-db-up$(RESET)"
	@echo ""
	@echo "$(BOLD)Step 2: Hydrate test data$(RESET)"
	@echo "$(CYAN)  make postgrest-hydrate-testdata$(RESET)"
	@echo ""
	@echo "$(BOLD)Step 3: Generate JWT tokens$(RESET)"
	@echo "$(CYAN)  make postgrest-gen-jwt$(RESET)  # Defaults to admin role"
	@echo "  # Optionally, generate reader token: $(CYAN)JWT_ROLE=reader make postgrest-gen-jwt$(RESET)"
	@echo "  # With custom email: $(CYAN)JWT_EMAIL=user@example.com make postgrest-gen-jwt$(RESET)"
	@echo "  # With custom expiry: $(CYAN)JWT_EXPIRES=24h make postgrest-gen-jwt$(RESET)"
	@echo ""
	@echo "$(BOLD)Step 4: Export JWT token$(RESET)"
	@echo "$(YELLOW)  # Copy the export commands from the previous output$(RESET)"
	@echo "$(CYAN)  export JWT_TOKEN=\"...\"$(RESET)"
	@echo ""
	@echo "$(BOLD)Step 5: Test authenticated endpoints with JWT$(RESET)"
	@echo "$(CYAN)  curl http://localhost:3000/networks -H \"Authorization: Bearer \$$JWT_TOKEN\" | jq$(RESET)"
	@echo "$(CYAN)  curl http://localhost:3000/portal_accounts -H \"Authorization: Bearer \$$JWT_TOKEN\" | jq$(RESET)"
	@echo "$(CYAN)  curl http://localhost:3000/organizations -H \"Authorization: Bearer \$$JWT_TOKEN\" | jq$(RESET)"
	@echo "$(CYAN)  curl -X POST http://localhost:3000/portal_applications -H \"Authorization: Bearer \$$JWT_TOKEN\" \\\
	  -H 'Content-Type: application/json' -H 'Prefer: return=representation' \\\
	  -d '{\"portal_account_id\":\"10000000-0000-0000-0000-000000000004\",\"portal_application_name\":\"CLI Quickstart App\",\"secret_key_hash\":\"demo\",\"secret_key_required\":false}' | jq$(RESET)"
	@echo ""
	@echo "$(BOLD)Step 6: Connect to database with psql$(RESET)"
	@echo "$(CYAN)  psql 'postgresql://postgres:portal_password@localhost:5435/portal_db'$(RESET)"
	@echo ""
	@echo "$(BOLD)Step 7: Query portal_applications (inside psql)$(RESET)"
	@echo "$(CYAN)  SELECT * FROM portal_applications;$(RESET)"
	@echo "$(YELLOW)  # Exit psql with \\q$(RESET)"
	@echo ""
	@echo "$(BOLD)Step 8: Test portal application creation$(RESET)"
	@echo "$(CYAN)  make test-postgrest-portal-app-creation$(RESET)"
	@echo ""
	@echo "$(BOLD)Step 9: Query portal_applications again$(RESET)"
	@echo "$(CYAN)  psql 'postgresql://postgres:portal_password@localhost:5435/portal_db' -c \"SELECT * FROM portal_applications;\"$(RESET)"
	@echo ""
	@echo "$(BOLD)Step 10: Regenerate OpenAPI + Swagger UI$(RESET)"
	@echo "$(CYAN)  make postgrest-generate-openapi$(RESET)"
	@echo "$(CYAN)  make postgrest-swagger-ui$(RESET)"
	@echo "$(YELLOW)  # Open http://localhost:8080 in your browser$(RESET)"
	@echo ""
	@echo "$(BOLD)Step 11: Explore the Swagger UI$(RESET)"
	@echo "  - Browse available endpoints"
	@echo "  - Test authenticated endpoints using JWT token"
	@echo "  - View request/response schemas"
	@echo "  - Try the /rpc/create_portal_application endpoint"
	@echo ""
	@echo "$(GREEN)$(BOLD)💡 Tip: Run 'make help' to see all available commands$(RESET)"
	@echo ""

# ============================================================================
# SERVICE MANAGEMENT
# ============================================================================

.PHONY: portal-db-up
portal-db-up: check-jwks ## Start all Portal DB services
	@echo "🚀 Starting all Portal DB services..."
	docker compose up -d
	@echo "✅ Services started!"
	@echo "   PostgreSQL DB: $(CYAN)localhost:5435$(RESET)"
	@echo "   PostgREST API: $(CYAN)http://localhost:3000$(RESET)"
	@echo ""
	@echo "🔧 You can connect to the database using:"
	@echo "   $(CYAN)psql 'postgresql://postgres:portal_password@localhost:5435/portal_db'$(RESET)"
	@echo ""
	@echo "🔐 To test the API with authentication:"
	@echo "   1. Generate a JWT token:"
	@echo "      $(CYAN)make postgrest-gen-jwt$(RESET)"
	@echo "   2. Export the token (copy from output above):"
	@echo "      $(CYAN)export POSTGREST_JWT_TOKEN=\"eyJ...\"$(RESET)"
	@echo "   3. Make an authenticated request:"
	@echo "      $(CYAN)curl -H \"Authorization: Bearer \$$POSTGREST_JWT_TOKEN\" \"http://localhost:3000/networks\" | jq '.'$(RESET)"

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

.PHONY: postgrest-swagger-ui
postgrest-swagger-ui: postgrest-generate-openapi ## Start Swagger UI to view OpenAPI specification (http://localhost:8081)
	@echo "🌐 Starting Swagger UI on $(CYAN)http://localhost:8081$(RESET) ..."
	@echo "   Press Ctrl+C to stop"
	@docker run --rm -p 8081:8080 \
		-e SWAGGER_JSON=/openapi.json \
		-v $(PWD)/api/openapi/openapi.json:/openapi.json:ro \
		swaggerapi/swagger-ui

.PHONY: postgrest-generate-sdks
postgrest-generate-sdks: postgrest-generate-openapi ## Generate Go and TypeScript SDKs from OpenAPI specification
	@echo "🔧 Generating Go and TypeScript SDKs from OpenAPI specification..."
	cd api/codegen && ./generate-sdks.sh

# ============================================================================
# AUTHENTICATION & TESTING
# ============================================================================

# Internal helper targets (prefixed with _) for pre-condition checks
.PHONY: _require-postgrest-api
_require-postgrest-api:
	@if ! docker ps | grep -q portal-db-api; then \
		echo "❌ PostgREST API is not running. Please start it first:"; \
		echo "   make portal-db-up"; \
		exit 1; \
	fi

.PHONY: _require-portal-db
_require-portal-db:
	@if ! docker ps | grep -q portal-db; then \
		echo "❌ Portal DB is not running. Please start it first:"; \
		echo "   make portal-db-up"; \
		exit 1; \
	fi

.PHONY: test-postgrest-auth
test-postgrest-auth: _require-postgrest-api ## Test JWT authentication flow
	@echo "🔐 Testing JWT authentication..."
	cd api/scripts && ./postgrest-test-auth.sh

.PHONY: test-postgrest-portal-app-creation
test-postgrest-portal-app-creation: _require-postgrest-api ## Test portal application creation and retrieval via authenticated Postgres flow
	@echo "🧪 Testing portal application creation..."
	cd api/scripts && ./test-postgrest-portal-app-creation.sh

.PHONY: postgrest-gen-jwt
postgrest-gen-jwt: _require-postgrest-api ## Generate JWT token (use JWT_ROLE=admin JWT_EMAIL=user@example.com JWT_EXPIRES=24h to customize)
	@echo "🔑 Generating JWT token..."
	cd api/scripts && ./postgrest-gen-jwt.sh --role $(JWT_ROLE) --email $(JWT_EMAIL) --expires $(JWT_EXPIRES)

.PHONY: grove-download-jwks
grove-download-jwks: ## 🌿 GROVE EMPLOYEES ONLY - Download JWKS from 1Password to api/jwks.json
	@echo "🌿 $(BOLD)GROVE EMPLOYEES ONLY$(RESET)"
	@echo "🔐 Downloading JWKS from 1Password..."
	@if ! command -v op >/dev/null 2>&1; then \
		echo "❌ 1Password CLI (op) is not installed. Please install it first:"; \
		echo "   brew install 1password-cli"; \
		exit 1; \
	fi
	@op item get 6hh3pjg7m2qdpkg26kf6gzmjs4 --format json | jq -r '.fields[] | select(.id == "notesPlain") | .value' | jq '.' > api/jwks.json
	@echo "✅ JWKS downloaded to api/jwks.json"

.PHONY: check-jwks
check-jwks: ## Check if api/jwks.json exists, show instructions if missing
	@if [ ! -f api/jwks.json ]; then \
		echo "❌ $(YELLOW)api/jwks.json not found!$(RESET)"; \
		echo ""; \
		echo "$(BOLD)🌿 GROVE EMPLOYEES: Download the JWKS file from 1Password:$(RESET)"; \
		echo "   $(CYAN)make grove-download-jwks$(RESET)"; \
		echo ""; \
		echo "$(YELLOW)Note: Requires 1Password CLI (op) to be installed and authenticated.$(RESET)"; \
		exit 1; \
	fi
	@echo "✅ api/jwks.json found"

# ============================================================================
# DATA HYDRATION
# ============================================================================
# Database connection for all hydrate scripts (matches docker-compose.yml)
DB_CONNECTION_STRING := postgresql://postgres:portal_password@localhost:5435/portal_db

.PHONY: postgrest-hydrate-testdata
postgrest-hydrate-testdata: _require-portal-db ## Populate database with test data for development
	@echo "🧪 Hydrating Portal DB with test data..."
	DB_CONNECTION_STRING="$(DB_CONNECTION_STRING)" ./scripts/hydrate-testdata.sh

.PHONY: hydrate-services
hydrate-services: _require-portal-db ## Hydrate services table from blockchain data
	@echo "🌐 Hydrating services table..."
	@if [ -z "$(NODE)" ] || [ -z "$(NETWORK)" ] || [ -z "$(SERVICE_IDS)" ]; then \
		echo "❌ Required parameters missing. Usage:"; \
		echo "   make hydrate-services NODE=<node_url> NETWORK=<network_id> SERVICE_IDS=<service_ids>"; \
		echo "   Example: make hydrate-services NODE=https://node.example.com NETWORK=pocket SERVICE_IDS=ethereum-mainnet,polygon-mainnet"; \
		exit 1; \
	fi
	./scripts/hydrate-services.sh --services="$(SERVICE_IDS)" --node="$(NODE)" --chain-id="$(NETWORK)" --db-string="$(DB_CONNECTION_STRING)"

.PHONY: hydrate-applications
hydrate-applications: _require-portal-db ## Hydrate applications table from blockchain data
	@echo "📱 Hydrating applications table..."
	@if [ -z "$(NODE)" ] || [ -z "$(NETWORK)" ] || [ -z "$(APPLICATION_ADDRESSES)" ]; then \
		echo "❌ Required parameters missing. Usage:"; \
		echo "   make hydrate-applications NODE=<node_url> NETWORK=<network_id> APPLICATION_ADDRESSES=<addresses>"; \
		echo "   Example: make hydrate-applications NODE=https://node.example.com NETWORK=pocket APPLICATION_ADDRESSES=addr1,addr2"; \
		exit 1; \
	fi
	./scripts/hydrate-applications.sh --apps="$(APPLICATION_ADDRESSES)" --node="$(NODE)" --chain-id="$(NETWORK)" --db-string="$(DB_CONNECTION_STRING)"

.PHONY: hydrate-gateways
hydrate-gateways: _require-portal-db ## Hydrate gateways table from blockchain data
	@echo "🔗 Hydrating gateways table..."
	@if [ -z "$(NODE)" ] || [ -z "$(NETWORK)" ] || [ -z "$(GATEWAY_ADDRESSES)" ]; then \
		echo "❌ Required parameters missing. Usage:"; \
		echo "   make hydrate-gateways NODE=<node_url> NETWORK=<network_id> GATEWAY_ADDRESSES=<addresses>"; \
		echo "   Example: make hydrate-gateways NODE=https://node.example.com NETWORK=pocket GATEWAY_ADDRESSES=addr1,addr2"; \
		exit 1; \
	fi
	./scripts/hydrate-gateways.sh --gateways="$(GATEWAY_ADDRESSES)" --node="$(NODE)" --chain-id="$(NETWORK)" --db-string="$(DB_CONNECTION_STRING)"

.PHONY: hydrate-prod
hydrate-prod: _require-portal-db ## Hydrate local database with production data
	@echo "🏭 Hydrating local database with production data..."
	@echo "   Note: Requires .env file with production database credentials"
	./scripts/hydrate-prod.sh

.PHONY: dehydrate-reset-db
dehydrate-reset-db: ## Reset the development database (deletes all data and reinitializes)
	@echo "⚠️  WARNING: This will delete ALL database data and reinitialize!"
	@echo "🔄 Resetting database..."
	@echo "   Stopping services..."
	docker compose down
	@echo "   Removing ./tmp/portal_db_data..."
	rm -rf ./tmp/portal_db_data
	@echo "   Starting services (will reinitialize schema)..."
	docker compose up -d || docker compose up -d
	@echo "✅ Database reset complete and reinitialized with fresh schema."
