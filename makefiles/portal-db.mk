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