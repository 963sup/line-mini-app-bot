import assert from "node:assert/strict";

export interface ResourceEntry {
  entity: string;
  collection: string;
  path: string;
  owner: string;
  kind: "resource" | "projection" | "command-result";
}

// This classifies HTTP representations, not DDD aggregates or storage collections.
export function checkResourceSemantics(
  resources: ResourceEntry[],
  capabilities: Record<string, unknown>,
  operations: { path: string; owner: string }[],
): void {
  const paths = new Set<string>();
  for (const resource of resources) {
    assert(Object.hasOwn(capabilities, resource.owner), `Unknown resource owner: ${resource.path}`);
    assert(
      ["resource", "projection", "command-result"].includes(resource.kind),
      `Unknown resource representation: ${resource.path}`,
    );
    assert(!paths.has(resource.path), `Duplicate resource path: ${resource.path}`);
    paths.add(resource.path);
    assert(
      operations.some(({ path, owner }) => path === resource.path && owner === resource.owner),
      `Resource owner differs from its operation: ${resource.path}`,
    );
  }
}
