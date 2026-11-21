"""Refresh Codex's development reference skill before each submitted prompt."""

import fcntl
import json
import os
from pathlib import Path
import subprocess
from tempfile import TemporaryDirectory


ROOT = Path(__file__).resolve().parents[2]
SKILL_NAME = "repomix-reference-development-entry"
SOURCES = (
    "AGENTS.md",
    "docs/development/README.md",
    "docs/development/environment/README.md",
    "docs/development/workflow/README.md",
    "docs/development/recipes/README.md",
    "docs/development/verification/README.md",
    "docs/development/tooling/README.md",
)
REFERENCES = (
    "references/summary.md",
    "references/project-structure.md",
    "references/files.md",
    "SKILL.md",
)


def refresh() -> Path:
    skills = ROOT / ".agents" / "skills"
    skill = skills / SKILL_NAME
    for directory in (skills.parent, skills, skill, skill / "references"):
        if directory.is_symlink():
            raise ValueError("the skill destination contains a symbolic link")
        directory.mkdir(exist_ok=True)

    # Serialize writes for chats sharing a checkout. Stage outside the skills
    # directory so a failed generation cannot replace the previous reference.
    lock = skills.parent / f".{SKILL_NAME}.lock"
    with os.fdopen(os.open(lock, os.O_CREAT | os.O_WRONLY | os.O_NOFOLLOW, 0o600), "w") as handle:
        fcntl.flock(handle, fcntl.LOCK_EX)
        with TemporaryDirectory(prefix=".repomix-skill-", dir=skills.parent) as directory:
            draft = Path(directory) / SKILL_NAME
            subprocess.run(
                [
                    "bash", "scripts/tooling/tool-exec.sh", "repomix", ".",
                    "--skill-generate", SKILL_NAME,
                    "--skill-output", str(draft),
                    "--skill-project-name", "Line Mini App Bot development handbook",
                    "--include", ",".join(SOURCES),
                    "--force", "--quiet",
                ],
                cwd=ROOT, stdin=subprocess.DEVNULL, stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL, timeout=50, check=True,
            )
            if not all((draft / name).is_file() for name in REFERENCES):
                raise ValueError("Repomix did not produce the required skill files")
            for name in REFERENCES:
                (draft / name).replace(skill / name)
    return skill / "SKILL.md"


def main() -> None:
    try:
        skill_file = refresh()
    except subprocess.TimeoutExpired:
        result = {"systemMessage": "Repomix skill refresh timed out. The original sources remain available."}
    except subprocess.CalledProcessError as error:
        result = {"systemMessage": f"Repomix skill refresh failed (exit {error.returncode}). Check the pinned CLI via tool-exec.sh."}
    except (OSError, ValueError) as error:
        result = {"systemMessage": f"Repomix skill refresh skipped: {error}."}
    else:
        result = {
            "hookSpecificOutput": {
                "hookEventName": "UserPromptSubmit",
                "additionalContext": (
                    f"Repomix refreshed the development reference for this turn. "
                    f"Read {skill_file} and follow its links only as needed for the task."
                ),
            },
        }
    print(json.dumps(result, ensure_ascii=False))


if __name__ == "__main__":
    main()
