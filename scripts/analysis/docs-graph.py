#!/usr/bin/env python3
"""Build a Graphify-readable Markdown heading/link graph, without an LLM."""
import argparse
import json
from pathlib import Path
import re
from urllib.parse import unquote, urlsplit


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--out", type=Path, default=Path("/workspace/tooling/cache/graphify-docs"))
    args = parser.parse_args()
    root = Path(__file__).resolve().parents[2]
    # Include maintained guides outside docs/ so entry points cannot silently drift.
    files = sorted(set((root / "docs").rglob("*.md"))
                   | set(root.glob("*.md"))
                   | set((root / "scripts").rglob("*.md"))
                   | set((root / ".agents").glob("*.md"))
                   | set((root / ".github").rglob("*.md")))
    nodes, edges, coverage, unresolved = [], [], {}, []
    ids = {path.resolve(): path.relative_to(root).as_posix() for path in files}
    for path in files:
        source = ids[path.resolve()]
        lines = path.read_text(encoding="utf-8").splitlines()
        coverage[source] = len(lines)
        nodes.append({"id": source, "label": source, "file_type": "document",
                      "source_file": source, "source_location": "document", "_origin": "structural_docs"})
        in_fence = False
        for number, line in enumerate(lines, 1):
            if re.match(r"^\s*(`{3,}|~{3,})", line):
                in_fence = not in_fence
                continue
            if in_fence:
                continue
            heading = re.match(r"^#{1,6}\s+(.+?)\s*#*\s*$", line)
            if heading:
                node_id = f"{source}#L{number}"
                nodes.append({"id": node_id, "label": heading[1], "file_type": "concept",
                              "source_file": source, "source_location": f"heading at line {number}",
                              "_origin": "structural_docs"})
                edges.append({"source": source, "target": node_id, "relation": "contains",
                              "confidence": "EXTRACTED", "confidence_score": 1.0,
                              "source_file": source, "source_location": f"line {number}"})
            for match in re.finditer(r"(?<!!)\[[^\]]+\]\(([^\s)]+)\)", line):
                url = urlsplit(match[1].strip("<>"))
                if url.scheme or url.netloc or not url.path:
                    continue
                target = (path.parent / unquote(url.path)).resolve()
                if target in ids:
                    edges.append({"source": source, "target": ids[target], "relation": "references",
                                  "confidence": "EXTRACTED", "confidence_score": 1.0,
                                  "source_file": source, "source_location": f"line {number}"})
                elif not target.exists():
                    unresolved.append({"file": source, "line": number, "target": match[1]})
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
             "graph": {"schema_version": 1, "scope": "maintained_project_docs", "extraction": "headings_and_inline_links",
                       "semantic_extraction": False, "llm_tokens": 0}, "nodes": nodes, "links": edges}
    (args.out / "graph.json").write_text(json.dumps(graph, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    (args.out / "coverage.json").write_text(json.dumps({"files": coverage, "unresolved_links": unresolved,
        "limits": "ATX headings and inline local links only; no semantic analysis or fragment validation"},
        ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"Docs: {len(files)} files, {sum(coverage.values())} lines; {len(nodes)} nodes, {len(edges)} edges")
    print(f"Unresolved local links: {len(unresolved)}; LLM tokens: 0; semantic extraction: false")
    print(args.out / "graph.json")


if __name__ == "__main__":
    main()
