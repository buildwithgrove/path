################################
###  Portal DB make targets  ###
################################

# These targets manage the local PostgreSQL database for portal development.
# The database runs in a Docker container and is accessible on port 5435.

.PHONY: portal_db_up
portal_db_up: check_docker ## Start local PostgreSQL database on port 5435
	@echo "🚀 Starting portal-db PostgreSQL container..."
	@cd portal-db && docker-compose up -d
	@echo "✅ PostgreSQL is starting up on port 5435"
	@echo "   Database: portal_db"
	@echo "   User: portal_user"
	@echo "   Password: portal_password"
	@echo "   Connection: postgresql://portal_user:portal_password@localhost:5435/portal_db"
	@echo ""
	@echo "🔧 To use the hydrate scripts, export the DB connection string:"
	@echo "   export DB_CONNECTION_STRING='postgresql://portal_user:portal_password@localhost:5435/portal_db'"

.PHONY: portal_db_env
portal_db_env: portal_db_up ## Start database and wait for readiness, then export connection string
	@echo "⏳ Waiting for PostgreSQL to be ready..."
	@timeout=60; \
	while [ $$timeout -gt 0 ]; do \
		if docker exec path-portal-db pg_isready -U portal_user -d portal_db >/dev/null 2>&1; then \
			echo "✅ PostgreSQL is ready!"; \
			break; \
		fi; \
		echo "⌛ PostgreSQL not ready yet, waiting... ($$timeout seconds left)"; \
		sleep 2; \
		timeout=$$((timeout-2)); \
	done; \
	if [ $$timeout -eq 0 ]; then \
		echo "❌ PostgreSQL failed to start within 60 seconds"; \
		exit 1; \
	fi
	@echo ""
	@echo "🎉 Portal DB is ready! Export the connection string with:"
	@echo "   export DB_CONNECTION_STRING='postgresql://portal_user:portal_password@localhost:5435/portal_db'"
	@echo ""
	@echo "📝 Or copy and paste this command:"
	@echo 'export DB_CONNECTION_STRING="postgresql://portal_user:portal_password@localhost:5435/portal_db"'

.PHONY: portal_db_down
portal_db_down: ## Stop local PostgreSQL database
	@echo "🛑 Stopping portal-db PostgreSQL container..."
	@cd portal-db && docker-compose down
	@echo "✅ PostgreSQL container stopped"

.PHONY: portal_db_status
portal_db_status: ## Check status of portal-db PostgreSQL container
	@echo "📊 Portal DB Status:"
	@cd portal-db && docker-compose ps

.PHONY: portal_db_logs
portal_db_logs: ## Show logs from portal-db PostgreSQL container
	@cd portal-db && docker-compose logs -f postgres

.PHONY: portal_db_clean
portal_db_clean: portal_db_down ## Stop database and remove all data volumes
	@echo "🧹 Cleaning up portal-db data volumes..."
	@cd portal-db && docker-compose down -v
	@docker volume prune -f --filter label=com.docker.compose.project=portal-db
	@echo "✅ All portal-db data removed"

.PHONY: portal_db_connect
portal_db_connect: ## Connect to the portal database using psql
	@echo "🔗 Connecting to portal database..."
	@docker exec -it path-portal-db psql -U portal_user -d portal_db