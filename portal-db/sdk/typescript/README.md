# TypeScript SDK for Portal DB

Auto-generated TypeScript client for the Portal Database API.

## Installation

```bash
npm install @grove/portal-db-sdk
```

> **TODO**: Publish this package to npm registry

## Quick Start

Built-in client methods (auto-generated):

```typescript
import { PortalApplicationsApi, Configuration } from "@grove/portal-db-sdk";

// Create client
const config = new Configuration({
  basePath: "http://localhost:3000"
});
const client = new PortalApplicationsApi(config);

// Use built-in methods - no manual paths needed!
const applications = await client.portalApplicationsGet();
```

React integration:

```typescript
import { PortalApplicationsApi, RpcCreatePortalApplicationApi, Configuration } from "@grove/portal-db-sdk";
import { useState, useEffect } from "react";

const config = new Configuration({ basePath: "http://localhost:3000" });
const portalAppsClient = new PortalApplicationsApi(config);
const createAppClient = new RpcCreatePortalApplicationApi(config);

function PortalApplicationsList() {
  const [applications, setApplications] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    portalAppsClient.portalApplicationsGet().then((apps) => {
      setApplications(apps);
      setLoading(false);
    });
  }, []);

  const createApp = async () => {
    await createAppClient.rpcCreatePortalApplicationPost({
      rpcCreatePortalApplicationPostRequest: {
        pPortalAccountId: "account-123",
        pPortalUserId: "user-456",
        pPortalApplicationName: "My App",
        pEmoji: "🚀"
      }
    });
    // Refresh list
    const apps = await portalAppsClient.portalApplicationsGet();
    setApplications(apps);
  };

  if (loading) return "Loading...";

  return (
    <div>
      <button onClick={createApp}>Create App</button>
      <ul>
        {applications.map(app => (
          <li key={app.portalApplicationId}>
            {app.emoji} {app.portalApplicationName}
          </li>
        ))}
      </ul>
    </div>
  );
}
```

## Authentication

Add JWT tokens to your requests:

```typescript
import { PortalApplicationsApi, Configuration } from "@grove/portal-db-sdk";

// With JWT auth
const config = new Configuration({
  basePath: "http://localhost:3000",
  accessToken: jwtToken
});

const client = new PortalApplicationsApi(config);
```
