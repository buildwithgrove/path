# Grove Portal DB TypeScript SDK

Type-safe TypeScript SDK for the Grove Portal DB API generated using [openapi-typescript](https://openapi-ts.dev/introduction).

## Installation

```bash
npm install @grove/portal-db-sdk
```

> **TODO**: Publish this package to npm registry

## Quick Start

React integration with [openapi-react-query](https://openapi-ts.dev/openapi-react-query/):

```bash
npm install openapi-react-query openapi-fetch @tanstack/react-query
```

```typescript
import createFetchClient from "openapi-fetch";
import createClient from "openapi-react-query";
import type { paths } from "@grove/portal-db-sdk";

// Create clients
const fetchClient = createFetchClient<paths>({
  baseUrl: "http://localhost:3000",
});
const $api = createClient(fetchClient);

// Use in React components
function PortalApplicationsList() {
  const { data: applications, error, isLoading } = $api.useQuery(
    "get", 
    "/portal_applications"
  );

  if (isLoading) return "Loading...";
  if (error) return `Error: ${error.message}`;

  return (
    <ul>
      {applications?.map(app => (
        <li key={app.portal_application_id}>
          {app.emoji} {app.portal_application_name}
        </li>
      ))}
    </ul>
  );
}

function CreateApplication() {
  const { mutate } = $api.useMutation("post", "/rpc/create_portal_application");
  
  return (
    <button onClick={() => mutate({
      body: { 
        p_portal_account_id: "account-123",
        p_portal_user_id: "user-456",
        p_portal_application_name: "My App",
        p_emoji: "🚀"
      }
    })}>
      Create Application
    </button>
  );
}
```

## Authentication

Add JWT tokens to your requests:

```typescript
import createFetchClient from "openapi-fetch";

// With JWT auth
const fetchClient = createFetchClient<paths>({
  baseUrl: "http://localhost:3000",
  headers: {
    Authorization: `Bearer ${jwtToken}`
  }
});

// Or set auth dynamically
fetchClient.use("auth", (request) => {
  const token = localStorage.getItem('jwt-token');
  if (token) {
    request.headers.set('Authorization', `Bearer ${token}`);
  }
});
```