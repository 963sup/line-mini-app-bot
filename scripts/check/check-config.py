"""Validate repository configuration without installing tools or contacting services."""
import json
import re
from pathlib import Path
import subprocess
import sys

if sys.version_info < (3, 11):
    raise SystemExit("Configuration checks require Python 3.11+.")
import tomllib

ROOT = Path(__file__).resolve().parents[2]


def check():
    codex_dir = ROOT / ".codex"
    config = tomllib.loads((codex_dir / "config.toml").read_text())
    mise = tomllib.loads((ROOT / "mise.toml").read_text())
    if "agents" in config or list((codex_dir / "agents").glob("*.toml")):
        raise ValueError("Subagent registrations must be removed")
    capabilities = ROOT / "docs/development/workflow/capabilities"
    for name in ("README", "planning", "architect", "explorer", "implementer", "reviewer", "tester", "documentation"):
        if not (capabilities / (name + ".md")).is_file():
            raise ValueError(f"Missing capability guide: {name}")
    for guide in capabilities.glob("*.md"):
        for reference in re.findall(r"docs/[A-Za-z0-9_./-]+\.md", guide.read_text()):
            if not (ROOT / reference).is_file():
                raise ValueError(f"Missing capability reference: {reference}")
    docs = list((ROOT / "docs").rglob("*.md"))
    for path in docs:
        # Preserve upstream Markdown: /docs/... links belong to ui.shadcn.com.
        # The local snapshot INDEX remains checked here; page hashes/coverage
        # are checked by shadcn-reference.py --check.
        if path.is_relative_to(ROOT / "docs/development/tooling/frontend/shadcn-reference/snapshot/pages"):
            continue
        for target in re.findall(r"\]\(([^)]+)\)", path.read_text()):
            if "://" in target or target.startswith("#"):
                continue
            if not (path.parent / target.split("#", 1)[0]).is_file():
                raise ValueError(f"Broken documentation link in {path.relative_to(ROOT)}: {target}")
    if not isinstance(mise.get("tools", {}).get("go"), str):
        raise ValueError("mise.toml must pin Go")
    package = json.loads((ROOT / "package.json").read_text())
    if package.get("packageManager") != "pnpm@" + mise["tools"]["npm:pnpm"]:
        raise ValueError("packageManager must match the pnpm version in mise.toml")
    if package.get("private") is not True:
        raise ValueError("The repository workspace must be private")
    dockerfile = (ROOT / "Dockerfile").read_text()
    for image, tool in (("node", "node"), ("golang", "go")):
        version = re.escape(mise["tools"][tool])
        if not re.search(r"^FROM " + image + ":" + version + r"-bookworm(?:-slim)?@sha256:[0-9a-f]{64}", dockerfile, re.M):
            raise ValueError(f"Dockerfile {image} version/digest must match mise.toml")
    if "npm install --global pnpm@" + mise["tools"]["npm:pnpm"] not in dockerfile:
        raise ValueError("Dockerfile pnpm must match mise.toml")
    module = (ROOT / "backend/go.mod").read_text()
    if not re.search(r"^go " + re.escape(mise["tools"]["go"]) + r"$", module, re.M):
        raise ValueError("backend/go.mod must match the pinned Go toolchain")
    for name in ("backend", "frontend/miniapp", "api/client"):
        manifest = json.loads((ROOT / name / "package.json").read_text())
        if manifest.get("private") is not True:
            raise ValueError(f"Workspace package must be private: {name}")
    # Reuse the same constrained parser as bootstrap rather than silently accepting
    # env constructs that would work in mise but fail before mise is installed.
    from importlib.util import module_from_spec, spec_from_file_location
    spec = spec_from_file_location("render_env", ROOT / "scripts/setup/render-env.py")
    module = module_from_spec(spec)
    spec.loader.exec_module(module)
    module.mise_version(mise)
    module.render_exports(mise["env"], {})
    json_paths = list((ROOT / ".vscode").glob("*.json"))
    json_paths += list(codex_dir.glob("*.json")) + list(ROOT.glob("*.config.json"))
    for path in json_paths:
        json.loads(path.read_text())
    scripts = sorted((ROOT / "scripts").rglob("*.sh"))
    for path in scripts:
        subprocess.run(["bash", "-n", str(path)], check=True)
    print(f"PASS: single-agent capabilities, mise env/tool config, {len(json_paths)} JSON files and {len(scripts)} shell scripts")


if __name__ == "__main__":
    try:
        check()
    except (OSError, ValueError, KeyError, TypeError, subprocess.CalledProcessError) as error:
        raise SystemExit(f"Configuration check failed: {error}") from None
