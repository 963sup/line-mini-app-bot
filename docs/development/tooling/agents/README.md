# MCP tool selection

依任務選服務。角色對照是閱讀入口，實際連線設定以[專案 MCP 設定](../../../../.codex/config.toml)為準。原生工具缺少時先查[專案載入](#project-loading-check)；修改指引見[本層規則](AGENTS.md)。

## 開發角色共同工作指引

接到任務時，先讀[根目錄指引](../../../../AGENTS.md)及適用的巢狀指引，遵循查證來源與任務範圍規則。依[開發流程](../../workflow.md)交接並重用已驗證的證據；依[驗收指引](../../acceptance.md)選擇檢查。業務變更另讀[業務操作契約](../../../contracts/business-operations.md)並遵循其符合性流程。透過 `scripts/tooling/tool-exec.sh` 執行專案工具，回報實際檢查結果及未驗證事項。

依角色與任務選讀下表。涉及 LINE 或 Firestore 時，另查[LINE 能力](../../../integrations/line/capabilities.md)與[Firestore 連線](../../../integrations/firestore/connection.md)；其他領域由[文件總覽](../../../README.md)定位。

| 角色 | 任務相關文件 |
| --- | --- |
| architect | [line-平台架構總覽](../../../architecture/runtime/overview.md) · [共用系統基礎](../../../architecture/README.md) · [graphify](graphify.md) |
| implementer | [Go 開發與驗收](../../../standards/go.md) · [驗收索引（原路徑）](../../acceptance.md) · [openaiDeveloperDocs](openai-docs.md) · [context7](context7.md) |
| explorer | [line-平台架構總覽](../../../architecture/runtime/overview.md) · [graphify](graphify.md) · [gopls](../backend/gopls.md) · [repomix](repomix.md) |
| orchestrator | [開發流程索引（原路徑）](../../workflow.md) · [驗收索引（原路徑）](../../acceptance.md) · [github](github.md) |
| explorer（技術研究） | [共用系統基礎](../../../architecture/README.md) · [openaiDeveloperDocs](openai-docs.md) · [context7](context7.md) · [repomix_skills](repomix-skills.md) |
| reviewer | [驗收索引（原路徑）](../../acceptance.md) · [line-平台架構總覽](../../../architecture/runtime/overview.md) · [repomix](repomix.md) · [github](github.md) |
| tester | [驗收索引（原路徑）](../../acceptance.md) · [gopls](../backend/gopls.md) |

## MCP 工具與角色

| Service | Primary roles |
| --- | --- |
| [graphify](graphify.md) | Explorer；Architect 亦可使用 |
| [openaiDeveloperDocs](openai-docs.md) | Explorer, Implementer |
| [context7](context7.md) | Explorer、Implementer；版本相符文件查證 |
| [shadcn](../frontend/shadcn.md) | Implementer；MINI App 元件 registry 與整合輔助 |
| [next-devtools](../frontend/nextjs-devtools.md) | Implementer；按需檢查本機開發伺服器的 runtime 狀態 |
| [repomix](repomix.md) | Explorer；Reviewer 亦可使用 |
| [mermaid](../shared/mermaid.md) | Documentation；架構、資料關係、操作流程與狀態圖，來源與渲染分開驗證 |
| [repomix_skills](repomix-skills.md) | 交接主責代理；接手者需重用一組原始碼時自動產生參考包 |
| [gopls](../backend/gopls.md) | Implementer；Explorer、Tester、Reviewer 亦可使用 |
| [github](github.md) | Explorer、Documentation、Orchestrator、Reviewer |
| [google-developer-knowledge、gcloud](google.md) | Explorer、Implementer、Reviewer、Orchestrator；查 Google 文件、核對 CLI 操作 |
| [cloud-run、toolbox、gke](../../../../.codex/config.toml) | 依任務選 Cloud Run、Firestore 或 GKE。啟用狀態與允許的工具依左側設定，當前可用性依原生呼叫驗證。 |

Roles share services; this mapping is guidance, not access control. Host tool
exposure, server allowlists and user authorization determine actual access.
Use OpenAI Docs first for Codex questions, Google Developer Knowledge and installed
Google skills for Google Cloud, and Context7 for other library/version questions.
Use Graphify/gopls for structure and verify source claims; Repomix bundles sources.
TypeScript and Next.js use Context7 and version-matched official documentation;
Next.js runtime inspection may additionally use the configured next-devtools
connector. There is no dedicated TypeScript language-service MCP server.
Role selection and writer ownership follow the [agent workflow](../../workflow.md#delegation-depth-and-role-selection).

## MCP 診斷

1. 從儲存庫根目錄執行 `python3 scripts/check/check-config.py` 核對設定。
2. 診斷單一服務時，將 `<name>` 換成專案設定的 server 名稱，執行 `bash scripts/tooling/tool-exec.sh python3 scripts/mcp/mcp-doctor.py --server <name>`。盤點所有服務時執行 `bash scripts/tooling/tool-exec.sh mise run doctor-mcp`。Doctor 驗證握手與允許的工具清單；選用服務缺少憑證時，記錄該項結果，再檢查其餘服務。
3. 在實際使用的 Codex 工作階段依[主機載入檢查](#project-loading-check)確認工具可見，再呼叫一個符合任務的工具。分開記錄設定、安裝、握手、主機載入與操作結果。

## Host environment inheritance

Local stdio servers explicitly forward proxy and CA environment variable names
through `env_vars`. gopls additionally forwards Go module routing/privacy settings.
Values remain runtime-owned; mise still supplies workspace tool/cache paths. This
helps preserve the host network environment without embedding credentials.
Direct doctor processes inherit their caller environment and therefore do not
prove that Codex's filtered stdio launch environment or project trust is correct.
Inspect `codex mcp list` and test a host-exposed tool separately.

## Task-scoped tool use

Check the current host's exposed tools once for the relevant service. Reuse the
verified inventory within the same host/session, server configuration and
allowlist; refresh when those change, a required tool is missing, or a call fails.
Do not run doctor or tools/list before every operation or independently in every
specialist. A protocol client's initial initialize/tools-list handshake is distinct
from unnecessary repeated discovery during a healthy connection.

For a new Cloud environment or an explicit inventory audit, compare every
server in the checkout configuration with `codex mcp list` and the current
chat's native tools. Record disabled servers separately. Reuse this inventory
across roles. For routine work, check only the relevant servers.

1. For a missing or failing tool, verify checkout, project trust, server
   enablement and allowlist. Prefer a native tool that matches the intended
   repository and revision. A host plugin may use a different source or index.
2. Prepare missing pinned dependencies using the existing installation task
   through `tool-exec.sh`. Run a single-server doctor when needed to identify
   startup, transport or credential failures. See the server-specific guide
   for preparation commands.
3. Use supported host rediscovery, or verify in a new session opened in the
   intended checkout. Confirm each enabled server with a native read-only
   call where its allowlist permits one. If the allowlist only exposes a
   write operation, record the native tool as exposed and functional acceptance
   as pending until an authorized operation succeeds. For example, Repomix Skills
   exposes only `generate_skill`. Record missing tools, credentials
   or permissions and continue independent work. Direct transport and CLI
   checks serve as diagnostics and do not substitute for the required MCP call.

Use only task-relevant services and bounded queries. Reuse a healthy connection
when the client supports it; keep related search/fetch or bundle/read operations
within that connection rather than reconnecting for each tool. Run a single-server
doctor for a diagnosed problem, not the full optional service suite for routine
work. On failure identify binary, arguments, transport, network or source mismatch
before retrying; do not repeatedly increase timeouts or launch unrelated servers.

The doctor bounds stdio waits and isolates each HTTP exchange in a worker that
is killed and reaped on timeout or interruption. Its existing server timeout
configuration owns the budget; this is not a hard real-time guarantee for OS
process creation. Offline regressions run in the configuration CI job; see the
[doctor regression tests](../../../../scripts/test/test_mcp_doctor.py).

Local Codex sessions may disable an unrelated server without deleting its
configuration using `-c 'mcp_servers.SERVER.enabled=false'`. `enabled_tools` and
`disabled_tools` filter tools in the parent configuration; a custom role inherits
that server set and cannot create a separate MCP sandbox. Host/plugin connections
have their own controls, so these local overrides do not reconfigure this chat.
Keep existing service capabilities unless task-specific overrides are chosen.

Configuration options and these boundaries were checked through OpenAI Docs MCP
against the [official MCP guide](https://developers.openai.com/codex/mcp) and
[subagent guide](https://learn.chatgpt.com/docs/agent-configuration/subagents) on
2026-11-14 (Asia/Taipei). Independent read-heavy subproblems can use agents; their
results should be concise evidence summaries, not raw intermediate output.

### Verification snapshot

On 2026-11-14 (Asia/Taipei), a read-only Context7 registration check used one
focused local Graphify CLI query (0.813 s), then one Repomix MCP connection:
initialize 0.384 s, tools/list 0.010 s, and two file reads 0.018 s / 0.004 s.
The measured corrected run had two MCP tool calls, one CLI query and no failures;
an earlier absolute-path call was rejected by Repomix's sandbox and corrected
to repository-relative paths. These are observed operation times, not model/agent
latency or a before/after performance claim.

The source results were summarized as the workflow's six-field handoff with
revision, relevant dirty paths and source locations. This verifies direct tools
and evidence transfer; the host used for that snapshot had no `agent_type` selection parameter,
so native execution of all registered custom roles remains unverified. Configuration,
direct MCP operation and host-exposed tools must still be reported separately.

References: [OpenAI MCP configuration](https://developers.openai.com/codex/mcp)
and Context7 `/openai/codex` stdio environment configuration.

[上層導覽](../README.md)

## Project loading check

Open the actual Git checkout in Codex and review the host's project trust decision. From that checkout run `python3 scripts/mcp/check-host-config.py`. This compares every configured server, including disabled entries, with the effective CLI inventory. Keep project endpoints in .codex/config.toml. Installing tools does not require registering copies in the user configuration.

If CLI configuration matches, check tool discovery in the target chat and verify the required native tool with a read-only operation. A managed host may expose a separate catalog. Report that boundary when host reload or discovery is unavailable.
