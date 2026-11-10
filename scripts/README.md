# Repository scripts

Run commands from the repository root. Scripts resolve repository files from
their own location; `tool-exec.sh` also switches to the repository root.

| Directory | Responsibility |
| --- | --- |
| `tooling/` | Bootstrap, environment, configuration checks, editor wrappers, application tasks, profiling, frontend assets |
| `analysis/` | Graphify documentation graph and Graphify/Repomix scope audit |
| `api/` | OpenAPI generation and backward compatibility checks |
| `release/` | Release smoke checks and Cloud Run metadata inspection |

```bash
bash scripts/tooling/bootstrap.sh core
bash scripts/tooling/tool-exec.sh mise run doctor
python3 scripts/tooling/check-config.py
bash scripts/tooling/tool-exec.sh pnpm api:check
bash scripts/tooling/tool-exec.sh mise run analysis-audit
```

Tool versions and task definitions remain in `mise.toml`. Frontend-only script
paths are wired through `frontend/miniapp/package.json`; repository API commands
through root `package.json`. Release scripts do not grant deployment permission.

See [the path migration](../docs/development/path-reorganization.md) for changed
entry points and verification boundaries.
