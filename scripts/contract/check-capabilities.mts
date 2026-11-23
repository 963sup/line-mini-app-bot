import assert from "node:assert/strict";
import { readdirSync, readFileSync, realpathSync, statSync } from "node:fs";
import { dirname, join, resolve, sep } from "node:path";
import { checkDispatch, type DispatchOperation } from "./check-dispatch.mts";

import { checkResourceSemantics, type ResourceEntry } from "./check-resource-semantics.mts";

type Json = null | boolean | number | string | Json[] | { [key: string]: Json };
interface Operation extends DispatchOperation {
  owner: string;
  core: string;
  specification: string;
  consumers: string[];
}
interface Inventory {
  schemaVersion: number;
  operations: Operation[];
  capabilities: Record<string, { contract?: string }>;
  resources: ResourceEntry[];
  namingExceptions: Record<string, string>;
  events: {
    owner: string;
    contract: string;
    adapter: string;
    handler: string;
    action: string;
    note: string;
  }[];
}
const root = realpathSync(resolve(import.meta.dirname, "../.."));
const contract = join(root, "api/contract");
const methods = new Set(["get", "post", "put", "patch", "delete", "head", "options", "trace"]);

function source(name: string): string {
  const path = realpathSync(resolve(root, name));
  assert(path.startsWith(root + sep) && statSync(path).isFile(), `Missing source: ${name}`);
  return readFileSync(path, "utf8");
}

function record(value: Json | undefined): Record<string, Json> {
  assert(
    value !== null && typeof value === "object" && !Array.isArray(value),
    "Expected JSON object",
  );
  return value;
}

function* references(value: Json): Generator<string> {
  if (Array.isArray(value)) {
    for (const child of value) yield* references(child);
  } else if (value && typeof value === "object") {
    if (Object.hasOwn(value, "$ref")) {
      assert(typeof value.$ref === "string", "Expected string reference");
      yield value.$ref;
    }
    for (const child of Object.values(value)) yield* references(child);
  }
}

function pointer(value: Json | undefined, fragment: string): Json {
  assert(value !== undefined, "Missing contract document");
  if (!fragment) return value;
  assert(fragment.startsWith("/"), `Invalid JSON pointer: ${fragment}`);
  for (const part of fragment.slice(1).split("/")) {
    const key = part.replaceAll("~1", "/").replaceAll("~0", "~");
    assert(
      value !== null && typeof value === "object" && Object.hasOwn(value, key),
      `Missing JSON pointer: ${fragment}`,
    );
    value = Array.isArray(value) ? value[Number(key)] : value[key];
    assert(value !== undefined, `Missing JSON pointer: ${fragment}`);
  }
  return value;
}

function reference(path: string, ref: string): [string, string] {
  const hash = ref.indexOf("#");
  const file = hash < 0 ? ref : ref.slice(0, hash);
  return [
    file ? realpathSync(resolve(dirname(path), file)) : path,
    hash < 0 ? "" : ref.slice(hash + 1),
  ];
}

function hasHandler(text: string, handler: string): boolean {
  assert(/^[A-Za-z_]\w*$/.test(handler), `Invalid handler name: ${handler}`);
  return new RegExp(String.raw`func\s+(?:\([^)]*\)\s*)?${handler}\(`).test(text);
}

const inventory = JSON.parse(source("api/contract/capabilities.json")) as Inventory;
assert.equal(inventory.schemaVersion, 2, "Unsupported inventory format");
const docs = new Map<string, Json>();
for (const entry of readdirSync(contract, { recursive: true, withFileTypes: true })) {
  if (!entry.isFile() || !entry.name.endsWith(".yaml")) continue;
  const path = realpathSync(join(entry.parentPath, entry.name));
  docs.set(path, JSON.parse(readFileSync(path, "utf8")) as Json);
}
const roots = new Map([
  [join(contract, "openapi.yaml"), "public"],
  [join(contract, "internal-openapi.yaml"), "internal"],
]);
const graph = new Map<string, Set<string>>();
for (const [path, document] of docs) {
  const dependencies = new Set<string>();
  for (const ref of references(document)) {
    const [target, fragment] = reference(path, ref);
    assert(docs.has(target), `External or missing reference: ${path}: ${ref}`);
    pointer(docs.get(target), fragment);
    if (target !== path) dependencies.add(target);
    assert(
      roots.has(path) || !roots.has(target),
      `Capability references composition root: ${path}: ${ref}`,
    );
  }
  graph.set(path, dependencies);
}
const visiting = new Set<string>();
const visited = new Set<string>();
function visit(path: string) {
  assert(!visiting.has(path), `Cyclic contract dependency: ${path}`);
  if (visited.has(path)) return;
  visiting.add(path);
  for (const target of graph.get(path) ?? []) visit(target);
  visiting.delete(path);
  visited.add(path);
}
for (const path of docs.keys()) visit(path);

