"""Export the project's non-secret mise environment before mise itself starts."""
import os
from pathlib import Path
import re
import shlex
import tomllib

root = Path(__file__).resolve().parent.parent
settings = tomllib.loads((root / "mise.toml").read_text())["env"]
for name, value in settings.items():
    if not isinstance(value, str):
        raise SystemExit(f"Unsupported bootstrap env value: {name}")
    match = re.fullmatch(r'{{ env\.(\w+) \| default\(value="([^"]*)"\) }}', value)
    if match:
        value = os.environ.get(match[1]) or match[2]
    elif "{{" in value:
        raise SystemExit(f"Unsupported bootstrap template: {name}")
    print(f"export {name}={shlex.quote(value)}")
