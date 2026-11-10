# Acceptance and evidence

Current phase: stabilize the foundation; do not add test plans, cases or tools.
The test-related matrix entries below are retained for the later testing phase.
Keep existing CI; this phase note does not authorize removing checks or claiming
unrun checks passed. Scope follows the [accepted baseline](../architecture/decisions/0002-baseline.md).



| Change | Required evidence |
| --- | --- |
| Configuration/scripts | `python3 scripts/tooling/check-config.py`; relevant task smoke check |
| Go behavior | format check, focused package tests, then shared `mise run check` and `mise run build` before delivery |
| Concurrency | race tests exercising the changed concurrent behavior |
| Dependencies | inspect go.mod/go.sum; download/verify modules; vulnerability review as appropriate |
| LINE authentication/webhook | invalid signature/token, unauthorized user and duplicate event cases |

Run tooling through `bash scripts/tooling/tool-exec.sh <command>`. Formatting can be
checked with `gofmt -l` on changed Go files; formatting changes require file
ownership. No go.mod means application verification is unavailable, not passed.
Run complete race suites once after integration instead of concurrently per role.

## Evidence limits

Graphify AST extraction can omit TOML, documentation and dynamic calls. Repomix
is a source bundle, not a correctness proof. Configuration parsing, MCP handshake,
host tool exposure and representative tool operations are separate checks.
Do not claim a fresh Cloud environment works based on an existing tool cache.

## Cloud Run and browser location

Cloud Run changes: verify image contents, HTTP routing, service account, IAM, secrets,
revision readiness and shutdown behavior; config parsing is not a deployment test.
Geolocation changes: permission denied, unavailable, timeout, unsupported/insecure context,
coordinate validation, missing fields, cancellation/watch cleanup and authorization.
Report LINE iOS/Android and external-browser acceptance separately; mocks do not prove device support.
