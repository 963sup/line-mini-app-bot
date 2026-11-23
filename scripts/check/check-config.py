"""Validate repository configuration without installing tools or contacting services."""
import ast
import html
from html.parser import HTMLParser
import json
import re
from pathlib import Path
import subprocess
import sys
import tempfile
from urllib.parse import unquote, urlsplit

if sys.version_info < (3, 11):
    raise SystemExit("Configuration checks require Python 3.11+.")
import tomllib

ROOT = Path(__file__).resolve().parents[2]


def markdown_content(text):
    """Ignore fenced examples while preserving line numbers for diagnostics."""
    lines = []
    fence = None
    for line in text.splitlines():
        marker = re.match(r"^ {0,3}(`{3,}|~{3,})", line)
        if fence:
            if marker and marker[1][0] == fence[0] and len(marker[1]) >= len(fence) and not line[marker.end():].strip():
                fence = None
            lines.append("")
        elif marker:
            fence = marker[1]
            lines.append("")
        else:
            lines.append(line)
    return "\n".join(lines)


def mermaid_targets(text):
    for number, line in enumerate(text.splitlines(), 1):
        match = re.match(r'^\s*click\s+\w+\s+(?:href\s+)?["\x27]([^"\x27]+)["\x27]', line)
        if match:
            yield number, html.unescape(match[1])


def documentation_targets(text, standalone_mermaid=False):
    """Yield supported inline, reference, HTML href and Mermaid destinations.

    Reference definitions are validated even when unused. Other fenced examples
    and inline code remain literal, never interpreted as documentation links.
    """
    if standalone_mermaid:
        yield from mermaid_targets(text)
        return
    content = markdown_content(text)
    cleaned = []
    for number, line in enumerate(content.splitlines(), 1):
        line = re.sub(r"(`+).*?\1", "", line)
        cleaned.append(line)
        definition = re.match(r'^ {0,3}\[[^\]]+\]:\s*(<[^>]+>|[^\s]+)', line)
        if definition:
            yield number, definition[1].strip("<>")
        for match in re.finditer(r'!?\[[^\]\n]*\]\((<[^>]+>|[^\s)]+)(?:\s+["\x27][^)]*)?\)', line):
            yield number, match[1].strip("<>")

    class Links(HTMLParser):
        def __init__(self):
            super().__init__(convert_charrefs=True)
            self.targets = []

        def handle_starttag(self, tag, attributes):
            for name, value in attributes:
                if name == "href" and value:
                    self.targets.append((self.getpos()[0], value))

    parser = Links()
    parser.feed("\n".join(cleaned))
    yield from parser.targets
    fence, diagram = None, False
    for number, line in enumerate(text.splitlines(), 1):
        marker = re.match(r"^ {0,3}(`{3,}|~{3,})(.*)$", line)
        if fence:
            if marker and marker[1][0] == fence[0] and len(marker[1]) >= len(fence) and not marker[2].strip():
                fence, diagram = None, False
            elif diagram:
                for _, target in mermaid_targets(line):
                    yield number, target
        elif marker:
            fence, diagram = marker[1], marker[2].strip().lower() == "mermaid"


def markdown_anchors(text):
    content = markdown_content(text)
    anchors = set(re.findall(r'<[^>]+\b(?:id|name)=["\x27]([^"\x27]+)', content))
    generated = set()
    for heading in re.findall(r"^ {0,3}#{1,6}\s+(.+?)\s*#*\s*$", content, re.M):
        heading = re.sub(r"<[^>]*>", "", heading)
        heading = re.sub(r"!?\[([^\]]+)\]\([^)]*\)", r"\1", heading)
        slug = re.sub(r"[^\w\- ]", "", html.unescape(heading).lower()).replace(" ", "-")
        candidate, number = slug, 0
        while candidate in generated:
            number += 1
            candidate = f"{slug}-{number}"
        generated.add(candidate)
    return anchors | generated


def check_documentation_links(root):
    """Check supported owned local destinations and Markdown anchors."""
    names = subprocess.check_output(
        ["git", "ls-files", "-z", "--cached", "--others", "--exclude-standard", "--", "*.md", "*.mermaid"],
        cwd=root,
    ).decode().split("\0")
    excluded = (".agents/", ".claude/", "docs/development/tooling/frontend/shadcn-reference/snapshot/pages/")
    paths = sorted(root / name for name in set(names)
                   if name and not name.startswith(excluded) and (root / name).is_file())
    anchors = {}
    for path in paths:
        for number, target in documentation_targets(path.read_text(), path.suffix == ".mermaid"):
            url = urlsplit(target)
            if url.scheme or url.netloc:
                continue
            destination = (path.parent / unquote(url.path)).resolve() if url.path else path.resolve()
            location = f"{path.relative_to(root)}:{number}"
            if not destination.exists():
                raise ValueError(f"Broken documentation link in {location}: {target}")
            if url.fragment and destination.is_file() and destination.suffix == ".md":
                if destination not in anchors:
                    anchors[destination] = markdown_anchors(destination.read_text())
                if unquote(url.fragment) not in anchors[destination]:
                    raise ValueError(f"Broken documentation anchor in {location}: {target}")
    return len(paths)


