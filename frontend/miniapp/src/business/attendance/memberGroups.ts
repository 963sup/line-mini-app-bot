import type { Request } from "../../identity/Session";

export async function memberGroups(request: Request, signal: AbortSignal, cursor?: string) {
  const page = await request(
    (client, signal) => client.GET("/api/v1/groups", { signal, params: { query: { cursor } } }),
    { signal },
  );
  return {
    items: page.items.map(({ groupId, name }) => ({ id: groupId, name: name || groupId })),
    nextCursor: page.nextCursor,
  };
}
