#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/dev-env.sh"
cd "$(dirname "${BASH_SOURCE[0]}")/../.."
# Name tools explicitly so mise does not install every missing project tool.
tools=()
while [[ ${1:-} == --tool ]]; do
  [[ $# -ge 3 ]] || { echo "Usage: tool-exec.sh [--tool TOOL] COMMAND [ARGS...]" >&2; exit 2; }
  tools+=("$2")
  shift 2
done
[[ $# -gt 0 ]] || { echo "A command is required" >&2; exit 2; }
if [[ ${#tools[@]} -eq 0 ]]; then
  tool="$(python3 - "$1" <<'PYTHON'
import sys, tomllib
from pathlib import Path
command = sys.argv[1]
aliases = {'python3': 'python', 'npm': 'node', 'npx': 'node'}
command = aliases.get(command, command)
for tool in tomllib.loads(Path('mise.toml').read_text())['tools']:
    if tool.split(':')[-1].split('/')[-1] == command:
        print(tool)
        break
PYTHON
)"
  [[ -z $tool ]] || tools+=("$tool")
fi
if [[ ${#tools[@]} -gt 0 ]]; then
  python3 - "${tools[@]}" <<'PYTHON'
import sys, tomllib
from pathlib import Path
configured = tomllib.loads(Path('mise.toml').read_text())['tools']
for tool in sys.argv[1:]:
    if tool not in configured:
        raise SystemExit(f"Tool must be declared in mise.toml without a version override: {tool}")
PYTHON
  exec mise exec --locked "${tools[@]}" -- "$@"
fi
# System commands need the configured PATH, without an unrelated installation.
export MISE_EXEC_AUTO_INSTALL=false
exec mise exec -- "$@"