def check_cloud_build_sources(root):
    """Check current static script entries and sibling Python dependencies offline.

    Git supplies ignore semantics; gcloud-specific include directives are rejected
    rather than silently approximated. Actual gcloud upload listing remains the
    release acceptance check for SDK-specific behavior.
    """
    root = root.resolve()
    configuration = (root / "cloudbuild.yaml").read_text()
    pending = [root / name for name in re.findall(r"\bscripts/[A-Za-z0-9_./-]+\.(?:py|sh|mts|ts)\b", configuration)]
    required = set()
    while pending:
        path = pending.pop().resolve()
        if path in required:
            continue
        if not path.is_relative_to(root) or not path.is_file():
            raise ValueError(f"Missing Cloud Build source: {path}")
        required.add(path)
        if path.suffix == ".py":
            tree = ast.parse(path.read_text(), filename=str(path))
            for node in ast.walk(tree):
                if (isinstance(node, ast.Call) and isinstance(node.func, ast.Attribute)
                        and node.func.attr == "with_name" and len(node.args) == 1
                        and isinstance(node.args[0], ast.Constant) and isinstance(node.args[0].value, str)
                        and node.args[0].value.endswith(".py")):
                    pending.append(path.with_name(node.args[0].value))
                elif isinstance(node, (ast.Import, ast.ImportFrom)):
                    modules = [alias.name for alias in node.names] if isinstance(node, ast.Import) else [node.module or ""]
                    for module in modules:
                        sibling = path.parent / (module.split(".")[0] + ".py")
                        if sibling.is_file():
                            pending.append(sibling)
    ignore = (root / ".gcloudignore").read_text()
    if re.search(r"(?m)^#!include:", ignore):
        raise ValueError("Cloud Build source check requires explicit .gcloudignore rules; include directives need supported expansion")
    with tempfile.TemporaryDirectory(prefix="cloud-build-sources-") as directory:
        temporary = Path(directory)
        (temporary / ".gitignore").write_text(ignore)
        subprocess.run(["git", "init", "-q", directory], check=True)
        for path in required:
            name = path.relative_to(root)
            (temporary / name).parent.mkdir(parents=True, exist_ok=True)
            (temporary / name).touch()
        names = sorted(path.relative_to(root).as_posix() for path in required)
        result = subprocess.run(["git", "-c", "core.excludesFile=/dev/null", "check-ignore", "--no-index", "--stdin"],
                                cwd=temporary, input="\n".join(names) + "\n", text=True, capture_output=True)
        if result.returncode not in (0, 1):
            raise ValueError("Cannot check Cloud Build upload rules: " + result.stderr)
        if result.stdout.strip():
            raise ValueError("Cloud Build source excluded from upload: " + ", ".join(result.stdout.splitlines()))
    return len(required)


def check_capability_routes(index, guides):
    section = re.search(r"^## (?:Select capabilities by task|依任務選擇能力)[ \t]*\n(.*?)(?=^## |\Z)", index.read_text(), re.M | re.S)
    if not section:
        raise ValueError(f"Missing capability selection section in {index}")
    routes = set()
    for row in section[1].splitlines():
        if row.lstrip().startswith("|"):
            for target in re.findall(r"\]\(([^)#]+)(?:#[^)]*)?\)", row):
                routes.add((index.parent / target).resolve())
    for guide in guides:
        if guide.name != "README.md" and guide.resolve() not in routes:
            raise ValueError(f"Capability selection table does not link to {guide.name}")


def check():
    codex_dir = ROOT / ".codex"
    config = tomllib.loads((codex_dir / "config.toml").read_text())
    mise = tomllib.loads((ROOT / "mise.toml").read_text())
    if "agents" in config or list((codex_dir / "agents").glob("*.toml")):
        raise ValueError("Subagent registrations must be removed")
    capabilities = ROOT / "docs/development/workflow/capabilities"
    for name in ("planning", "architect", "explorer", "implementer", "reviewer", "tester", "documentation"):
        if not (capabilities / (name + ".md")).is_file():
            raise ValueError(f"Missing capability guide: {name}")
    guides = list(capabilities.glob("*.md"))
    check_capability_routes(ROOT / "README.md", guides)
    for guide in guides:
        for reference in re.findall(r"docs/[A-Za-z0-9_./-]+\.md", guide.read_text()):
            if not (ROOT / reference).is_file():
                raise ValueError(f"Missing capability reference: {reference}")
    document_count = check_documentation_links(ROOT)
    check_cloud_build_sources(ROOT)
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
    print(f"PASS: single-agent capabilities, mise env/tool config, {document_count} documents, {len(json_paths)} JSON files and {len(scripts)} shell scripts")


if __name__ == "__main__":
    try:
        check()
    except (OSError, ValueError, KeyError, TypeError, subprocess.CalledProcessError) as error:
        raise SystemExit(f"Configuration check failed: {error}") from None
