#!/usr/bin/env bash
set -euo pipefail

cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.."
exec npx --yes skills add yamadashy/repomix --skill repomix-explorer --agent codex --yes
