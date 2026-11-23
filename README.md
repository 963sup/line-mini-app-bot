# LINE MINI App Bot

A Go and Next.js monorepo for LINE Messaging API webhooks, MINI App APIs and Firestore-backed business features. A single Go service serves HTTP APIs and the statically exported frontend on Google Cloud Run.

Optional [LINE integration](docs/integrations/line/setup.md) provides direct and group commands, Flex messages, postbacks, MINI App authentication and automatic Rich Menu synchronization. The repository also includes a read-only Firestore connectivity check. Deployment, external-service integration and device acceptance require separate evidence; source code and passing local checks do not establish production readiness.

Current capabilities are tracked in the [domain index](docs/domains/README.md#domain-matrix):

- [Attendance](docs/domains/attendance/README.md): clock-in/out, breaks, records, corrections, leave registration, review and group boards. Attendance state is maintained separately for each person and group; location evidence does not enforce a work radius.
- Repository features include issues, discussions, objects and subobjects, templates, labels, milestones, insights and Wiki. Projects and journals have separate entry points in the [feature documentation](docs/domains/README.md).
- Enterprise, organization and team management, chat summaries, weather and PDF text extraction have implementations. The domain index distinguishes implemented capabilities from proposals.

The [Go + Next.js/React geolocation example](docs/integrations/geolocation/browser-api.md) is available at `http://localhost:8080/location`. It sends browser coordinates to Go for validation without storing them.

The [accepted baseline](docs/decisions/0002-baseline.md) records the original attendance milestone and architectural decisions. Current feature rules and acceptance criteria belong to their domain specifications. Follow the [staged verification workflow](docs/development/environment/verification.md) to verify execution and deployment foundations, data and LINE integration, then business behavior.

## Architecture and deployment

Source is organized into `backend/` (Go), `frontend/miniapp/` (Next.js) and `api/` (OpenAPI and the TypeScript client). The backend entry point is `backend/main.go`. Business responsibilities are organized by domain; see the [current code layout](docs/architecture/runtime/code-layout.md), [Monorepo decision](docs/decisions/0004-monorepo.md) and [documentation index](docs/README.md).

During development, the MINI App is available at `http://localhost:3000/miniapp`. Production uses Next.js static export served by Go alongside HTTP APIs. API contracts, generation and compatibility procedures are maintained in the [API workflow](api/README.md).

The GitHub `main` managed trigger builds the image with `Dockerfile` and updates Cloud Run. It does not execute `cloudbuild.yaml`, which is retained for manual submissions or a custom trigger. Runtime environment variables and secret references live in the Cloud Run service configuration; secrets are bound through Secret Manager. See [Cloud Run deployment](docs/operations/cloud-run.md) for deployment methods, initial setup and IAM.

Shared HTTPS endpoint handling for the three MINI App environments is implemented. Desktop Flex entry points and mobile permanent links are described in [LINE environment configuration](docs/integrations/line/configuration.md#3-mini-app).

<a id="技術棧與-sdk"></a>

## Technology stack and SDKs

| Area | Technology and purpose | Configuration and sources |
| --- | --- | --- |
| Backend | Go and the standard library's `net/http` | [Backend guide](backend/README.md), [go.mod](backend/go.mod) |
| Frontend | Next.js, React, TypeScript | [Frontend dependencies](frontend/miniapp/package.json), [static export configuration](frontend/miniapp/next.config.ts) |
| UI | shadcn/ui, Radix UI, Tailwind CSS, Lucide icons | [Frontend guide](frontend/README.md) |
| API contracts | OpenAPI; Go and TypeScript generation with oapi-codegen and openapi-typescript | [API workflow](api/README.md) |
| Storage | Cloud Firestore | [Persistence documentation](docs/architecture/README.md#資料與儲存) |
| Workspace and builds | Node.js, pnpm workspaces, Turborepo; mise-managed toolchain | [Workspace configuration](package.json), [toolchain configuration](mise.toml) |
| Deployment | Docker multi-stage builds, Cloud Build, Cloud Run, Secret Manager | [Dockerfile](Dockerfile), [deployment guide](docs/operations/cloud-run.md) |
| Quality checks | GitHub Actions, Go tests and analysis tools, Node Test Runner/tsx/jsdom | [CI configuration](.github/workflows/check.yml), [toolchain configuration](mise.toml), [frontend test commands](frontend/miniapp/package.json) |

Primary SDKs and packages are listed below. Complete dependencies and pinned versions are defined in [backend/go.mod](backend/go.mod), the [frontend package.json](frontend/miniapp/package.json) and the [API client package.json](api/client/package.json).

| SDK or package | Purpose |
| --- | --- |
| `github.com/line/line-bot-sdk-go/v8` | LINE Messaging API and webhook integration |
| `cloud.google.com/go/firestore` | Firestore access |
| `github.com/openai/openai-go/v3` | OpenAI integration through the official Go SDK |
| `google.golang.org/api`, `golang.org/x/oauth2` | Google API and authentication support |
| `github.com/getkin/kin-openapi`, `github.com/oapi-codegen/nethttp-middleware`, `github.com/oapi-codegen/runtime` | OpenAPI contract handling, request validation and generated Go code support |
| `openapi-fetch` | TypeScript API client |

See [dependency and contract ownership](docs/architecture/runtime/dependencies.md) for technical baselines and maintenance responsibilities. This inventory does not establish end-to-end acceptance of external services or production environments.

<a id="開始開發"></a>

## Start developing

Run commands from the repository root:

```bash
bash scripts/bootstrap.sh core
bash scripts/tooling/tool-exec.sh pnpm build
bash scripts/tooling/tool-exec.sh pnpm check
bash scripts/tooling/tool-exec.sh pnpm test
bash scripts/tooling/tool-exec.sh pnpm dev
```

For Cloud and MCP installation details and limitations, see the [toolchain setup guide](docs/development/tooling/shared/toolchain-setup.md). Set the Codex Cloud install script to `scripts/bootstrap.sh`; without arguments, it uses cloud mode. The local commands above use core mode to install Go, web dependencies and the Mermaid CLI. Use `bash scripts/bootstrap.sh all` when Graphify, Repomix and MCP tools are needed. Firestore Toolbox requires Google ADC; GKE MCP is disabled by default.

Verify tool installation, MCP handshakes and availability to the chat host separately. Cached artifacts do not establish that a fresh environment has been verified. Tool versions and non-secret environment settings are maintained in `mise.toml`; use `scripts/tooling/tool-exec.sh` for routine commands.

- [Documentation and Backend/API/Frontend/Scripts/Deploy entry points](docs/README.md#按模組閱讀)
- [Locate code from requirements](docs/architecture/runtime/code-layout.md#從需求定位修改範圍)
- [Change acceptance matrix](docs/development/tooling/testing/commands.md#變更驗收矩陣)
- [Architecture and flows](docs/architecture/runtime/overview.md)
- [Go development](docs/standards/go.md)
- [Contributing](CONTRIBUTING.md)
- [Report security issues](.github/SECURITY.md)

Choose checks according to the change's impact using the acceptance matrix. Follow the [engineering and delivery requirements](AGENTS.md#engineering-decisions-and-delivery-requirements) for regression coverage. Do not commit LINE tokens, channel secrets or cloud credentials.

<a id="儲存庫分析"></a>

## Repository analysis

Run the existing CLI tasks from the repository root:

```bash
bash scripts/tooling/tool-exec.sh mise run graphify-update
bash scripts/tooling/tool-exec.sh graphify query "registerLINEWebhook"
bash scripts/tooling/tool-exec.sh mise run repomix
```

Graphify writes `graphify-out/graph.json`; the current task performs AST extraction only. Parsing of documents, TOML, HTML/CSS and lockfiles is not guaranteed. Use Repomix's `repomix-output.xml` to read document and configuration text, then check the actual sources. Successful packaging or queries do not establish code correctness or integration acceptance.

Both tools exclude upstream agent installations under `.agents`, `.claude` and `agent`, as well as `skills-lock.json`. Repomix explicitly permits the maintained `.agents/AGENTS.md` and `.agents/README.md` entry points. `.codex` remains an analysis input, subject to each tool's supported formats and built-in exclusions. Ignore rules do not guarantee complete inclusion.

<a id="執行服務"></a>

## Run the service

```bash
bash scripts/tooling/tool-exec.sh mise run dev
# The LINE webhook is not registered without LINE credentials.
# See the LINE integration guide for connection setup.
```

The service reads `PORT` (default 8080), Messaging API credentials and MINI App environment settings; it does not automatically load `.env`. Configuration names are listed in [.env.example](.env.example). The default MINI App environment is `developing`; `review` and `published` are also supported.

The MINI App entry point is `/miniapp`. The selected environment supplies the public Channel ID, LIFF ID and endpoint; missing required values prevent startup. Login also requires session database configuration. See the [LINE integration guide](docs/integrations/line/setup.md) for operating instructions and verification limitations.

No license has been selected. All rights are reserved under [LICENSE](LICENSE).

<a id="巢狀目錄導覽"></a>

## Nested directory guide

Select the target directory below. Read AGENTS.md files along the path from the repository root to the target, and use the nearest README.md for implementation entry points. For cross-module questions, start with [documentation by task](docs/README.md#按任務開始). Follow the [nested documentation maintenance standards](AGENTS.md#nested-documentation-maintenance-standards) when adding or editing guidance.

| Directory or file | Description | Agent guidance |
| --- | --- | --- |
| `.agents/` | [Skill maintenance](docs/development/tooling/agents/skills.md) | [Root guidance](AGENTS.md) |
| `.codex/` | [Codex configuration](.codex/config.toml) · [CLI operations](docs/development/tooling/agents/codex.md) | [Root guidance](AGENTS.md) |
| `.github/` | [GitHub operations](docs/development/tooling/shared/github.md) · [CI configuration](.github/workflows/check.yml) | [Root guidance](AGENTS.md) |
| `.vscode/` | [Workspace tasks](.vscode/tasks.json) · [Editor settings](.vscode/settings.json) | [Root guidance](AGENTS.md) |
| `api/` | [HTTP API contracts](api/README.md) | [AGENTS.md](api/AGENTS.md) |
| `assets/` | [Application assets](assets/README.md) | [AGENTS.md](assets/AGENTS.md) |
| `backend/` | [Go backend](backend/README.md) | [AGENTS.md](backend/AGENTS.md) |
| `cloudbuild.yaml`, `firestore.indexes.json` | [Deployment and index configuration](docs/operations/README.md) | [Operations guidance](docs/operations/AGENTS.md) |
| `docs/` | [Project documentation](docs/README.md) | [AGENTS.md](docs/AGENTS.md) |
| `scripts/` | [Script tooling](scripts/README.md) | [AGENTS.md](scripts/AGENTS.md) |
| `frontend/` | [MINI App frontend](frontend/README.md) | [AGENTS.md](frontend/AGENTS.md) |

When entering `frontend/miniapp/`, also read its [Next.js guidance](frontend/miniapp/AGENTS.md).

<a id="業務操作開發原則"></a>

## Business operation development principles

Use the [resource-oriented business operation model](docs/contracts/business-operations.md): Actor, Resource, Action, Context, Authority, Policy, Decision and Result. Define operation contracts before development, and compare implementation with evidence at delivery. Reads do not require event creation; use Merge only when adopting a proposal.

When inconsistencies are found, follow the [Graphify/Repomix/Context7 compliance workflow](docs/development/workflow/implementation.md#business-operation-compliance) and fix them within task scope. Standardize the operation model and isolate business rules; share necessary mechanisms while avoiding unnecessary abstractions.
