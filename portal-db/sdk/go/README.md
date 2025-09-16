# Portal DB Go SDK

This Go SDK provides a type-safe client for the Portal DB API, generated using [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen).

## Installation

```bash
go get github.com/grove/path/portal-db/sdk/go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/grove/path/portal-db/sdk/go"
)

func main() {
    // Create a new client with typed responses
    client, err := portaldb.NewClientWithResponses("http://localhost:3000")
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Example: Get all services with typed response
    resp, err := client.GetServicesWithResponse(ctx, &portaldb.GetServicesParams{})
    if err != nil {
        log.Fatal(err)
    }
    
    // Check response status and access typed data
    switch resp.StatusCode() {
    case 200:
        if resp.JSON200 != nil {
            services := *resp.JSON200 // []portaldb.Services
            fmt.Printf("Found %d services\n", len(services))
            
            for _, service := range services {
                fmt.Printf("- Service: %s (ID: %s)\n", service.ServiceName, service.ServiceId)
                if service.Active != nil && *service.Active {
                    fmt.Printf("  Status: Active\n")
                }
                if service.ComputeUnitsPerRelay != nil {
                    fmt.Printf("  Compute Units: %d\n", *service.ComputeUnitsPerRelay)
                }
                fmt.Printf("  Domains: %v\n", service.ServiceDomains)
            }
        }
    case 401:
        fmt.Println("Authentication required")
    default:
        fmt.Printf("Unexpected status: %d\n", resp.StatusCode())
    }
}
```

## Authentication

For authenticated endpoints, add your JWT token to requests:

```go
import (
    "context"
    "net/http"
    
    "github.com/grove/path/portal-db/sdk/go"
)

func authenticatedExample() {
    client, err := portaldb.NewClientWithResponses("http://localhost:3000")
    if err != nil {
        log.Fatal(err)
    }
    
    // Add JWT token to requests
    token := "your-jwt-token-here"
    ctx := context.Background()
    
    // Use RequestEditorFn to add authentication header
    requestEditor := func(ctx context.Context, req *http.Request) error {
        req.Header.Set("Authorization", "Bearer "+token)
        return nil
    }
    
    // Make authenticated request with typed response
    resp, err := client.GetServicesWithResponse(ctx, &portaldb.GetServicesParams{}, requestEditor)
    if err != nil {
        log.Fatal(err)
    }
    
    if resp.StatusCode() == 200 && resp.JSON200 != nil {
        fmt.Printf("Authenticated: Found %d services\n", len(*resp.JSON200))
    } else {
        fmt.Printf("Authentication failed: %d\n", resp.StatusCode())
    }
}
```

## Available Endpoints

Based on your PostgREST schema, the SDK includes methods for:

- **Services** (`/services`) - Blockchain services from Pocket Network
  - `GetServicesWithResponse(ctx, params)` - List all services → `*[]Services`
  - `PostServicesWithResponse(ctx, params, body)` - Create a new service
  - `PatchServicesWithResponse(ctx, params, body)` - Update services
  - `DeleteServicesWithResponse(ctx, params)` - Delete services

- **Networks** (`/networks`) - Supported blockchain networks  
  - `GetNetworksWithResponse(ctx, params)` - List all networks → `*[]Networks`
  - `PostNetworksWithResponse(ctx, params, body)` - Create a new network
  - `PatchNetworksWithResponse(ctx, params, body)` - Update networks
  - `DeleteNetworksWithResponse(ctx, params)` - Delete networks

- **Portal Plans** (`/portal_plans`) - Subscription plans
  - `GetPortalPlansWithResponse(ctx, params)` - List all plans → `*[]PortalPlans`
  - `PostPortalPlansWithResponse(ctx, params, body)` - Create a new plan
  - `PatchPortalPlansWithResponse(ctx, params, body)` - Update plans
  - `DeletePortalPlansWithResponse(ctx, params)` - Delete plans

- **Service Endpoints** (`/service_endpoints`) - Endpoint types for services
  - `GetServiceEndpointsWithResponse(ctx, params)` - List all endpoints → `*[]ServiceEndpoints`
  - `PostServiceEndpointsWithResponse(ctx, params, body)` - Create a new endpoint
  - `PatchServiceEndpointsWithResponse(ctx, params, body)` - Update endpoints
  - `DeleteServiceEndpointsWithResponse(ctx, params)` - Delete endpoints

- **Service Fallbacks** (`/service_fallbacks`) - Fallback URLs for services
  - `GetServiceFallbacksWithResponse(ctx, params)` - List all fallbacks → `*[]ServiceFallbacks`
  - `PostServiceFallbacksWithResponse(ctx, params, body)` - Create a new fallback
  - `PatchServiceFallbacksWithResponse(ctx, params, body)` - Update fallbacks
  - `DeleteServiceFallbacksWithResponse(ctx, params)` - Delete fallbacks

## Error Handling

```go
resp, err := client.GetServicesWithResponse(ctx, &portaldb.GetServicesParams{})
if err != nil {
    // Handle network/client errors
    log.Printf("Client error: %v", err)
    return
}

switch resp.StatusCode() {
case 200:
    // Success - access typed data
    if resp.JSON200 != nil {
        services := *resp.JSON200 // []portaldb.Services
        fmt.Printf("Found %d services\n", len(services))
        for _, service := range services {
            fmt.Printf("Service: %s\n", service.ServiceName)
        }
    }
case 404:
    // Not found
    fmt.Println("Resource not found")
case 401:
    // Unauthorized
    fmt.Println("Authentication required")
default:
    // Other status codes
    fmt.Printf("Unexpected status: %d\n", resp.StatusCode())
}
```

## Configuration

You can customize the client behavior:

```go
import (
    "net/http"
    "time"
)

// Custom HTTP client with timeout
httpClient := &http.Client{
    Timeout: 30 * time.Second,
}

client, err := portaldb.NewClientWithResponses(
    "http://localhost:3000",
    portaldb.WithHTTPClient(httpClient),
)
```

## Development

This SDK was generated from the OpenAPI specification served by PostgREST. To regenerate:

```bash
# From the portal-db/api directory
make generate-sdks

# Or directly run the script
cd api/codegen
./generate-sdks.sh
```

## Generated Files

- `models.go` - Generated data models and type definitions (44KB)
- `client.go` - Generated API client methods and HTTP logic (189KB)
- `go.mod` - Go module definition
- `README.md` - This documentation

### File Organization

For better readability, the SDK is split into two main files:
- **`models.go`** - Contains all data structures, constants, and type definitions
- **`client.go`** - Contains the HTTP client, request builders, and API methods

This separation makes it easier to:
- Browse and understand the data models separately from client logic
- Navigate large codebases more efficiently
- Maintain and review changes to specific parts of the SDK

## Support

For issues with the generated SDK, please check:
1. [API README](../../api/README.md) - PostgREST API documentation
2. [oapi-codegen documentation](https://github.com/oapi-codegen/oapi-codegen) - SDK generation tool
3. Your database schema and PostgREST configuration

## Related Documentation

- **API Documentation**: [../../api/README.md](../../api/README.md)
- **OpenAPI Specification**: `../../api/openapi/openapi.json`
- **Database Schema**: `../../schema/001_schema.sql`
