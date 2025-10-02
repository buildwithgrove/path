# Portal Database API

<!-- TODO_DOCUMENTATION(@commoddity): Add section describing potential deployment to production using Pulumi, similar to how Porta database itself is deploted in the infra repo. -->

This folder contains **PostgREST configuration** and **SDK generation tools** for the Portal Database.

PostgREST automatically creates a REST API from the Portal DB PostgreSQL database schema.

## 🚀 Quick Start <!-- omit in toc -->

### 1. Start PostgREST <!-- omit in toc -->

```bash
# From portal-db directory
make postgrest-up
```

This starts:

- **PostgreSQL**: Database with Portal schema
- **PostgREST**: API server on http://localhost:3000

### 2. Hydrate the database with test data <!-- omit in toc -->

Hydrate the database with test data:

```bash
make hydrate-testdata
```

### 3. Test the API <!-- omit in toc -->

```bash
# View all networks (public data)
curl http://localhost:3000/networks

# View API documentation
curl http://localhost:3000/ | jq
```

# Table of Contents <!-- omit in toc -->

- [Portal Database API](#portal-database-api)
  - [🤔 What is This?](#-what-is-this)
  - [🏗️ How it Works](#️-how-it-works)
  - [📁 Folder Structure](#-folder-structure)
  - [⚙️ Configuration](#️-configuration)
    - [PostgREST Configuration (`postgrest.conf`)](#postgrest-configuration-postgrestconf)
    - [Database Roles (`../schema/002_postgrest_init.sql`)](#database-roles-schema002_postgrest_initsql)
  - [🔐 Authentication](#-authentication)
    - [How JWT Authentication Works](#how-jwt-authentication-works)
    - [Generate JWT Tokens](#generate-jwt-tokens)
    - [Use JWT Tokens](#use-jwt-tokens)
    - [Permission Levels](#permission-levels)
    - [Test Authentication](#test-authentication)
  - [💾 Database Transactions](#-database-transactions)
  - [🛠️ Go SDK Generation](#️-go-sdk-generation)
    - [Generate SDK](#generate-sdk)
    - [Generated Files](#generated-files)
  - [🔧 Development](#-development)
    - [Available Commands](#available-commands)
    - [After Database Schema Changes](#after-database-schema-changes)
    - [Query Features Examples](#query-features-examples)
  - [🚀 Next Steps](#-next-steps)
    - [For Beginners](#for-beginners)
  - [📚 Resources](#-resources)

## 🤔 What is This?

**PostgREST** is a tool that reads your PostgreSQL database and automatically generates a complete REST API. No code required - it introspects your tables, views, and functions to create endpoints.

**This folder provides:**

- ✅ **PostgREST Configuration**: Database connection, JWT auth, and API settings
- ✅ **JWT Authentication**: Role-based access control using database roles
- ✅ **Go SDK Generation**: Type-safe Go client from the OpenAPI specification
- ✅ **Testing Scripts**: JWT token generation and authentication testing

## 🏗️ How it Works

```
Database Schema  →  PostgREST  →  OpenAPI Spec  →  Go SDK
     │                 │                │            │
   Tables           Auto-gen         Endpoints     Type-safe
   Views            OpenAPI           + Auth        Client
   Functions        Spec              CRUD ops
```

1. **Database Schema**: Your PostgreSQL tables, views, and functions
2. **PostgREST**: Reads schema and creates REST endpoints automatically
3. **OpenAPI Spec**: PostgREST generates API documentation
4. **Go SDK**: Generated from OpenAPI spec for type-safe client code

## 📁 Folder Structure

```
api/
├── scripts/                # Helper scripts
│   ├── gen-jwt.sh          # Generate JWT tokens for testing
│   └── test-auth.sh        # Test authentication flow
├── codegen/                # SDK generation configuration
│   ├── codegen-*.yaml      # oapi-codegen config files
│   ├── generate-openapi.sh # OpenAPI specification generation scripts
│   └── generate-sdks.sh    # SDK generation scripts
├── openapi/                # Generated API documentation
│   └── openapi.json        # OpenAPI 3.0 specification
├── postgrest.conf           # Main PostgREST config file
└── README.md               # This file
```

## ⚙️ Configuration

### PostgREST Configuration (`postgrest.conf`)

Key settings for PostgREST:

```ini
# Database connection
db-uri = "postgresql://authenticator:password@postgres:5432/portal_db"
db-schemas = "public,api"        # Schemas to expose via API
db-anon-role = "anon"            # Default role for unauthenticated requests

# JWT Authentication
jwt-secret = "your-secret-key"   # Secret for verifying JWT tokens
jwt-role-claim-key = ".role"     # JWT claim containing database role

# Server settings
server-host = "0.0.0.0"
server-port = 3000
```

### Database Roles (`../schema/002_postgrest_init.sql`)

<!-- TODO_FUTURE(@commoddity): add more granular permissions -->

| Role            | Purpose                   | Permissions                           |
| --------------- | ------------------------- | ------------------------------------- |
| `authenticator` | PostgREST connection role | Can switch to other roles             |
| `anon`          | Anonymous users           | Public data only (networks, services) |
| `authenticated` | Logged-in users           | User data (accounts, applications)    |

## 🔐 Authentication

### How JWT Authentication Works

```
1. Generate JWT Token (external)
   ├── Role: "authenticated"
   ├── Email: "user@example.com"
   └── Secret: Shared with PostgREST

2. Client Request
   └── Header: Authorization: Bearer <JWT_TOKEN>

3. PostgREST Processing (happens automatically)
   ├── Verify JWT signature
   ├── Extract 'role' claim
   ├── Execute: SET ROLE <extracted_role>;
   └── Run query with role permissions

4. Database Query
   └── Permissions enforced by PostgreSQL roles
```

### Generate JWT Tokens

```bash
make gen-jwt
```

**Example Output:**

```
🔑 JWT Token Generated ✨
👤 Role: authenticated
📧 Email: john@doe.com
🎟️ Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Use JWT Tokens

```bash
# Set token as variable (from script output)
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Access protected endpoints
curl -H "Authorization: Bearer $TOKEN" \
     http://localhost:3000/portal_accounts

# Get current user info from JWT claims
curl -X POST -H "Authorization: Bearer $TOKEN" \
     http://localhost:3000/rpc/me
```

### Permission Levels

**Anonymous (`anon` role)**

- ✅ `networks` - Blockchain networks
- ✅ `services` - Available services
- ✅ `portal_plans` - Subscription plans
- ❌ User accounts or private data

**Authenticated (`authenticated` role)**

- ✅ All anonymous permissions
- ✅ `organizations` - Organization data
- ✅ `portal_accounts` - User accounts
- ✅ `portal_applications` - User applications

### Test Authentication

```bash
# Run complete authentication test suite
make test-auth
```

This tests:

- Anonymous access to public data
- JWT token generation
- Authenticated access to protected data
- JWT claims access via `/rpc/me`

## 💾 Database Transactions

For complex multi-step operations, create PostgreSQL functions that PostgREST automatically exposes as RPC endpoints:

```sql
-- Example: ../schema/003_postgrest_transactions.sql
CREATE OR REPLACE FUNCTION public.create_portal_application(
    p_portal_account_id VARCHAR(36),
    p_portal_user_id VARCHAR(36),
    p_portal_application_name VARCHAR(42) DEFAULT NULL
) RETURNS JSON AS $$
BEGIN
    -- Multi-step transaction logic here
    -- All operations are atomic within the function
END;
$$ LANGUAGE plpgsql VOLATILE SECURITY DEFINER;
```

**Usage:**

```bash
curl -X POST http://localhost:3000/rpc/create_portal_application \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"p_portal_account_id": "...", "p_portal_user_id": "..."}'
```

**Test transactions:**

```bash
make test-portal-app
```

## 🛠️ Go SDK Generation

### Generate SDK

When the PostgREST API is running on port `3000`, you can generate the Go SDK using the following command:

```bash
# Generate both OpenAPI spec and Go SDK
make generate-all

# Or generate individually
make generate-openapi  # OpenAPI specification only
```

### Generated Files

## 🔧 Development

### Available Commands

```bash
# Start PostgREST and PostgreSQL
make postgrest-up

# Stop services
make postgrest-down

# View logs
make postgrest-logs

# Generate SDK after schema changes
make generate-all

# Test authentication
make test-auth

# Populate with test data
make hydrate-testdata
```

### After Database Schema Changes

When you modify tables or add new functions:

1. **Update schema**: Edit `../schema/001_schema.sql`
2. **Restart database**: `make postgrest-down && make postgrest-up`
3. **Regenerate SDK**: `make generate-all`

### Query Features Examples

**Filtering:**

```bash
curl "http://localhost:3000/services?active=eq.true"
curl "http://localhost:3000/services?service_name=ilike.*Ethereum*"
```

**Field Selection:**

```bash
curl "http://localhost:3000/services?select=service_id,service_name"
```

**Sorting & Pagination:**

```bash
curl "http://localhost:3000/services?order=service_name.asc&limit=10&offset=20"
```

**Joins (Resource Embedding):**

```bash
curl "http://localhost:3000/services?select=*,service_endpoints(*)"
```

For complete query syntax, see [PostgREST API Documentation](https://postgrest.org/en/stable/api.html).

## 🚀 Next Steps

### For Beginners

1. **Explore the API**: Try the curl examples above
2. **Generate SDK**: Run `make generate-all`
3. **Read Go SDK docs**: Check `../sdk/go/README.md`
4. **Test authentication**: Run `make test-auth`
5. **Add test data**: Run `make hydrate-testdata`

## 📚 Resources

- [PostgREST Documentation](https://postgrest.org/en/stable/)
- [PostgreSQL Row Level Security](https://www.postgresql.org/docs/current/ddl-rowsecurity.html)
- [JWT.io](https://jwt.io/) - JWT token debugging
- [OpenAPI Specification](https://swagger.io/specification/)
