"""Use Graphify CLI to refresh a small documentation subgraph for the turn."""

from hashlib import sha256
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys
from tempfile import TemporaryDirectory
import tomllib


ROOT = Path(__file__).resolve().parents[2]
DEFAULT_SEED = "docs/development/README.md"
TOPICS = (
    (("repomix", "graphify", "hook", "mcp"), "docs/development/tooling/README.md"),
    (("cloud run", "cloud build", "部署"), "docs/development/recipes/diagnose-deployment.md"),
    (("前端", "mini app", "next.js"), "docs/development/recipes/change-frontend.md"),
    (("ddd", "聚合", "領域"), "docs/development/recipes/change-domain.md"),
    (("openapi", "契約", "http api"), "docs/development/recipes/change-api.md"),
    (("交接", "handoff"), "docs/development/workflow/handoff.md"),
)


def choose_seed(prompt: str, documents: set[str]) -> str:
    for name in re.findall(r"[\w./-]+\.md\b", prompt):
        if name.removeprefix("./") in documents:
            return name.removeprefix("./")
    for terms, path in TOPICS:
        if path in documents and any(term in prompt.casefold() for term in terms):
            return path
    return DEFAULT_SEED


def graphify_command() -> list[str]:
    settings = tomllib.loads((ROOT / "mise.toml").read_text())["env"]
    wrapper = ["bash", "scripts/tooling/tool-exec.sh"]
    search_path = os.pathsep.join((settings["UV_TOOL_BIN_DIR"], os.defpath, os.environ.get("PATH", "")))
    if shutil.which("graphify", path=search_path):
        return [*wrapper, "graphify"]
    return [
        *wrapper, "--tool", "python", "--tool", "uv", "uv", "tool", "run",
        "--no-python-downloads", "--from", settings["GRAPHIFY_PACKAGE"], "graphify",
    ]


def refresh(request: dict[str, str]) -> tuple[Path, str]:
    output = ROOT / ".agents" / "graphify-subgraphs"
    for directory in (output.parent, output):
        if directory.is_symlink():
            raise ValueError("the Graphify destination contains a symbolic link")
        directory.mkdir(exist_ok=True)
    session = sha256(request.get("session_id", "manual").encode()).hexdigest()[:20]
    artifact = output / f"{session}.txt"

    with TemporaryDirectory(prefix=".graphify-", dir=output.parent) as temporary:
        work = Path(temporary)
        subprocess.run(
            [sys.executable, "scripts/generate/docs-graph.py", "--out", str(work)],
            cwd=ROOT, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, check=True, timeout=10,
        )
        graph = work / "graph.json"
        nodes = json.loads(graph.read_text())["nodes"]
        seed = choose_seed(request.get("prompt", ""), {node["id"] for node in nodes if node["file_type"] == "document"})
        result = subprocess.run(
            [*graphify_command(), "query", "Project documentation references", "--graph", str(graph),
             "--seed", seed, "--depth", "1", "--budget", "1000"],
            cwd=ROOT, stdout=subprocess.PIPE, stderr=subprocess.DEVNULL, text=True, check=True, timeout=42,
        )
        if not result.stdout.strip():
            raise ValueError("Graphify returned an empty subgraph")
        draft = work / "subgraph.txt"
        # Graphify reports the temporary graph path in its first line.
        draft.write_text(f"Documentation subgraph for {seed}\n" + result.stdout.partition("\n")[2])
        draft.replace(artifact)
    return artifact, seed


def main() -> None:
    try:
        artifact, seed = refresh(json.load(sys.stdin))
    except subprocess.TimeoutExpired:
        result = {"systemMessage": "Graphify subgraph refresh timed out. Use the current source files."}
    except subprocess.CalledProcessError as error:
        result = {"systemMessage": f"Graphify subgraph refresh failed (exit {error.returncode})."}
    except (OSError, KeyError, TypeError, ValueError) as error:
        result = {"systemMessage": f"Graphify subgraph refresh skipped: {error}."}
    else:
        result = {
            "hookSpecificOutput": {
                "hookEventName": "UserPromptSubmit",
                "additionalContext": f"Graphify refreshed a documentation subgraph for {seed}. Read {artifact} for related source locations.",
            },
        }
    print(json.dumps(result, ensure_ascii=False))


if __name__ == "__main__":
    main()
