# Portal DB Go SDK

This Go SDK provides a type-safe client for the Portal DB API, generated using [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen).

## Installation

```bash
go get github.com/buildwithgrove/path/portal-db/sdk/go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/buildwithgrove/path/portal-db/sdk/go"
)

func main() {
    // Create a new client with typed responses
    client, err := portaldb.NewClientWithResponses("http://localhost:3000")
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Example: Get public data
    resp, err := client.GetNetworksWithResponse(ctx, &portaldb.GetNetworksParams{})
    if err != nil {
        log.Fatal(err)
    }
    
    if resp.StatusCode() == 200 && resp.JSON200 != nil {
        networks := *resp.JSON200
        fmt.Printf("Found %d networks\n", len(networks))
    }
}
```

## Authentication

For authenticated endpoints, add your JWT token to requests:

```go
import (
    "context"
    "net/http"
    
    "github.com/buildwithgrove/path/portal-db/sdk/go"
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
    
    // Make authenticated request
    resp, err := client.GetPortalAccountsWithResponse(
        ctx, 
        &portaldb.GetPortalAccountsParams{}, 
        requestEditor,
    )
    if err != nil {
        log.Fatal(err)
    }
    
    if resp.StatusCode() == 200 && resp.JSON200 != nil {
        accounts := *resp.JSON200
        fmt.Printf("Found %d accounts\n", len(accounts))
    } else {
        fmt.Printf("Authentication failed: %d\n", resp.StatusCode())
    }
}
```

## Query Features

The SDK supports PostgREST's powerful query features for filtering, selecting, and pagination:

### Filtering and Selection

```go
// Filter active services with specific fields
resp, err := client.GetServicesWithResponse(ctx, &portaldb.GetServicesParams{
    Active: func() *string { s := "eq.true"; return &s }(),
    Select: func() *string { s := "service_id,service_name,active,network_id"; return &s }(),
    Limit:  func() *string { s := "3"; return &s }(),
})
if err != nil {
    log.Fatal(err)
}

if resp.StatusCode() == 200 && resp.JSON200 != nil {
    services := *resp.JSON200
    fmt.Printf("Found %d active services\n", len(services))
}
```

### Specific Resource Lookup

```go
// Get a specific service by ID
resp, err := client.GetServicesWithResponse(ctx, &portaldb.GetServicesParams{
    ServiceId: func() *string { s := "eq.ethereum-mainnet"; return &s }(),
})
if err != nil {
    log.Fatal(err)
}

if resp.StatusCode() == 200 && resp.JSON200 != nil {
    services := *resp.JSON200
    fmt.Printf("Found service: %s\n", (*services)[0].ServiceName)
}
```

## RPC Functions

Access custom database functions via the RPC endpoint:

```go
// Get current user info from JWT claims
resp, err := client.PostRpcMeWithResponse(
    ctx, 
    &portaldb.PostRpcMeParams{}, 
    portaldb.PostRpcMeJSONRequestBody{}, 
    requestEditor,
)
if err != nil {
    log.Fatal(err)
}

if resp.StatusCode() == 200 {
    fmt.Printf("User info: %s\n", string(resp.Body))
}
```

## Error Handling

```go
resp, err := client.GetNetworksWithResponse(ctx, &portaldb.GetNetworksParams{})
if err != nil {
    // Handle network/client errors
    log.Printf("Client error: %v", err)
    return
}

switch resp.StatusCode() {
case 200:
    // Success - access typed data
    if resp.JSON200 != nil {
        networks := *resp.JSON200
        fmt.Printf("Found %d networks\n", len(networks))
    }
case 401:
    // Unauthorized
    fmt.Println("Authentication required")
default:
    // Other status codes
    fmt.Printf("Unexpected status: %d\n", resp.StatusCode())
}
```

## Development

This SDK was generated from the OpenAPI specification served by PostgREST. 

To regenerate run the following make target while the PostgREST API is running:

```bash
# From the portal-db directory
make generate-all
```

## Generated Files

- `models.go` - Generated data models and type definitions
- `client.go` - Generated API client methods and HTTP logic
- `go.mod` - Go module definition
- `README.md` - This documentation

## Related Documentation

- **API Documentation**: [../../api/README.md](../../api/README.md)
- **OpenAPI Specification**: `../../api/openapi/openapi.json`
