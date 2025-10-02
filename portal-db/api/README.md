# Portal Database API <!-- omit in toc -->

**PostgREST configuration** and **SDK generation tools** for the Portal Database.

PostgREST automatically generates a REST API from the PostgreSQL database schema.

## 🚀 Quick Start <!-- omit in toc -->

```bash
# From portal-db directory
make portal-db-up                    # Start PostgreSQL + PostgREST (port 3000)
make postgrest-hydrate-testdata      # Add test data
curl http://localhost:3000/networks | jq  # Test API
```

## Table of Contents <!-- omit in toc -->

- [How it Works](#how-it-works)
- [⚙️ Configuration](#️-configuration)
- [Authentication (Optional)](#authentication-optional)
- [💾 Database Transactions](#-database-transactions)
- [🛠️ SDK Generation](#️-sdk-generation)
- [🔧 Development Commands](#-development-commands)
- [Query Examples](#query-examples)
- [📚 Resources](#-resources)

## How it Works

**PostgREST** introspects PostgreSQL schema and auto-generates REST endpoints:

```bash
Database Schema → PostgREST → OpenAPI Spec → Go/TypeScript SDKs
```

## ⚙️ Configuration

PostgREST configuration in `postgrest.conf`:

```ini
db-uri = "postgresql://authenticator:password@postgres:5432/portal_db"
db-schemas = "public,api"
server-port = 3000
```

Database roles defined in `../schema/002_postgrest_init.sql`:

- `anon` - Public data (networks, services)
- `authenticated` - User data (accounts, applications)

## Authentication (Optional)

JWT authentication available for protected endpoints:

```bash
make postgrest-gen-jwt      # Generate test token
make test-postgrest-auth    # Test authentication flow
```

See `scripts/gen-jwt.sh` for details.

## 💾 Database Transactions

PostgreSQL functions are auto-exposed as RPC endpoints:

```sql
CREATE FUNCTION public.create_portal_application(...) RETURNS JSON AS $$
  -- Multi-step transaction logic
$$ LANGUAGE plpgsql;
```

Usage: `curl -X POST http://localhost:3000/rpc/create_portal_application -d '{...}'`

Test: `make test-postgrest-portal-app-creation`

## 🛠️ SDK Generation

```bash
make postgrest-generate-openapi   # OpenAPI spec only
```

## 🔧 Development Commands

Run `make help` for a list of available commands.

## Query Examples

```bash
# Filtering
curl "http://localhost:3000/services?active=eq.true"

# Field selection
curl "http://localhost:3000/services?select=service_id,service_name"

# Sorting & pagination
curl "http://localhost:3000/services?order=service_name.asc&limit=10"

# Joins
curl "http://localhost:3000/services?select=*,service_endpoints(*)"
```

## 📚 Resources

- [PostgREST Documentation](https://postgrest.org/en/stable/)
- [OpenAPI Specification](https://swagger.io/specification/)
