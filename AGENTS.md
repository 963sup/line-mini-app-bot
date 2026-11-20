# Repository agent guidance

## Nested directory instructions

Root instructions apply throughout the repository. Before changing a subtree, read its AGENTS.md and README.md; more specific instructions supplement root policy within their scope, subject to user and host instructions. Directory entry points are listed in [README.md](README.md#巢狀目錄導覽).

涉及跨目錄任務時，讀取所有相關目錄的指引。保留既有變更與上游技能；巢狀指引不增加外部寫入授權。

## 巢狀文件維護標準

- **README.md** 至少交代目錄用途、主要入口與相關規格。有需要時補上實際架構、資料流、執行方式和測試命令。
- **AGENTS.md** 至少交代適用範圍、上層指引、局部修改規則與驗證方式。涉及授權、資料相容、交易或產生碼時，指出責任歸屬及需要同步的項目。
- 目錄有明確責任且需要導覽時建立 README。有上層未涵蓋的局部規則時補 AGENTS，其餘規則由上層繼承。更新對應行為時，順手核對文件中的入口、連結及驗證命令。
- 明確區分現行結構、必須遵守的規則和按需採用的參考範本。只有實際需要的章節或子目錄才建立。驗證需特殊環境時，列明前置條件。

## Mandatory documentation verification

Repository-wide rule for every role and task, subject to higher-priority host instructions:

- For tool-specific sources and login procedures, follow the [tool guidance](docs/development/tooling/mcp/AGENTS.md) even when no files change. For other technical questions, and whenever implementation is not fully certain, use Context7 before deciding or implementing. Resolve the library ID, query the relevant current/version-matched documentation, and compare with repository dependencies. Memory and confidence alone do not replace this lookup. If the Context7 plugin reports quota exhaustion, follow the [anonymous Context7 fallback](docs/development/tooling/mcp/context7.md#anonymous-fallback) before reporting a blocker.
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
- Orchestrator selects registered roles using the [shared workflow](docs/development/workflow.md#delegation-depth-and-role-selection).
  Small tasks can be completed directly; roles are capabilities, not a required fan-out.
- Assign one writer per file scope; parallelize only independent tasks. Respect
  the configured bounds and host's actual concurrency and depth limits.
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

角色能力與設定由 [.codex/agents](.codex/agents) 維護；角色選用、交接欄位、一檔一寫入者及整合責任依 [shared workflow](docs/development/workflow.md#delegation-depth-and-role-selection)。角色是可選能力，不是每個任務都必須經過的階段；主機限制優先，委派不增加外部寫入權限。

## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

When the user types `/graphify`, use the installed graphify skill or instructions before doing anything else.

Rules:
- When graphify-out/graph.json exists, run `graphify query "<question>"` before your first grep, search or source-file read for a codebase question, and put this rule in every subagent prompt that explores code. Read raw files only for lines the graph or an earlier read already located. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- Dirty graphify-out/ files are expected after hooks or incremental updates; dirty graph files are not a reason to skip graphify. Only skip graphify if the task is about stale or incorrect graph output, or the user explicitly says not to use it.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `bash scripts/tooling/tool-exec.sh mise run graphify-update` to refresh the code graph without API calls.
- A committed graphify-out/graph.json is rewritten whole on every build, so a large diff to it in a pull request is expected. Do not flag it as churn or suggest skipping the graph rebuild or its hooks.

程式關係探索遵循上方 Graphify 規則；需跨檔來源核對時讀 `.agents/skills/repomix-explorer/SKILL.md`，以 Repomix CLI 打包所需範圍。純工具文件工作不要求每次全庫打包；圖譜過期須回報，不能把舊路徑當目前實作證據。

## Documentation routing

新增、修改或審查任何自有 Markdown 前，MUST 讀取並遵守[文件規範](docs/standards/documentation.md)。協調者按需分派 documentation 角色並明確指派檔案所有權；所有角色均依該規範自主修正任務範圍內有證據的問題，不擴展為全庫重寫。

Read docs/README.md for the document index. Follow docs/development/workflow.md and
docs/development/acceptance.md for assignments and acceptance. Roles reference scoped
documents in their developer_instructions; read relevant pages on demand, not all
docs for every task. MCP selection and diagnosis: docs/development/tooling/mcp/README.md.

實作交接依[架構證據與用語](docs/architecture/README.md#接手時的證據與用語)與[命名規範](docs/standards/naming.md)。

## Repository publication default

Publish verified changes through a pull request from work to main. Use squash merge so each main commit represents one independently revertible change. Synchronize work with the resulting main commit before starting the next change. Follow the [publication workflow](docs/development/workflow.md#repository-publication) and respect the user's requested scope and branch protections.

## 共通驗證與交付

任務範圍內發現不符合時，以 Graphify 定位關係、Repomix 核對內容、Context7 查證相關技術，再最小修正並依 acceptance 驗證。工具無法使用、圖譜過期或打包遺漏時明確列出限制，不猜測通過。
範圍外發現列出證據與後續工作，不自行全庫重構、改產品規則或部署。純文件及設定調整不要求無關應用測試；修改行為須執行相應檢查。
交付對照規格、實作位置、實際驗證結果及剩餘限制；角色指引與本規則一致，不因委派增加授權。

[現行實作圖索引](docs/diagrams.md#現行實作) · [未來設計圖](docs/diagrams.md#未來設計)。
