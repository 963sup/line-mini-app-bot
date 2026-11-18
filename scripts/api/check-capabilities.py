#!/usr/bin/env python3
"""Check the explicit API capability inventory against source contracts and adapters.

Source contracts intentionally use JSON syntax, a subset of YAML. This check
proves structural coverage, not runtime authorization or business correctness.
"""
import json
import re
from pathlib import Path
from check_dispatch import check_dispatch

ROOT = Path(__file__).resolve().parents[2]
CONTRACT = ROOT / "api/contract"
METHODS = {"get", "post", "put", "patch", "delete", "head", "options", "trace"}


def require(condition, message):
    if not condition:
        raise ValueError(message)


def source(name):
    path = (ROOT / name).resolve()
    require(path.is_relative_to(ROOT) and path.is_file(), f"Missing source: {name}")
    return path.read_text()


def references(value):
    if isinstance(value, dict):
        if "$ref" in value:
            yield value["$ref"]
        for child in value.values():
            yield from references(child)
    elif isinstance(value, list):
        for child in value:
            yield from references(child)


def pointer(value, fragment):
    for part in fragment.lstrip("/").split("/") if fragment else []:
        value = value[part.replace("~1", "/").replace("~0", "~")]
    return value


def main():
    inventory = json.loads(source("api/capabilities.json"))
    require(inventory["schemaVersion"] == 1, "Unsupported inventory format")
    docs = {p.resolve(): json.loads(p.read_text()) for p in CONTRACT.rglob("*.yaml")}
    roots = {(CONTRACT / name).resolve(): audience for name, audience in
             [("openapi.yaml", "public"), ("internal-openapi.yaml", "internal")]}
    graph = {p: set() for p in docs}
    for path, document in docs.items():
        for ref in references(document):
            file, _, fragment = ref.partition("#")
            target = (path.parent / file).resolve() if file else path
            require(target in docs, f"External or missing reference: {path}: {ref}")
            pointer(docs[target], fragment)
            if target != path:
                graph[path].add(target)
            require(path in roots or target not in roots,
                    f"Capability references composition root: {path}: {ref}")
    visiting, visited = set(), set()

    def visit(path):
        require(path not in visiting, f"Cyclic contract dependency: {path}")
        if path in visited:
            return
        visiting.add(path)
        for target in graph[path]:
            visit(target)
        visiting.remove(path)
        visited.add(path)

    for path in docs:
        visit(path)
    operations = {}
    for path, audience in roots.items():
        doc = docs[path]
        require(all(set(item) == {"$ref"} for item in doc["paths"].values()),
                f"Inline capability operation in root: {path}")
        require(all(set(item) == {"$ref"} for group in doc.get("components", {}).values()
                    for item in group.values()), f"Inline component in root: {path}")
        for route, item in doc["paths"].items():
            file, _, fragment = item["$ref"].partition("#")
            target = (path.parent / file).resolve()
            for method, operation in pointer(docs[target], fragment).items():
                if method not in METHODS:
                    continue
                oid = operation["operationId"]
                require(oid not in operations, f"Duplicate operationId: {oid}")
                operations[oid] = (audience, method.upper(), route, target)
    registered = set()
    for operation in inventory["operations"]:
        oid = operation["operationId"]
        require(oid not in registered, f"Duplicate inventory operation: {oid}")
        registered.add(oid)
        cap = inventory["capabilities"][operation["owner"]]
        expected = (operation["audience"], operation["method"], operation["path"],
                    (CONTRACT / cap["contract"]).resolve())
        require(operations.get(oid) == expected, f"Operation ownership/route mismatch: {oid}")
        text = source(operation["adapter"])
        require(re.search(r"func\s+(?:\([^)]*\)\s*)?" + re.escape(operation["handler"]) + r"\(", text),
                f"Missing adapter method: {oid}")
        require("http.NewServeMux()" not in text and "r.URL.Query().Get(" not in text
                and 'r.PathValue("' not in text, f"Adapter re-parses generated routing: {oid}")
        source(operation["core"])
        source(operation["specification"])
        for consumer in operation["consumers"]:
            consumer_text = source(consumer)
            if operation["audience"] == "public":
                require('"' + operation["path"] + '"' in consumer_text, f"Stale consumer: {oid}: {consumer}")
    check_dispatch(source, inventory["operations"])
    require(registered == set(operations), f"Unowned operations: {set(operations) ^ registered}")
    owned_files = {(CONTRACT / cap["contract"]).resolve() for cap in inventory["capabilities"].values() if "contract" in cap}
    require(set(docs) == set(roots) | owned_files | {(CONTRACT / "common/transport.yaml").resolve()},
            "Unclassified contract source (assign its capability first)")
    for resource in inventory["resources"]:
        entity, collection, path = (resource[k] for k in ("entity", "collection", "path"))
        require(entity[0].isupper() and entity.lower() != collection and collection.endswith("s"),
                f"Review singular entity/plural collection: {entity}/{collection}")
        require(path.rsplit("/", 1)[-1] == collection, f"Collection path mismatch: {path}")
        require(any(entity in d.get("components", {}).get("schemas", {}) for d in docs.values()),
                f"Missing resource schema: {entity}")
    for _, _, route, _ in operations.values():
        require(route in inventory["namingExceptions"] or any(
            route == r["path"] or route.startswith(r["path"] + "/") for r in inventory["resources"]),
            f"Unclassified resource path or naming exception: {route}")
    for event in inventory["events"]:
        require(event["owner"] in inventory["capabilities"], "Unknown event capability")
        source(event["contract"])
        text = source(event["adapter"])
        require(re.search(r"func\s+(?:\([^)]*\)\s*)?" + re.escape(event["handler"]) + r"\(", text), "Missing event adapter")
        require(bool(event["owner"] and event["action"] and event["note"]), "Incomplete event contract")
    print(f"PASS: {len(operations)} operations, {len(docs)} contract sources, "
          f"{len(inventory['events'])} event entries; ownership, references, naming and dispatch")


if __name__ == "__main__":
    main()
