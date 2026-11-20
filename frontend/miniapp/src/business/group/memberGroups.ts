import type { Request } from "../../identity/Session";

export function listMemberGroups(request: Request, signal: AbortSignal, cursor?: string) {
  return request(
    (client, signal) => client.GET("/api/v1/groups", { signal, params: { query: { cursor } } }),
    { signal },
  );
}
