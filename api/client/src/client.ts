import createClient from 'openapi-fetch';
import type { paths } from './generated/schema';
export function createAPIClient(baseUrl: string, token?: string) {
 return createClient<paths>({ baseUrl, headers: token ? { Authorization: `Bearer ${token}` } : undefined, cache: 'no-store', credentials: 'omit' });
}
export type { paths, components, operations } from './generated/schema';
