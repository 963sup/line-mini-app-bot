#!/usr/bin/env python3
"""Classify the Git universe; optionally run isolated, cross-checked CLI analysis."""
import argparse
import fnmatch
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import tempfile
import tomllib
import xml.etree.ElementTree as ET

ROOT = Path(__file__).resolve().parents[1]


def git(*args):
    return subprocess.check_output(["git", *args], cwd=ROOT)


def sha256_file(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def build_manifest():
    config = tomllib.loads((ROOT / "analysis-scope.toml").read_text())
    paths = set(git("ls-files", "-z").decode().split("\0")) - {""}
    paths.difference_update(git("ls-files", "--deleted", "-z").decode().split("\0"))
    paths.update(config["additional"])
    rows = []
    for name in sorted(paths):
        path = ROOT / name
        rule = next((r for r in config["exclude"] if fnmatch.fnmatchcase(name, r["pattern"])), None)
        row = {"path": name}
        if rule:
            if not all(rule.get(k) for k in ("reason", "impact", "owner")):
                raise ValueError(f"Incomplete exclusion: {rule}")
            row.update(status="EXCLUDED", exclusion=rule)
        elif path.is_symlink() or not path.is_file():
            raise ValueError(f"Missing or symlink source: {name}")
        elif path.suffix in config["managed_suffixes"] or path.name in config["managed_names"]:
            row.update(status="MANAGED", sha256=sha256_file(path), required=["repomix"],
                       graphify="optional; observed coverage only")
        else:
            raise ValueError(f"Unclassified file: {name}; extend the scope contract")
        rows.append(row)
    return {"schema": config["schema"], "contract_sha256": sha256_file(ROOT / "analysis-scope.toml"),
            "commit": git("rev-parse", "HEAD").decode().strip(),
            "dirty": bool(git("status", "--porcelain")), "files": rows}


def run_audit(manifest):
    out = Path(tempfile.mkdtemp(prefix="analysis-audit-"))
    source = out / "source"
    source.mkdir()
    managed = {r["path"] for r in manifest["files"] if r["status"] == "MANAGED"}
    for name in managed:
        target = source / name
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(ROOT / name, target)
    # An isolated source tree makes the contract authoritative. No project ignores
    # or tool config can silently remove files; security scanning remains enabled.
    wrapper = ["bash", str(ROOT / "scripts/tool-exec.sh")]
    repomix = subprocess.check_output(wrapper + ["which", "repomix"], cwd=ROOT, text=True).strip()
    config = out / "repomix.json"
    config.write_text(json.dumps({"ignore": {"useGitignore": False, "useDefaultPatterns": False,
                                             "customPatterns": []}}))
    # Keep security scanning and verify the final bundle: tool-level ignores
    # must never silently override the scope contract.
    bundle = out / "repomix.xml"
    command = [repomix, str(source), "--config", str(config),
                        "--no-gitignore", "--no-default-patterns", "--style", "xml",
                        "--parsable-style", "--output", str(bundle)]
    with (out / "repomix.log").open("w") as log:
        subprocess.run(command, cwd=out, stdout=log, stderr=log, check=True)
    bundled = {e.attrib["path"] for e in ET.parse(bundle).iter("file") if "path" in e.attrib}
    with (out / "graphify.log").open("w") as log:
        subprocess.run(wrapper + ["graphify", "extract", str(source), "--code-only",
                                  "--out", str(out)], cwd=ROOT, stdout=log, stderr=log, check=True)
    graph = json.loads((out / "graphify-out/graph.json").read_text())
    nodes = graph.get("nodes", [])
    edges = graph.get("links", graph.get("edges", []))
    ids = {n["id"] for n in nodes}
    dangling = [e for e in edges if e["source"] not in ids or e["target"] not in ids]
    observed = {n.get("source_file", "") for n in nodes} - {""}
    current = build_manifest()
    stale = current != manifest
    report = {"manifest": manifest, "repomix_missing": sorted(managed - bundled),
              "repomix_extra": sorted(bundled - managed), "graph_nodes": len(nodes),
              "graph_edges": len(edges), "graph_source_values": sorted(observed),
              "graph_missing_sources": sorted(managed - observed),
              "dangling_edges": len(dangling), "duplicate_ids": len(nodes) - len(ids),
              "stale": stale}
    report["tool_versions"] = {
        tool: subprocess.check_output(wrapper + [tool, "--version"], cwd=ROOT, text=True).strip()
        for tool in ("repomix", "graphify")
    }
    report["artifact_sha256"] = {"repomix.xml": sha256_file(bundle),
                                  "graphify-out/graph.json": sha256_file(out / "graphify-out/graph.json")}
    (out / "report.json").write_text(json.dumps(report, indent=2) + "\n")
    print(f"Audit evidence: {out}")
    if report["repomix_missing"] or report["repomix_extra"] or dangling or len(ids) != len(nodes) or stale:
        raise ValueError("Audit failed; inspect report.json (missing coverage is never success)")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--audit", action="store_true")
    parser.add_argument("--verify", type=Path, help="Reject stale or altered saved report evidence")
    args = parser.parse_args()
    manifest = build_manifest()
    print(f"Scope: {len(manifest['files'])} classified files")
    if args.verify:
        report = json.loads(args.verify.read_text())
        if report["manifest"] != manifest or report["stale"]:
            raise ValueError("Saved evidence is stale")
        for name, expected in report["artifact_sha256"].items():
            if sha256_file(args.verify.parent / name) != expected:
                raise ValueError(f"Altered analysis artifact: {name}")
        if report["repomix_missing"] or report["repomix_extra"] or report["dangling_edges"] or report["duplicate_ids"]:
            raise ValueError("Saved audit did not pass")
        print("Saved evidence matches current sources and artifacts")
    if args.audit:
        run_audit(manifest)


if __name__ == "__main__":
    main()
