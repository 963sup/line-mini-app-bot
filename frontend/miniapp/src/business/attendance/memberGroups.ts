import { listMemberGroups } from "@/business/group/memberGroups";
import type { Request } from "../../identity/Session";

export async function memberGroups(request: Request, signal: AbortSignal, cursor?: string) {
  const page = await listMemberGroups(request, signal, cursor);
  return {
    items: page.items.map(({ groupId, name }) => ({ id: groupId, name: name || groupId })),
    nextCursor: page.nextCursor,
  };
}
