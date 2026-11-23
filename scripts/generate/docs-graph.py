#!/usr/bin/env python3
"""Build a Graphify-readable Markdown heading/link graph, without an LLM."""
import argparse
from functools import cache
import runpy
import json
from pathlib import Path
import re
import subprocess
from urllib.parse import unquote, urlsplit


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--out", type=Path, default=Path("/workspace/tooling/cache/graphify-docs"))
    args = parser.parse_args()
    root = Path(__file__).resolve().parents[2]
    # Use Git's ignore rules so installed dependencies and build output stay out.
    paths = subprocess.check_output(
        ["git", "ls-files", "-z", "--cached", "--others", "--exclude-standard",
         "--", "*.md", "*.mermaid"], cwd=root,
    ).decode().split("\0")
    upstream = (".agents/skills/", "docs/development/tooling/frontend/shadcn-reference/snapshot/")
    files = sorted(root / name for name in set(paths)
                   if name and not name.startswith(upstream) and (root / name).is_file())
    documentation = runpy.run_path(str(root / "scripts/check/check-config.py"))
    documentation_targets = documentation["documentation_targets"]
    nodes, edges, coverage, unresolved = [], [], {}, []
    resolve = cache(Path.resolve)  # Reuse link targets only within this build.
    ids = {resolve(path): path.relative_to(root).as_posix() for path in files}
    for path in files:
        source = ids[resolve(path)]
        text = path.read_text(encoding="utf-8")
        lines = text.splitlines()
        coverage[source] = len(lines)
        nodes.append({"id": source, "label": source, "file_type": "document",
                      "source_file": source, "source_location": "document", "_origin": "structural_docs"})
        for number, line in enumerate(documentation["markdown_content"](text).splitlines(), 1):
            heading = None if path.suffix == ".mermaid" else re.match(r"^#{1,6}\s+(.+?)\s*#*\s*$", line)
            if heading:
                node_id = f"{source}#L{number}"
                nodes.append({"id": node_id, "label": heading[1], "file_type": "concept",
                              "source_file": source, "source_location": f"heading at line {number}",
                              "_origin": "structural_docs"})
                edges.append({"source": source, "target": node_id, "relation": "contains",
                              "confidence": "EXTRACTED", "confidence_score": 1.0,
                              "source_file": source, "source_location": f"line {number}"})
        for number, destination in documentation_targets(text, path.suffix == ".mermaid"):
            url = urlsplit(destination)
            if url.scheme or url.netloc or not url.path:
                continue
            target = resolve(path.parent / unquote(url.path))
            if target in ids:
                edges.append({"source": source, "target": ids[target], "relation": "references",
                              "confidence": "EXTRACTED", "confidence_score": 1.0,
                              "source_file": source, "source_location": f"line {number}"})
            elif not target.exists():
                unresolved.append({"file": source, "line": number, "target": destination})
    # Keep repeated citations as provenance rather than parallel edges: Graphify
    # may load a DiGraph and otherwise discard their additional locations.
    unique = {}
    for edge in edges:
        key = (edge["source"], edge["target"], edge["relation"])
        if key not in unique:
            unique[key] = {**edge, "source_locations": []}
        unique[key]["source_locations"].append(edge["source_location"])
    edges = list(unique.values())
    args.out.mkdir(parents=True, exist_ok=True)
    graph = {"directed": True, "multigraph": False,
             "graph": {"schema_version": 1, "scope": "maintained_project_docs", "extraction": "headings_and_local_destinations",
                       "semantic_extraction": False, "llm_tokens": 0}, "nodes": nodes, "links": edges}
    (args.out / "graph.json").write_text(json.dumps(graph, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    (args.out / "coverage.json").write_text(json.dumps({"files": coverage, "unresolved_links": unresolved,
        "limits": "ATX headings and inline/reference/HTML/Mermaid local links; no semantic analysis or fragment validation"},
        ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"Docs: {len(files)} files, {sum(coverage.values())} lines; {len(nodes)} nodes, {len(edges)} edges")
    print(f"Unresolved local links: {len(unresolved)}; LLM tokens: 0; semantic extraction: false")
    print(args.out / "graph.json")


if __name__ == "__main__":
    main()
