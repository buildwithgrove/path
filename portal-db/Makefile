# Portal DB API Makefile
# Provides convenient commands for managing the PostgREST API

.PHONY: postgrest-up postgrest-down postgrest-logs
.PHONY: setup-db generate-openapi generate-sdks test-api clean

# ============================================================================
# SERVICE MANAGEMENT
# ============================================================================

postgrest-up: ## Start PostgREST API service and PostgreSQL DB
	@echo "🚀 Starting Portal DB API services..."
	cd .. && docker compose up -d postgres postgrest
	@echo "✅ Services started!"
	@echo "   PostgreSQL DB: localhost:5435"
	@echo "   PostgREST API: http://localhost:3000"

postgrest-down: ## Stop PostgREST API service and PostgreSQL DB
	@echo "🛑 Stopping Portal DB API services..."
	cd .. && docker compose down
	@echo "✅ Services stopped!"

postgrest-logs: ## Show logs from PostgREST API service and PostgreSQL DB
	cd .. && docker compose logs -f

# ============================================================================
# DATABASE SETUP
# ============================================================================

setup-db: ## Set up database roles and permissions for API access
	@echo "🔧 Setting up database roles and permissions..."
	@if ! docker ps | grep -q path-portal-db; then \
		echo "❌ Portal DB is not running. Please start it first:"; \
		echo "   cd .. && make portal_db_up"; \
		exit 1; \
	fi
	@echo "   Applying API database setup..."
	docker exec -i path-portal-db psql -U portal_user -d portal_db < postgrest/schema.sql
	@echo "✅ Database setup completed!"
	@echo "   Created roles: web_anon, web_user, web_admin"
	@echo "   Applied row-level security policies"
	@echo "   Created API helper functions"

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
