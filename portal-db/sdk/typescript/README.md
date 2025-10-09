# Portal DB TypeScript SDK

Type-safe TypeScript client for the Portal DB API, generated from OpenAPI specification using [openapi-typescript](https://github.com/openapi-ts/openapi-typescript) and [openapi-fetch](https://github.com/openapi-ts/openapi-typescript/tree/main/packages/openapi-fetch).

## Installation

```bash
npm install openapi-fetch
```

## Usage

### Basic GET Request

```typescript
import createClient from 'openapi-fetch';
import type { paths } from './types';

const client = createClient<paths>({
  baseUrl: 'http://localhost:3000',
  headers: {
    'Authorization': `Bearer ${JWT_TOKEN}`,
  },
});

// Fetch all services
const { data, error } = await client.GET('/services');

if (error) {
  console.error('Error:', error);
} else {
  console.log('Services:', data);
}
```

### Query with Filters

```typescript
// GET with PostgREST filters
const { data, error } = await client.GET('/services', {
  params: {
    query: {
      active: 'eq.true',              // Filter: active = true
      service_id: 'like.*ethereum*',  // Pattern match
    }
  }
});
```

## Documentation

- **openapi-fetch**: https://openapi-ts.dev/openapi-fetch/
- **PostgREST API Reference**: https://postgrest.org/en/stable/references/api/tables_views.html

## Type Safety

All endpoints, parameters, and responses are fully typed based on the OpenAPI specification. TypeScript will provide autocomplete and catch errors at compile time.
