import createClient from "openapi-fetch";
import type { paths } from "./generated/schema";
export function createAPIClient(baseUrl: string, token?: string, csrfToken?: string) {
  return createClient<paths>({
    baseUrl,
    headers: {
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(csrfToken ? { "X-CSRF-Token": csrfToken } : {}),
    },
    cache: "no-store",
    credentials: "same-origin",
  });
}
import type { paths as LocationPaths } from "./generated/geolocation";
export function createLocationClient(baseUrl: string) {
  return createClient<LocationPaths>({ baseUrl, credentials: "omit", cache: "no-store" });
}
export type { paths, components, operations } from "./generated/schema";
