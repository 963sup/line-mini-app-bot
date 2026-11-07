# Acceptance and evidence



| Change | Required evidence |
| --- | --- |
| Configuration/scripts | `python3 scripts/check-config.py`; relevant task smoke check |
| Go behavior | format check, focused package tests, then shared `mise run check` and `mise run build` before delivery |
| Concurrency | race tests exercising the changed concurrent behavior |
| Dependencies | inspect go.mod/go.sum; download/verify modules; vulnerability review as appropriate |
| Firestore behavior | emulator integration tests; report emulator absence explicitly |
| LINE authentication/webhook | invalid signature/token, unauthorized user and duplicate event cases |

Run tooling through `bash scripts/tool-exec.sh <command>`. Formatting can be
checked with `gofmt -l` on changed Go files; formatting changes require file
ownership. No go.mod means application verification is unavailable, not passed.
Run complete race suites once after integration instead of concurrently per role.

## Evidence limits

Graphify AST extraction can omit TOML, documentation and dynamic calls. Repomix
is a source bundle, not a correctness proof. Configuration parsing, MCP handshake,
host tool exposure and representative tool operations are separate checks.
Do not claim a fresh Cloud environment works based on an existing tool cache.
