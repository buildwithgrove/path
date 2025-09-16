# Portal DB PostgREST API

This directory contains the PostgREST API setup for the Portal Database, providing a RESTful API automatically generated from your PostgreSQL schema.

## 🚀 Quick Start

```bash
# 1. Start the portal database (from parent directory)
cd .. && make portal_db_up

# 2. Set up API database roles and permissions (from api directory)
cd api && make setup-db

# 3. Start the API services (PostgreSQL + PostgREST)
make up

# 4. Generate OpenAPI spec and Go SDK
make generate-all

# 5. Test the API
curl http://localhost:3000/networks
```

## 📋 Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Getting Started](#getting-started)
- [API Endpoints](#api-endpoints)
- [Authentication](#authentication)
- [SDK Generation](#sdk-generation)
- [Development](#development)
- [Deployment](#deployment)
- [Troubleshooting](#troubleshooting)

## 🔍 Overview

This PostgREST API provides:

- **Automatic REST API** generation from your PostgreSQL schema
- **Type-safe Go SDK** with automatic code generation
- **OpenAPI 3.0 specification** for integration
- **Row-level security** for data access control
- **JWT authentication** for secure access
- **CORS support** for web applications

### Key Features

- ✅ **Zero-code API generation** - PostgREST introspects your database schema
- ✅ **Secure by default** - Row-level security policies protect sensitive data
- ✅ **Standards-compliant** - OpenAPI 3.0 specification (converted from Swagger 2.0)
- ✅ **Go SDK generation** - Type-safe client using oapi-codegen
- ✅ **Rich querying** - Filtering, sorting, pagination, and joins
- ✅ **Docker-based deployment** - Easy containerized setup

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐
│   Client Apps   │    │    Go SDK       │
│  (HTTP clients) │    │  (Type-safe)    │
└─────────────────┘    └─────────────────┘
         │                       │
         └───────────────────────┼────────────────────┐
                                 │                    │
              ┌──────────────────────────────────┐    │
              │          PostgREST API           │    │
              │        (Port 3000)               │    │
              │     OpenAPI 3.0 Spec             │    │
              └──────────────────────────────────┘    │
                                 │                    │
              ┌──────────────────────────────────┐    │
              │       PostgreSQL Database        │    │
              │     Portal DB (Port 5435)        │    │
              │    Row-Level Security            │    │
              └──────────────────────────────────┘    │
                                                      │
              ┌──────────────────────────────────┐    │
              │       oapi-codegen               │    │
              │    SDK Generation Tool           │◄───┘
              └──────────────────────────────────┘
```

### Components

- **PostgREST**: Auto-generates REST API from PostgreSQL schema
- **PostgreSQL**: Portal database with row-level security policies
- **oapi-codegen**: Generates type-safe Go SDK from OpenAPI specification
- **OpenAPI 3.0**: API specification (converted from PostgREST's Swagger 2.0)

## 🎯 Getting Started

### Prerequisites

- Docker and Docker Compose [[memory:8858267]]
- Make (for convenient commands)
- Node.js (for OpenAPI conversion): `brew install node`
- Go 1.22+ (for SDK generation): `brew install go`
- Your portal database running (see [parent README](../README.md))

### Installation

1. **Start the portal database** (if not already running):
   ```bash
   cd .. && make portal_db_up
   ```

2. **Set up API database roles**:
   ```bash
   make setup-db
   ```

3. **Start the API services**:
   ```bash
   make postgrest-up
   ```

4. **Verify the setup**:
   ```bash
   # Test API endpoint
   curl http://localhost:3000/networks
   ```

### Services URLs

- **PostgreSQL Database**: `localhost:5435`
- **PostgREST API**: http://localhost:3000
- **OpenAPI Specification**: http://localhost:3000/ (with `Accept: application/openapi+json`)

## 🌐 API Endpoints

### Public Endpoints (No Authentication Required)

| Method | Endpoint             | Description                        |
| ------ | -------------------- | ---------------------------------- |
| `GET`  | `/networks`          | List all supported networks        |
| `GET`  | `/services`          | List all available services        |
| `GET`  | `/portal_plans`      | List all portal subscription plans |
| `GET`  | `/service_endpoints` | List service endpoint types        |

### Protected Endpoints (Authentication Required)

| Method | Endpoint                          | Description                               |
| ------ | --------------------------------- | ----------------------------------------- |
| `GET`  | `/organizations`                  | List organizations (filtered by access)   |
| `GET`  | `/portal_accounts`                | List portal accounts (user's access only) |
| `GET`  | `/portal_applications`            | List applications (user's access only)    |
| `POST` | `/portal_applications`            | Create new application                    |
| `PUT`  | `/portal_applications?id=eq.{id}` | Update application                        |
| `GET`  | `/api/current_user_info`          | Get current user information              |
| `GET`  | `/api/user_accounts`              | Get user's accessible accounts            |
| `GET`  | `/api/user_applications`          | Get user's accessible applications        |

### Query Features

PostgREST provides powerful querying capabilities:

#### Basic Filtering
```bash
# Get only active services
curl "http://localhost:3000/services?active=eq.true"

# Get Ethereum mainnet service specifically
curl "http://localhost:3000/services?service_id=eq.ethereum-mainnet"

# Get services with compute units greater than 1
curl "http://localhost:3000/services?compute_units_per_relay=gt.1"

# Get services using pattern matching
curl "http://localhost:3000/services?service_name=ilike.*Ethereum*"
```

#### Field Selection & Sorting
```bash
# Select only specific fields
curl "http://localhost:3000/services?select=service_id,service_name,active"

# Sort by service name ascending
curl "http://localhost:3000/services?order=service_name.asc"

# Sort by multiple fields
curl "http://localhost:3000/services?order=active.desc,service_name.asc"
```

#### Pagination & Counting
```bash
# Paginate results (limit 2, skip first 2)
curl "http://localhost:3000/services?limit=2&offset=2"

# Get total count in response header
curl -I -H "Prefer: count=exact" "http://localhost:3000/services"

# Get count with results
curl -H "Prefer: count=exact" "http://localhost:3000/services"
```

#### Advanced Filtering
```bash
# Multiple conditions (AND)
curl "http://localhost:3000/services?active=eq.true&compute_units_per_relay=eq.1"

# OR conditions
curl "http://localhost:3000/services?or=(service_id.eq.ethereum-mainnet,service_id.eq.polygon-mainnet)"

# NULL checks
curl "http://localhost:3000/services?service_owner_address=is.null"

# Array operations
curl "http://localhost:3000/services?service_domains=cs.{eth-mainnet.gateway.pokt.network}"
```

#### Resource Embedding (JOINs)
```bash
# Get services with their endpoints
curl "http://localhost:3000/services?select=service_id,service_name,service_endpoints(endpoint_type)"

# Get services with fallback URLs
curl "http://localhost:3000/services?select=*,service_fallbacks(fallback_url)"

# Get portal plans with accounts count
curl "http://localhost:3000/portal_plans?select=*,portal_accounts(count)"
```

#### Aggregation
```bash
# Count services by status
curl "http://localhost:3000/services?select=active,count=exact&group_by=active"

# Get unique endpoint types
curl "http://localhost:3000/service_endpoints?select=endpoint_type&group_by=endpoint_type"
```

For complete query syntax, see [PostgREST Documentation](https://postgrest.org/en/stable/api.html).

## 📋 Practical Examples

### CRUD Operations

#### Creating Resources (POST)
```bash
# Create a new service fallback
curl -X POST http://localhost:3000/service_fallbacks \
  -H "Content-Type: application/json" \
  -d '{
    "service_id": "ethereum-mainnet",
    "fallback_url": "https://eth-mainnet.alchemy.com/v2/fallback"
  }'

# Create a new service endpoint
curl -X POST http://localhost:3000/service_endpoints \
  -H "Content-Type: application/json" \
  -d '{
    "service_id": "base-mainnet",
    "endpoint_type": "REST"
  }'

# Create with return preference (get created object back)
curl -X POST http://localhost:3000/service_fallbacks \
  -H "Content-Type: application/json" \
  -H "Prefer: return=representation" \
  -d '{
    "service_id": "polygon-mainnet",
    "fallback_url": "https://polygon-mainnet.nodereal.io/v1/fallback"
  }'
```

#### Reading Resources (GET)
```bash
# Basic reads
curl "http://localhost:3000/services"
curl "http://localhost:3000/networks"
curl "http://localhost:3000/portal_plans"

# Filtered reads with practical filters
curl "http://localhost:3000/services?active=eq.true&select=service_id,service_name"
curl "http://localhost:3000/portal_plans?plan_usage_limit=not.is.null"
curl "http://localhost:3000/service_endpoints?endpoint_type=eq.JSON-RPC"

# Complex queries
curl "http://localhost:3000/services?select=service_id,service_name,service_endpoints(endpoint_type),service_fallbacks(fallback_url)&active=eq.true"
```

#### Updating Resources (PATCH)
```bash
# Enable a service
curl -X PATCH "http://localhost:3000/services?service_id=eq.base-mainnet" \
  -H "Content-Type: application/json" \
  -d '{"active": true}'

# Update service with multiple fields
curl -X PATCH "http://localhost:3000/services?service_id=eq.arbitrum-one" \
  -H "Content-Type: application/json" \
  -d '{
    "quality_fallback_enabled": true,
    "hard_fallback_enabled": true
  }'

# Update with return preference
curl -X PATCH "http://localhost:3000/services?service_id=eq.polygon-mainnet" \
  -H "Content-Type: application/json" \
  -H "Prefer: return=representation" \
  -d '{"compute_units_per_relay": 2}'
```

#### Deleting Resources (DELETE)
```bash
# Delete specific fallback
curl -X DELETE "http://localhost:3000/service_fallbacks?service_id=eq.ethereum-mainnet&fallback_url=eq.https://eth-mainnet.alchemy.com/v2/fallback"

# Delete with conditions
curl -X DELETE "http://localhost:3000/service_endpoints?service_id=eq.base-mainnet&endpoint_type=eq.REST"
```

### Response Format Examples

#### JSON (Default)
```bash
curl "http://localhost:3000/services?limit=1"
# Returns: [{"service_id":"ethereum-mainnet","service_name":"Ethereum Mainnet",...}]
```

#### CSV Format
```bash
curl -H "Accept: text/csv" "http://localhost:3000/services?select=service_id,service_name,active"
# Returns: CSV formatted data
```

#### Single Object (Not Array)
```bash
curl -H "Accept: application/vnd.pgrst.object+json" \
     "http://localhost:3000/services?service_id=eq.ethereum-mainnet"
# Returns: {"service_id":"ethereum-mainnet",...} (object, not array)
```

### Business Logic Examples

#### Get Complete Service Information
```bash
# Get service with all related data
curl "http://localhost:3000/services?select=*,service_endpoints(*),service_fallbacks(*)&service_id=eq.ethereum-mainnet"
```

#### Portal Analytics Queries
```bash
# Count services by network
curl "http://localhost:3000/services?select=network_id,count=exact&group_by=network_id"

# Get plan distribution
curl "http://localhost:3000/portal_accounts?select=portal_plan_type,count=exact&group_by=portal_plan_type"

# List all endpoint types available
curl "http://localhost:3000/service_endpoints?select=endpoint_type&group_by=endpoint_type"
```

#### Health Check Queries
```bash
# Check API connectivity
curl -I "http://localhost:3000/networks"

# Verify data integrity
curl "http://localhost:3000/services?select=count=exact&active=eq.true"

# Get system overview
curl "http://localhost:3000/?select=*" | jq '.paths | keys'
```

### Error Handling Examples

#### Testing Error Responses
```bash
# Invalid table
curl "http://localhost:3000/nonexistent"
# Returns: 404 with error details

# Invalid column
curl "http://localhost:3000/services?invalid_column=eq.test"
# Returns: 400 with column error

# Invalid data type
curl -X POST "http://localhost:3000/services" \
  -H "Content-Type: application/json" \
  -d '{"service_id": 123}'
# Returns: 400 with type validation error

# Constraint violation
curl -X POST "http://localhost:3000/services" \
  -H "Content-Type: application/json" \
  -d '{"service_id": "test", "service_domains": []}'
# Returns: 400 with constraint error
```

## 🔐 Authentication

### JWT Authentication

The API supports JWT (JSON Web Token) authentication for secure access to protected endpoints.

#### Using a Token

Include the JWT token in the `Authorization` header:

```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     http://localhost:3000/portal_accounts
```

**Note**: JWT token generation and validation would need to be implemented separately or integrated with your existing authentication system.

### Row-Level Security (RLS)

The API implements PostgreSQL Row-Level Security to ensure users can only access their own data:

- **Users** can only view/edit their own profile
- **Portal Accounts** are filtered by user membership
- **Applications** are filtered by account access
- **Admin users** have elevated permissions

## 🛠️ SDK Generation

### Automatic Generation

Generate the OpenAPI specification and type-safe Go SDK:

```bash
# Generate both OpenAPI spec and Go SDK
make generate-all

# Or generate individually
make generate-openapi  # Generate OpenAPI specification
make generate-sdks     # Generate Go SDK
```

### Go SDK Usage

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    
    portaldb "github.com/grove/path/portal-db/sdk/go"
)

func main() {
    // Create client with typed responses
    client, err := portaldb.NewClientWithResponses("http://localhost:3000")
    if err != nil {
        panic(err)
    }
    
    ctx := context.Background()
    
    // Example 1: Get all active services
    resp, err := client.GetServicesWithResponse(ctx, &portaldb.GetServicesParams{
        Active: &[]string{"eq.true"}[0],
    })
    if err != nil {
        panic(err)
    }
    
    if resp.StatusCode() == 200 && resp.JSON200 != nil {
        services := *resp.JSON200
        fmt.Printf("Found %d active services:\n", len(services))
        for _, service := range services {
            fmt.Printf("- %s (%s)\n", service.ServiceName, service.ServiceId)
        }
    }
    
    // Example 2: Get specific service with endpoints
    serviceResp, err := client.GetServicesWithResponse(ctx, &portaldb.GetServicesParams{
        ServiceId: &[]string{"eq.ethereum-mainnet"}[0],
        Select:    &[]string{"*,service_endpoints(endpoint_type)"}[0],
    })
    if err != nil {
        panic(err)
    }
    
    if serviceResp.StatusCode() == 200 && serviceResp.JSON200 != nil {
        service := (*serviceResp.JSON200)[0]
        fmt.Printf("Service: %s supports endpoints: %v\n", 
            service.ServiceName, service.ServiceEndpoints)
    }
}

// Example with authentication
func authenticatedExample() {
    token := "your-jwt-token"
    
    client, err := portaldb.NewClientWithResponses("http://localhost:3000")
    if err != nil {
        panic(err)
    }
    
    ctx := context.Background()
    
    // Add JWT token to requests
    requestEditor := func(ctx context.Context, req *http.Request) error {
        req.Header.Set("Authorization", "Bearer "+token)
        return nil
    }
    
    // Make authenticated request
    resp, err := client.GetPortalAccountsWithResponse(ctx, 
        &portaldb.GetPortalAccountsParams{}, requestEditor)
    if err != nil {
        panic(err)
    }
    
    if resp.StatusCode() == 200 && resp.JSON200 != nil {
        accounts := *resp.JSON200
        fmt.Printf("User has access to %d accounts\n", len(accounts))
    }
}
```

**For complete Go SDK documentation, see [SDK README](../sdk/go/README.md)**

<!-- TODO_IMPLEMENT: Add TypeScript client generation -->
<!-- The PostgREST API would benefit from a TypeScript/JavaScript SDK for frontend -->
<!-- and Node.js applications. This should use the same OpenAPI spec to generate -->
<!-- a type-safe TypeScript client similar to the Go SDK. Consider using: -->
<!-- - @apidevtools/swagger-parser for OpenAPI parsing -->
<!-- - @openapitools/openapi-generator-cli for TS client generation -->
<!-- - or a custom generator that produces more idiomatic TypeScript -->
<!-- Target output: sdk/typescript/ directory with npm package -->
<!-- Priority: Medium - would improve frontend developer experience -->

## 🔧 Development

### Development Environment

Start a full development environment with all services:

```bash
make postgrest-up  # Starts PostgREST and PostgreSQL services
```

### Available Commands

| Command                | Description                         |
| ---------------------- | ----------------------------------- |
| `make postgrest-up`    | Start PostgREST and PostgreSQL      |
| `make postgrest-down`  | Stop PostgREST and PostgreSQL       |
| `make postgrest-logs`  | Show service logs                   |
| `make setup-db`        | Set up database roles and permissions |
| `make generate-openapi`| Generate OpenAPI specification      |
| `make generate-sdks`   | Generate Go SDK from OpenAPI spec   |
| `make generate-all`    | Generate both OpenAPI spec and SDKs |

### Database Schema Changes

When you modify the database schema:

1. **Update the schema** in `../schema/001_schema.sql`
2. **Restart the database**:
   ```bash
   cd .. && make portal_db_clean && make portal_db_up
   ```
3. **Reapply API setup**:
   ```bash
   make setup-db
   ```
4. **Regenerate SDKs**:
   ```bash
   make generate-all
   ```

## 🚀 Deployment

### Production Considerations

1. **Security**:
   - Change default JWT secrets
   - Use environment-specific configuration
   - Enable HTTPS/TLS
   - Configure proper CORS origins

2. **Performance**:
   - Tune PostgreSQL connection pool
   - Add API rate limiting
   - Configure caching headers
   - Monitor database performance

3. **Monitoring**:
   - Add health checks
   - Configure logging
   - Set up metrics collection
   - Monitor API response times

### Container Deployment

The API can be deployed using the provided Docker Compose setup:

```bash
# Production deployment
docker compose -f docker-compose.yml up -d
```

### Kubernetes Deployment

For Kubernetes deployment, create appropriate manifests based on the Docker Compose configuration. Consider using your existing Helm charts pattern.

## 🔍 Troubleshooting

### Common Issues

#### API Not Starting

```bash
# Check if portal database is running
docker ps | grep path-portal-db

# Check PostgREST logs
make api-logs

# Verify database connection
make test-api
```

#### Authentication Issues

```bash
# Verify auth service is running
curl http://localhost:3001/health

# Check JWT token validity
curl -X POST http://localhost:3001/auth/verify \
  -H "Content-Type: application/json" \
  -d '{"token":"YOUR_TOKEN"}'
```

#### Permission Errors

```bash
# Reapply database setup
make setup-db

# Check user permissions in database
psql postgresql://portal_user:portal_password@localhost:5435/portal_db \
  -c "SELECT * FROM portal_users WHERE portal_user_email = 'your-email@example.com';"
```

### Useful Commands

```bash
# View all available routes
curl http://localhost:3000/

# Check database connection
curl http://localhost:3000/networks

# Test authentication
curl -H "Authorization: Bearer YOUR_TOKEN" \
     http://localhost:3000/portal_accounts

# View OpenAPI specification
curl -H "Accept: application/openapi+json" \
     http://localhost:3000/
```

## 📚 Additional Resources

- [PostgREST Documentation](https://postgrest.org/en/stable/)
- [PostgreSQL Row Level Security](https://www.postgresql.org/docs/current/ddl-rowsecurity.html)
- [OpenAPI Specification](https://swagger.io/specification/)
- [JWT.io](https://jwt.io/) - JWT token debugging

<!-- TODO_IMPROVE: Add Swagger UI integration for better API exploration -->
<!-- TODO_IMPROVE: Add API versioning strategy documentation -->
<!-- TODO_DOCUMENT: Add troubleshooting guide for common PostgREST configuration issues -->
<!-- TODO_IMPLEMENT: Add automated API testing with generated SDKs -->
<!-- TODO_CONSIDERATION: Consider adding GraphQL endpoint alongside REST API -->

## 🤝 Contributing

When contributing to the API:

1. Update documentation for any new endpoints
2. Regenerate SDKs after schema changes
3. Test both authenticated and public endpoints
4. Update this README for any new features

## 📄 License

This API setup is part of the Grove PATH project. See the main project license for details.
