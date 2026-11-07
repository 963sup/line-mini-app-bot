# LINE / Go / Cloud Run

Source scripts/cloud-env.sh in every shell, or use scripts/tool-exec.sh.
Run bash scripts/bootstrap.sh for Go setup, then mise run doctor.
Install optional tooling using mise run install-cloud or install-mcp.

SDK dependencies belong in the future application's go.mod:
- github.com/line/line-bot-sdk-go/v8
- cloud.google.com/go/firestore
net/http, html/template and log/slog require no extra packages.

The repository has no application module yet. mise run dev expects cmd/server
and defaults PORT to 8080. Cloud Run requires listening on :PORT.
mise run container-build requires an application Dockerfile and a Docker daemon.

Use the pinned gcloud with: mise exec gcloud@592.0.0 -- gcloud ...
Install emulator component with: gcloud components install cloud-firestore-emulator --quiet
Then run mise run firestore-emulator. Only the local application process should set
FIRESTORE_EMULATOR_HOST=127.0.0.1:8085 and a local Google Cloud project ID.
Never carry the emulator variable into a production deployment.

Cloud Run, Artifact Registry and Secret Manager require an actual project, region
and authorized identity. Prefer Cloud Run service identity / ADC, not downloaded
service-account keys. No deployment or remote writes are part of environment setup.
Supply LINE secrets securely at runtime; local HMAC signature verification needs the
actual channel secret, not a proxy-only placeholder. Verify webhook signatures on
the original request bytes, and validate MINI App ID tokens server-side for the
expected channel. Avoid logging tokens, secret values or whole webhook payloads.

## Bootstrap and editor

The base image must provide Bash, curl, tar, checksum utilities and Python 3.11+
before mise starts. Python parses the single-source environment table in
mise.toml. The managed Python pin cannot satisfy this earlier bootstrap step.
Run bash scripts/bootstrap.sh core (also the default). Remote MCP dependencies
and credentials are optional: run bash scripts/bootstrap.sh all explicitly.
Use mise run check-config for offline repository configuration validation.

VS Code recommends golang.go. Its Go and gopls paths use repository wrappers so
the same mise tools and writable paths are used. Run Go: setup before opening
Go files. .vscode configuration is intended for this Cloud /workspace checkout;
it is not a claim of portability to arbitrary host paths.

GitHub Actions validates repository configuration on every push/PR and runs Go
checks only when go.mod exists. It reads the Go version from mise.toml but does
not apply Cloud-specific /workspace environment variables to hosted runners.
Until an application module exists, CI explicitly reports application checks
NOT RUN. Local doctor tests the toolchain, not application behavior.

## Analysis and Codex

mise run repomix writes an ignored repomix-output.xml with security checking.
Generated skills, graph data and common secret files are excluded to keep the
pack focused. This scanner is not a guarantee that arbitrary secrets are absent.
mise run graphify-update writes graphify-out/graph.json in the project. AGENTS.md
and the Graphify MCP configuration refer to that same file. The graph is generated
and ignored, so rebuild it after restoring a clean checkout.

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
