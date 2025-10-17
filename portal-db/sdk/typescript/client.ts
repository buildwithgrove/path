/**
 * Grove Portal DB API Client
 * 
 * This client uses openapi-fetch for type-safe API requests.
 * It's lightweight with zero dependencies beyond native fetch.
 * 
 * @example
 * ```typescript
 * import createClient from './client';
 * 
 * const client = createClient({ baseUrl: 'http://localhost:3000' });
 * 
 * // GET request with full type safety
 * const { data, error } = await client.GET('/portal_accounts');
 * 
 * // POST request with typed body
 * const { data, error } = await client.POST('/portal_accounts', {
 *   body: { 
 *     portal_plan_type: 'PLAN_FREE',
 *     // ... other fields
 *   }
 * });
 * ```
 */
import createClient from 'openapi-fetch';
import type { paths } from './types';

export type { paths } from './types';

/**
 * Create a new API client instance
 * 
 * @param options - Client configuration options
 * @param options.baseUrl - Base URL for the API (default: http://localhost:3000)
 * @param options.headers - Default headers to include with every request
 * @returns Type-safe API client
 */
export default function createPortalDBClient(options?: {
  baseUrl?: string;
  headers?: HeadersInit;
}) {
  return createClient<paths>({
    baseUrl: options?.baseUrl || 'http://localhost:3000',
    headers: options?.headers,
  });
}

// Re-export for convenience
export { createClient };
