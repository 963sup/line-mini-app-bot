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
ROLES = {
    "orchestrator", "architect", "researcher", "explorer",
    "implementer", "debugger", "tester", "reviewer", "documentation", "integrator",
    "graphify", "repomix", "shadcn", "golang", "typescript", "nextjs", "context7",
}


def check():
    codex_dir = ROOT / ".codex"
    config = tomllib.loads((codex_dir / "config.toml").read_text())
    mise = tomllib.loads((ROOT / "mise.toml").read_text())
    if config.get("features", {}).get("multi_agent") is not True:
        raise ValueError("Codex multi_agent must be enabled")
    roles = {name: role for name, role in config["agents"].items() if isinstance(role, dict)}
    if set(roles) != ROLES:
        raise ValueError("Codex role registry must contain the documented roles")
    for name, role in roles.items():
        path = (codex_dir / role["config_file"]).resolve()
        if not path.is_relative_to(codex_dir.resolve()):
            raise ValueError(f"Agent file must remain inside .codex: {name}")
        agent = tomllib.loads(path.read_text())
        if agent.get("name") != name or not agent.get("description", "").strip():
            raise ValueError(f"Missing or mismatched agent identity: {name}")
        instructions = agent.get("developer_instructions")
        if not isinstance(instructions, str) or not instructions.strip():
            raise ValueError(f"Missing agent instructions: {name}")
    # Catch stale role-document paths after moving or renaming documentation.
    for role in roles.values():
        instructions = tomllib.loads((codex_dir / role["config_file"]).read_text())["developer_instructions"]
        for reference in re.findall(r"docs/[A-Za-z0-9_./-]+\.md", instructions):
            if not (ROOT / reference).is_file():
                raise ValueError(f"Missing role document: {reference}")
    docs = list((ROOT / "docs").rglob("*.md"))
    for path in docs:
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
    spec = spec_from_file_location("render_env", ROOT / "scripts/tooling/render-env.py")
    module = module_from_spec(spec)
    spec.loader.exec_module(module)
    module.render_exports(mise["env"], {})
    json_paths = list((ROOT / ".vscode").glob("*.json"))
    json_paths += list(codex_dir.glob("*.json")) + list(ROOT.glob("*.config.json"))
    for path in json_paths:
        json.loads(path.read_text())
    scripts = sorted((ROOT / "scripts").rglob("*.sh"))
    for path in scripts:
        subprocess.run(["bash", "-n", str(path)], check=True)
    print(f"PASS: {len(roles)} agents, mise env/tool config, {len(json_paths)} JSON files and {len(scripts)} shell scripts")


if __name__ == "__main__":
    try:
        check()
    except (OSError, ValueError, KeyError, TypeError, subprocess.CalledProcessError) as error:
        raise SystemExit(f"Configuration check failed: {error}") from None
