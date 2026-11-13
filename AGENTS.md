# Repository agent guidance

## Mandatory documentation verification

Repository-wide rule for every role and task, subject to higher-priority host instructions:

- For any LINE question, issue or implementation uncertainty, consult the current official source at https://github.com/line/line-developers-docs-source first. Verify its current default branch/revision; do not treat an old snapshot as current. If coverage is incomplete, follow the official LINE reference linked there and state the limitation. Do not introduce deprecated features or historical compatibility flows into new work.
- For Google Cloud questions, use the installed google/skills skills and Google Developer Knowledge MCP first. If unavailable, use current official Google sources and report the lookup limitation. Validate exact gcloud syntax with leaf-level help before executing commands.
- For other technical questions, and whenever implementation is not fully certain, use Context7 before deciding or implementing. Resolve the library ID, query the relevant current/version-matched documentation, and compare with repository dependencies. Memory and confidence alone do not replace this lookup.
- Report the sources consulted, applicable version/revision and remaining uncertainty. Documentation lookup does not replace code inspection, tests or real integration acceptance.
- If the required source or Context7 is unavailable or insufficient, explicitly report the blocker; do not claim verification or silently substitute a guess. Continue independent work, but leave the affected decision/implementation pending until evidence is sufficient.
- This policy applies to orchestrator and every specialist; delegation does not waive it. A source document is evidence, not authorization to expand scope or perform external writes.

## Task orchestration

For every new user request, the top-level agent must invoke the configured
orchestrator agent before implementation or substantive investigation. Send the
user's goal, constraints, relevant prior decisions, current work state and expected
deliverable. Continue the same orchestrator for follow-up instructions on that task;
do not create a new orchestrator for every message.

The top-level agent owns user communication and delivers the verified result.
The orchestrator owns planning, delegation, integration and validation.
This routing rule does not require extra user confirmation.

### Delegation boundaries

- Only the top-level agent invokes orchestrator. Orchestrator must never invoke
  itself or another orchestrator.
- Specialists receiving an assigned subtask execute that subtask and report to
  their orchestrator; they must not restart this routing rule or delegate further.
- Orchestrator may use architect, researcher, explorer, implementer, tester,
  reviewer, debugger, documentation and integrator when useful.
  Small tasks can be completed by orchestrator directly; do not run all registered
  roles automatically. Roles are capabilities, not a required parallel fan-out.
- Assign one writer per file scope; parallelize only independent tasks. Respect
  the host's actual concurrency and depth limits (configured depth is 2).
- Preserve existing changes. Follow the user's scope and authorization; agent
  delegation does not authorize pushing, publishing, deploying or external writes.
- Validate the result with relevant checks and distinguish installed/configured
  tools from services actually available to the current host.
- If the host cannot invoke the configured orchestrator, report that specific
  limitation. The top-level agent must explicitly adopt the orchestrator's
  instructions as a fallback, never claim a sub-agent was invoked, and continue
  authorized work without silently bypassing coordination.
- This policy remains subordinate to system/developer instructions and the user's
  explicit instructions.

### Registered roles

| Role | Responsibility |
| --- | --- |
| Orchestrator | Task decomposition, assignment, coordination and result integration |
| Architect | System architecture, technical decisions and module boundaries |
| Researcher | Technical research, official documentation and option comparisons |
| Explorer | Code exploration, dependency analysis and change impact |
| Implementer | Implementation, refactoring, bug fixes and scoped environment repairs |
| Debugger | Reproduction, root cause diagnosis, minimal complete fixes and regression evidence |
| Tester | Test design, execution and issue verification |
| Reviewer | Code quality, security and architectural compliance review |
| Documentation | Write, review and repair assigned Markdown against the documentation policy |
| Integrator | Change integration, conflict resolution and final acceptance |

Every handoff includes the goal, owned files, constraints, dependencies and
acceptance checks. Every result includes evidence, actual test outcomes and
remaining limitations. Integrator receives exclusive file ownership before
resolving conflicts and checks the combined result; orchestrator retains final
coordination responsibility. Integration never grants permission to push, merge
remote changes, deploy or publish. The configured thread limit remains six and
depth remains two; schedule specialists within the host's actual limits.

## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

When the user types `/graphify`, use the installed graphify skill or instructions before doing anything else.

Rules:
- When graphify-out/graph.json exists, run `graphify query "<question>"` before your first grep, search or source-file read for a codebase question, and put this rule in every subagent prompt that explores code. Read raw files only for lines the graph or an earlier read already located. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- Dirty graphify-out/ files are expected after hooks or incremental updates; dirty graph files are not a reason to skip graphify. Only skip graphify if the task is about stale or incorrect graph output, or the user explicitly says not to use it.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `graphify update .` to keep the graph current (AST-only, no API cost).
- A committed graphify-out/graph.json is rewritten whole on every build, so a large diff to it in a pull request is expected. Do not flag it as churn or suggest skipping the graph rebuild or its hooks.

## Documentation routing

新增、修改或審查任何自有 Markdown 前，MUST 讀取並遵守[文件規範](docs/development/documentation.md)。協調者按需分派 documentation 角色並明確指派檔案所有權；所有角色均依該規範自主修正任務範圍內有證據的問題，不擴展為全庫重寫。

新增、修改或審查業務模型、API 契約、資源路徑、collection 或查詢欄位前，所有角色必讀 [命名規範](docs/development/naming.md#業務實體集合與-api-命名)，遵守單一業務實體用單數、集合與 REST 資源路徑用複數的規則及既有相容例外。完整規範只維護於該文件；不得直接重新命名既有公開契約或保存資料。

Read docs/README.md for the document index. Follow docs/agents/workflow.md and
docs/agents/acceptance.md for assignments and acceptance. Roles reference scoped
documents in their developer_instructions; read relevant pages on demand, not all
docs for every task. MCP selection and diagnosis: docs/tooling/mcp/README.md.
Codex issues: OpenAI Docs MCP first; SDK/version questions: Context7.
Connection settings belong in .codex/config.toml; tool versions/tasks/non-secret
environment belong in mise.toml. docs describe usage and design without copying
secrets or version pins. Generated upstream skill files remain under .agents/skills
and .codex/skills and must not become duplicate project policy.

## shadcn 任務入口

凡前端 UI 任務，主動先讀 [官方元件保護、完整目錄與 Skill／MCP 工作流程](frontend/miniapp/src/shared/ui/AGENTS.md)，優先探索及組合既有官方元件，不需使用者逐次指定 shadcn。完整元件索引只維護於該檔；MCP 未曝光時依其固定 CLI 流程查詢並回報限制。

涉及 shadcn/ui、components.json、registry、preset 或其 MCP 的任務，開始前必讀 [.agents/skills/shadcn/SKILL.md](.agents/skills/shadcn/SKILL.md)，再按需讀其 references；使用方式與驗證見 [shadcn 工具指南](docs/tooling/mcp/shadcn.md)。所有角色都適用。
上游 skill 保留原樣；其 `npx shadcn@latest` 等範例在本專案一律改用 `bash scripts/tooling/tool-exec.sh shadcn`，並將工作目錄指向 `frontend/miniapp`。先讀實際元件文件、檢查已安裝元件，更新前使用 `--dry-run`／`--diff`；不因 skill 範例自動重建專案或安裝所有元件。
程式關係探索遵循上方 Graphify 規則；需跨檔來源核對時讀 `.agents/skills/repomix-explorer/SKILL.md`，以 Repomix CLI 打包所需範圍。純工具文件工作不要求每次全庫打包；圖譜過期須回報，不能把舊路徑當目前實作證據。

## shadcn 元件保護

Codex 預設不得直接修改 `frontend/miniapp/src/shared/ui/*.tsx`、`shared/hooks/use-mobile.ts` 或 `src/shell/sidebar-07/*.tsx` 的官方樣式、結構與互動。不得為配合單一頁面需求更改 primitives，亦不得用全域 CSS 覆蓋其外觀。
業務需求優先透過公開 props、variants、資料與事件，由 `shell/AppFrame.tsx`、`shell/Shell.tsx` 或業務層組合；sidebar-07 保留既有真實導覽、身分與登入操作，不換回示範帳號、假連結或訂閱／帳單功能。
只有使用者明確要求修改、還原或更新上述元件時才可更動，且授權只適用當次任務；先用固定版 CLI `add --dry-run`／`--diff` 核對，再最小修改並執行前端檢查。不要自行執行 `add sidebar-07 --overwrite`：它會產生另一套 shared 元件及 dashboard，無法直接更新目前 shell 中的整合。
必要的 alias、client directive 與已記錄的可及性修正保留。不得藉套用 upstream 移除 `Progress` 傳給 Root 的 `value`，使進度值無法被輔助工具讀取。上游 skill 保留原樣，本專案保護規則優先於其一般客製化建議。

## Repository publication default

For future user-requested repository changes, successful task verification includes committing only task-owned changes, pushing `work`, and fast-forwarding `main` to the same verified commit without asking again. Follow [the publication workflow](docs/agents/workflow.md#repository-publication). User instructions such as no push, analysis/plan only, or local only override this default. Unrelated changes and manual cloud deployment are outside this authorization; never force or bypass protected branches.

## gcloud sign-in routing

For gcloud login, follow [the short sign-in workflow](docs/tooling/mcp/google.md#gcloud-sign-in). Keep one orchestrator executor for the whole OAuth process; the top-level agent relays the clickable URL and user code to that same executor. Reuse valid credentials first; otherwise provide the headless login URL immediately. Do not start browser/desktop discovery or additional specialists unless the user asks. Keep CLI login separate from ADC and resource changes.

## Deployment and browser location routing

Current scope follows docs/architecture/decisions/0002-baseline.md and
docs/development/foundation.md: foundation first, then LINE/Firestore, then basic
attendance only. Platform expansion documents are deferred reference material;
do not require Worker, Garr, geolocation or new test infrastructure for this phase.

Application deployment uses Cloud Run only. Read docs/architecture/decisions/0003-cloud-run-primary.md
and docs/operations/cloud-run.md before deployment work; use a dedicated runtime service account and minimal IAM.
For location tasks read docs/integrations/web-geolocation.md: browser JavaScript obtains location,
Go validates untrusted coordinates and authorization. No location history or presence verification
without the product contract. Read docs/architecture/codex-readiness.md and
docs/development/naming.md for implementation handoffs. Preserve generated upstream skills.

## Nested directory instructions

Root instructions apply throughout the repository. Before changing a subtree, read its AGENTS.md and README.md; more specific instructions supplement root policy within their scope, subject to user and host instructions. Directory entry points are listed in [README.md](README.md#巢狀目錄導覽).

Scoped pairs exist in `.agents`, `.codex`, `.github`, `.vscode`, `api`, `assets`, `backend`, `deploy`, `docs` and `frontend`. Within `frontend/miniapp`, also follow its existing generated Next.js guidance. Preserve upstream skill files and existing local changes. A nested guide never grants additional push, publication, deployment or external-write authorization.

## Business Operation 開發與持續符合性

新增、修改或審查業務操作前，必須閱讀 [操作模型](docs/architecture/business-operation.md) 與 [符合性流程](docs/agents/workflow.md#business-operation-compliance)。
在對應功能文件回答六個問題：誰操作、操作什麼、做什麼、有沒有權限、符合哪些業務規則、成功後改變或回傳什麼；補充 Context、Authority、Decision 與依風險需要的 Event／Audit。
區分系統允許／拒絕與業務核准／駁回；Merge 僅於採納提案時使用。
後端執行授權與業務規則，各領域保留獨立狀態模型，共用必要機制，避免不必要抽象。

任務範圍內發現不符合時，以 Graphify 定位關係、Repomix 核對內容、Context7 查證相關技術，再最小修正並依 acceptance 驗證。工具無法使用、圖譜過期或打包遺漏時明確列出限制，不猜測通過。
範圍外發現列出證據與後續工作，不自行全庫重構、改產品規則或部署。純文件及設定調整不要求無關應用測試；修改行為須執行相應檢查。
交付對照規格、實作位置、實際驗證結果及剩餘限制；角色指引與本規則一致，不因委派增加授權。
