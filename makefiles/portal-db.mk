################################
###  Portal DB make targets  ###
################################

# These targets manage the local PostgreSQL database for portal development.
# The database runs in a Docker container and is accessible on port 5435.


.PHONY: portal_db_quickstart
portal_db_quickstart: ## Quick start guide for Portal DB (starts services, hydrates data, tests endpoints)
	@echo ""
	@echo "$(BOLD)$(CYAN)🗄️ Portal DB Quick Start$(RESET)"
	@echo ""
	@echo "$(BOLD)Step 1: Starting Portal DB services...$(RESET)"
	@cd ./portal-db && make postgrest-up
	@echo ""
	@echo "$(BOLD)Step 2: Hydrating test data...$(RESET)"
	@cd ./portal-db && make hydrate-testdata
	@echo ""
	@echo "$(BOLD)Step 3: Testing public endpoint (networks)...$(RESET)"
	@curl -s http://localhost:3000/networks | jq
	@echo ""
	@echo "$(BOLD)Step 4: Generating JWT token...$(RESET)"
	@cd ./portal-db && make gen-jwt
	@echo ""
	@echo "$(BOLD)Step 5: Set your JWT token:$(RESET)"
	@echo "$(YELLOW)export TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYXV0aGVudGljYXRlZCIsImVtYWlsIjoiam9obkBkb2UuY29tIiwiZXhwIjoxNzU4MjEzNjM5fQ.i1_Mrj86xsdgsxDqLmJz8FDd9dd-sJhlS0vBQXGIHuU$(RESET)"
	@echo ""
	@echo "$(BOLD)Step 6: Testing authenticated endpoints...$(RESET)"
	@echo "$(CYAN)Testing portal_accounts:$(RESET)"
	@curl -s http://localhost:3000/portal_accounts -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYXV0aGVudGljYXRlZCIsImVtYWlsIjoiam9obkBkb2UuY29tIiwiZXhwIjoxNzU4MjEzNjM5fQ.i1_Mrj86xsdgsxDqLmJz8FDd9dd-sJhlS0vBQXGIHuU" | jq
	@echo ""
	@echo "$(CYAN)Testing rpc/me:$(RESET)"
	@curl -s http://localhost:3000/rpc/me -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYXV0aGVudGljYXRlZCIsImVtYWlsIjoiam9obkBkb2UuY29tIiwiZXhwIjoxNzU4MjEzNjM5fQ.i1_Mrj86xsdgsxDqLmJz8FDd9dd-sJhlS0vBQXGIHuU" -H "Content-Type: application/json" | jq
	@echo ""
	@echo "$(BOLD)Step 7: Testing portal app creation...$(RESET)"
	@cd ./portal-db && make test-portal-app
	@echo ""
	@echo "$(GREEN)$(BOLD)✅ Quick start complete!$(RESET)"
	@echo ""

.PHONY: portal_db_up
portal_db_up: check_docker ## Creates the Portal Postgres DB with the base schema (`./schema/001_portal_init.sql`) and runs the Portal DB on port `:5435`.
	@echo "🚀 Starting portal-db PostgreSQL container..."
	@cd portal-db && docker compose up -d
	@echo "✅ PostgreSQL is starting up on port 5435"
	@echo "   Database: portal_db"
	@echo "   User: portal_user"
	@echo "   Password: portal_password"
	@echo "   Connection: postgresql://portal_user:portal_password@localhost:5435/portal_db"
	@echo ""
	@echo "🔧 To use the hydrate scripts, export the DB connection string:"
	@echo "   export DB_CONNECTION_STRING='postgresql://portal_user:portal_password@localhost:5435/portal_db'"