const operations = new Map<string, [string, string, string, string]>();
function onlyRef(value: Json): string {
  const item = record(value);
  assert(
    Object.keys(item).length === 1 && typeof item.$ref === "string",
    "Expected a reference-only entry in composition root",
  );
  return item.$ref;
}
for (const [path, audience] of roots) {
  const document = record(docs.get(path));
  for (const group of Object.values(record(document.components ?? {}))) {
    for (const item of Object.values(record(group))) onlyRef(item);
  }
  for (const [route, item] of Object.entries(record(document.paths))) {
    const [target, fragment] = reference(path, onlyRef(item));
    for (const [method, operation] of Object.entries(record(pointer(docs.get(target), fragment)))) {
      if (!methods.has(method)) continue;
      const id = record(operation).operationId;
      assert(
        typeof id === "string" && !operations.has(id),
        `Missing or duplicate operationId: ${id}`,
      );
      operations.set(id, [audience, method.toUpperCase(), route, target]);
    }
  }
}

const registered = new Set<string>();
for (const operation of inventory.operations) {
  const id = operation.operationId;
  assert(!registered.has(id), `Duplicate inventory operation: ${id}`);
  registered.add(id);
  const file = inventory.capabilities[operation.owner]?.contract;
  assert(file, `Missing capability contract: ${id}`);
  assert.deepEqual(
    operations.get(id),
    [operation.audience, operation.method, operation.path, realpathSync(resolve(contract, file))],
    `Operation ownership/route mismatch: ${id}`,
  );
  const text = source(operation.adapter);
  assert(hasHandler(text, operation.handler), `Missing adapter method: ${id}`);
  assert(
    !["http.NewServeMux()", "r.URL.Query().Get(", 'r.PathValue("'].some((part) =>
      text.includes(part),
    ),
    `Adapter re-parses generated routing: ${id}`,
  );
  source(operation.core);
  source(operation.specification);
  for (const consumer of operation.consumers) {
    const text = source(consumer);
    assert(
      operation.audience !== "public" || text.includes(`"${operation.path}"`),
      `Stale consumer: ${id}: ${consumer}`,
    );
  }
}
checkDispatch(source, inventory.operations);
assert.deepEqual(registered, new Set(operations.keys()), "Unowned operations");
const owned = Object.values(inventory.capabilities).flatMap((capability) =>
  capability.contract ? [realpathSync(resolve(contract, capability.contract))] : [],
);
assert.deepEqual(
  new Set(docs.keys()),
  new Set([...roots.keys(), ...owned, join(contract, "common/transport.yaml")]),
  "Unclassified contract source (assign its capability first)",
);

checkResourceSemantics(inventory.resources, inventory.capabilities, inventory.operations);
for (const { entity, collection, path } of inventory.resources) {
  assert(
    /^[A-Z]/.test(entity) && entity.toLowerCase() !== collection && collection.endsWith("s"),
    `Review singular entity/plural collection: ${entity}/${collection}`,
  );
  assert.equal(path.split("/").at(-1), collection, `Collection path mismatch: ${path}`);
  assert(
    [...docs.values()].some((doc) =>
      Object.hasOwn(record(record(record(doc).components ?? {}).schemas ?? {}), entity),
    ),
    `Missing resource schema: ${entity}`,
  );
}
for (const [, , route] of operations.values()) {
  assert(
    Object.hasOwn(inventory.namingExceptions, route) ||
      inventory.resources.some(({ path }) => route === path || route.startsWith(path + "/")),
    `Unclassified resource path or naming exception: ${route}`,
  );
}
for (const event of inventory.events) {
  assert(Object.hasOwn(inventory.capabilities, event.owner), "Unknown event capability");
  source(event.contract);
  assert(hasHandler(source(event.adapter), event.handler), "Missing event adapter");
  assert(event.owner && event.action && event.note, "Incomplete event contract");
}
console.log(
  `PASS: ${operations.size} operations, ${docs.size} contract sources, ${inventory.events.length} event entries; ownership, references, naming and dispatch`,
);
