/**
 * Grove Portal DB TypeScript SDK
 * 
 * Generated types from OpenAPI specification using openapi-typescript.
 * For a complete fetch client, consider using openapi-fetch alongside these types.
 * 
 * @see https://openapi-ts.dev/introduction
 */

// Export all generated types
export type * from './types';

// Re-export commonly used types for convenience
export type { paths, components, operations } from './types';

/**
 * Basic fetch wrapper with type safety
 * 
 * For production use, consider openapi-fetch:
 * npm install openapi-fetch
 * 
 * Example with openapi-fetch:
 * ```typescript
 * import createClient from 'openapi-fetch';
 * import type { paths } from './types';
 * 
 * const client = createClient<paths>({ baseUrl: 'http://localhost:3000' });
 * const { data, error } = await client.GET('/users');
 * ```
 */
export async function createTypedFetch<T extends keyof paths>(
  baseUrl: string,
  options?: {
    headers?: Record<string, string>;
    timeout?: number;
  }
) {
  const defaultHeaders = {
    'Content-Type': 'application/json',
    ...options?.headers,
  };

  return {
    async request<
      Path extends keyof paths,
      Method extends keyof paths[Path],
      RequestBody = paths[Path][Method] extends { requestBody: { content: { 'application/json': infer T } } } ? T : never,
      ResponseBody = paths[Path][Method] extends { responses: { 200: { content: { 'application/json': infer T } } } } ? T : unknown
    >(
      method: Method,
      path: Path,
      init?: RequestInit & { body?: RequestBody }
    ): Promise<ResponseBody> {
      const url = `${baseUrl}${String(path)}`;
      
      const response = await fetch(url, {
        method: String(method).toUpperCase(),
        headers: defaultHeaders,
        body: init?.body ? JSON.stringify(init.body) : undefined,
        ...init,
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const contentType = response.headers.get('content-type');
      if (contentType?.includes('application/json')) {
        return await response.json();
      }

      return {} as ResponseBody;
    }
  };
}
