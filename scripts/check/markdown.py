"""Run Markdown tools on the same set of repository-owned files."""
import argparse
from pathlib import Path
import re
import subprocess

ROOT = Path(__file__).resolve().parents[2]
EXCLUDED = (
    ".agents/",
    ".claude/",
    "docs/development/tooling/frontend/shadcn-reference/snapshot/",
    "docs/references/ddd/",
)
LICENSES = {
    "docs/development/tooling/frontend/shadcn-reference/LICENSE.md",
    "frontend/miniapp/src/shared/ui/LICENSE.md",
}


CONVENTIONAL_NAMES = {"AGENTS.md", "README.md", "CHANGELOG.md", "CONTRIBUTING.md", "SECURITY.md"}
GITHUB_TEMPLATES = {
    ".github/ISSUE_TEMPLATE/bug_report.md",
    ".github/ISSUE_TEMPLATE/feature_request.md",
    ".github/pull_request_template.md",
}
SNAKE_CASE = re.compile(r"[a-z0-9]+(?:_[a-z0-9]+)*")
ADR = re.compile(r"[0-9]{4}_[a-z0-9]+(?:_[a-z0-9]+)*")


def valid_name(name):
    parts = Path(name).parts
    if parts[0] == "docs" and not all(SNAKE_CASE.fullmatch(part) for part in parts[1:-1]):
        return False
    if parts[-1] in CONVENTIONAL_NAMES or name in GITHUB_TEMPLATES:
        return True
    pattern = ADR if parts[:2] == ("docs", "decisions") else SNAKE_CASE
    return pattern.fullmatch(Path(name).stem) is not None


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("mode", choices=("write", "check", "lint"))
    mode = parser.parse_args().mode
    names = subprocess.check_output(
        ["git", "ls-files", "-z", "--cached", "--others", "--exclude-standard", "--", "*.md"],
        cwd=ROOT,
    ).decode().split("\0")
    paths = [
        name for name in sorted(set(names))
        if name and not name.startswith(EXCLUDED) and name not in LICENSES
        and (ROOT / name).is_file() and not (ROOT / name).is_symlink()
    ]
    if not paths:
        parser.error("No owned Markdown files found.")
    print(f"Markdown {mode}: {len(paths)} files", flush=True)
    if mode == "lint":
        invalid = [name for name in paths if not valid_name(name)]
        if invalid:
            parser.exit(1, "".join(f"{name}: invalid documentation path\n" for name in invalid))
        command = ["markdownlint-cli2", "--no-globs"]
    else:
        command = ["prettier", f"--{mode}", "--ignore-path", ".gitignore"]
    return subprocess.call([*command, *paths], cwd=ROOT)


if __name__ == "__main__":
    raise SystemExit(main())
