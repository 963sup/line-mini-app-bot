# Cloud development environment

首版僅做 [基礎出勤](../features/attendance.md)：上班、下班與本人紀錄。
依 [分階段落地](../development/foundation.md) 先穩定執行／部署基礎，再接入資料與 LINE，
最後開發出勤；具體規則集中在功能規格確認。架構與階段範圍依
[已接受的基線](../architecture/decisions/0002-baseline.md)：目前先穩定基礎，
不新增測試方案、案例或工具，保留現有 CI。

Source scripts/tooling/dev-env.sh in every shell, or use scripts/tooling/tool-exec.sh.
Run bash scripts/tooling/bootstrap.sh for Go setup, then mise run doctor.
Install optional tooling using mise run install-cloud or install-mcp.

SDK dependencies are managed in the existing application's go.mod/go.sum:
- github.com/line/line-bot-sdk-go/v8
- cloud.google.com/go/firestore
net/http, html/template and log/slog require no extra packages.

The repository contains go.mod, cmd/server and a Dockerfile. The HTTP skeleton
currently provides an optional LINE webhook connectivity prototype, without health
routes. Attendance and Firestore behavior is not implemented. See the
[prototype guide](../development/line-prototype.md). mise run dev defaults PORT to 8080.
Cloud Run injects `PORT`; mise run container-build also needs a Docker daemon.

Use the pinned gcloud with: mise exec gcloud@592.0.0 -- gcloud ...

Cloud Run is the application deployment target and uses a runtime service
account plus Secret Manager bindings.
Firestore access requires a Google project, explicit authentication and appropriate
IAM. See the [Cloud Run deployment guide](../operations/cloud-run.md).
No deployment or remote writes
are part of environment setup.
Supply LINE secrets securely at runtime; local HMAC signature verification needs the
actual channel secret, not a proxy-only placeholder. Verify webhook signatures on
the original request bytes, and validate MINI App ID tokens server-side for the
expected channel. Avoid logging tokens, secret values or whole webhook payloads.

## Bootstrap and editor

The base image must provide Bash, curl, tar, checksum utilities and Python 3.11+
before mise starts. Python parses the single-source environment table in
mise.toml. The managed Python pin cannot satisfy this earlier bootstrap step.
Run bash scripts/tooling/bootstrap.sh core (also the default). Remote MCP dependencies
and credentials are optional: run bash scripts/tooling/bootstrap.sh all explicitly.
Use mise run check-config for offline repository configuration validation.

VS Code recommends golang.go. Its Go and gopls paths use repository wrappers so
the same mise tools and writable paths are used. Run Go: setup before opening
Go files. .vscode configuration is intended for this Cloud /workspace checkout;
it is not a claim of portability to arbitrary host paths.

GitHub Actions validates repository configuration on every push/PR and runs Go
checks only when go.mod exists. It reads the Go version from mise.toml but does
not apply Cloud-specific /workspace environment variables to hosted runners.
The module now exists, so CI selects its Go checks. This describes workflow
configuration, not evidence that a particular run passed. Local doctor checks
the toolchain, not application behavior.

## Analysis and Codex

mise run repomix writes an ignored repomix-output.xml with security checking.
Generated skills, graph data and common secret files are excluded to keep the
pack focused. This scanner is not a guarantee that arbitrary secrets are absent.
mise run graphify-update writes graphify-out/graph.json in the project. AGENTS.md
and the Graphify MCP configuration refer to that same file. The graph is generated
and ignored, so rebuild it after restoring a clean checkout.

mise run graphify-docs builds a separate heading/link graph of maintained project
documentation, including README and scripts guides, outside the source tree.
Use graphify query/path with --graph /workspace/tooling/cache/graphify-docs/graph.json
to follow related documents. Corroborate consistency against source text; this is
not semantic extraction. After correcting documents, rebuild and query again.

The eight standalone agent files include name, description and instructions.
Project .codex configuration requires the host to trust and load the project.
Transport probes demonstrate MCP connectivity only; they do not prove the current
chat host loaded the servers or skill registrations. The two Graphify skill
locations target different consumers; their coexistence is not proof of unique
skill discovery in every host. Keep explicit platform selection if the host
reports duplicate skills. No automatic user-profile trust or permission changes
are performed by bootstrap. Instruction-level no-recursion rules remain in
AGENTS.md; runtime max_depth enforcement has not been independently verified.

Sources consulted through actual read-only Context7 and OpenAI Docs MCP calls:
- https://developers.openai.com/codex/config-basic
- https://developers.openai.com/codex/config-reference
- https://developers.openai.com/codex/multi-agent
- https://github.com/jdx/mise/blob/main/docs/continuous-integration.md
- https://github.com/microsoft/vscode-docs/blob/main/docs/languages/go.md
