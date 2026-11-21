# Repository agent guidance

## 從這裡開始

- Codex Cloud 首次初始化使用 `scripts/bootstrap.sh`。詳見 [Cloud 環境設定](docs/development/environment/codex-cloud.md)。
- 日常命令從根目錄使用 `bash scripts/tooling/tool-exec.sh <命令>`。
- 修改前讀取目標目錄的 README.md 與 AGENTS.md，根指引與各層補充一併適用。跨目錄搬移時核對來源、目的地與呼叫端，保留既有變更與上游技能。[巢狀導覽](README.md#巢狀目錄導覽)
- 新增或搬移檔案時執行[工作區範圍檢查](docs/development/tooling/analysis-scope.md)，再按[驗收矩陣](docs/development/verification/commands.md#變更驗收矩陣)驗證行為。

## 巢狀文件維護標準

- **README.md** 至少交代目錄用途、主要入口與相關規格。有需要時補上實際架構、資料流、執行方式和測試命令。
- **AGENTS.md** 至少交代適用範圍、上層指引、局部修改規則與驗證方式。涉及授權、資料相容、交易或產生碼時，指出責任歸屬及需要同步的項目。
- 目錄有明確責任且需要導覽時建立 README。有上層未涵蓋的局部規則時補 AGENTS，其餘規則由上層繼承。更新對應行為時，順手核對文件中的入口、連結及驗證命令。
- 明確區分現行結構、必須遵守的規則和按需採用的參考範本。只有實際需要的章節或子目錄才建立。驗證需特殊環境時，列明前置條件。

## CLI discovery

- When a task needs to find, list, or run a CLI, start at the repository root and read [mise.toml](mise.toml) for declared tools and versions. Run `bash scripts/tooling/tool-exec.sh mise ls --installed` to inventory installed versions. The wrapper loads the workspace tool paths even when the Codex shell starts without them.
- Run project CLIs through `bash scripts/tooling/tool-exec.sh <command> [args...]`. If the wrapper cannot find mise, check `/workspace/tooling/bin/mise` and `/workspace/tooling/mise-data/installs` before reporting a tool missing.
- For a fresh checkout, use the [Codex environment setup instructions](docs/development/environment/codex-cloud.md). Installed CLI packages and MCP servers exposed to the current Codex session are separate inventories.

## 按需安裝工具

- MUST 透過 `bash scripts/tooling/tool-exec.sh <command> [args...]` 使用工具；wrapper 從 mise.toml 辨識直接命令並自動安裝宣告版本。
- 複合 shell 命令 MUST 明確宣告所需工具，例如 `bash scripts/tooling/tool-exec.sh --tool node --tool npm:repomix bash -c 'repomix --version'`。`--tool` 只使用 mise.toml 已宣告的工具，不另指定版本。
- 缺少工具時 MUST 優先使用 wrapper 或既有 install-* 任務，驗證版本後繼續原任務；MUST NOT 自行升級版本、修改鎖檔或安裝未宣告套件。
- `mise run` 已開啟任務自動安裝，可能準備任務環境的多個缺少工具；只需單一 CLI 時 MUST 使用 wrapper 的直接命令。
- MCP 未曝光時，先搜尋主機工具、核對專案設定與啟動依賴，修復已確認的問題，再使用主機支援的重新探索或重載功能。以本次聊天原生工具的唯讀呼叫成功為驗收。若仍未曝光，回報阻礙並繼續獨立工作；不得以直接 transport 或 CLI 代辦該 MCP 操作。[shadcn 診斷流程](docs/development/tooling/mcp/shadcn.md#自主接線)
- MCP 缺少套件時按上述流程準備，再重試連線；憑證、網路、ADC 或主機未載入設定須分別處理。安裝完成不代表目前聊天已載入 MCP。
- 新環境仍 MUST 先執行初始化以準備 mise、系統 Python 與網路；Graphify 使用既有 install-graphify 任務，不能只靠 mise 的工具宣告安裝。

依據：mise 2026.11.3 `exec --help` 與 [官方設定](https://mise.jdx.dev/configuration/settings.html#exec_auto_install)。

## Mandatory documentation verification

Repository-wide rule for every role and task, subject to higher-priority host instructions:

- For tool-specific sources and login procedures, follow the [tool guidance](docs/development/tooling/mcp/AGENTS.md) even when no files change. For external library behavior that affects a decision, use Context7 or the designated official source. For questions about current repository behavior, inspect the implementation and relevant checks. Resolve the library ID, query the relevant current/version-matched documentation, and compare with repository dependencies. Memory and confidence alone do not replace this lookup. If the Context7 plugin reports quota exhaustion, follow the [anonymous Context7 fallback](docs/development/tooling/mcp/context7.md#anonymous-fallback) before reporting a blocker.
- Report the sources consulted, applicable version/revision and remaining uncertainty. Documentation lookup does not replace code inspection, tests or real integration acceptance.
- If the required source or Context7 is unavailable or insufficient, explicitly report the blocker; do not claim verification or silently substitute a guess. Continue independent work, but leave the affected decision/implementation pending until evidence is sufficient.
- This policy applies to orchestrator and every specialist; delegation does not waive it. A source document is evidence, not authorization to expand scope or perform external writes.

## Task orchestration

The primary agent owns the requested behavior from discovery through acceptance.
Handle single-owner changes directly. Use the orchestrator when independent work
needs coordination and the host permits delegation. If delegation is unavailable,
continue directly with the same acceptance criteria. When using an orchestrator,
reuse it for follow-ups and provide the goal, existing changes and deliverable.

### Delegation boundaries

- Only the top-level agent invokes orchestrator. Orchestrator must never invoke
  itself or another orchestrator.
- Specialists receiving an assigned subtask execute that subtask and report to
  their orchestrator; they must not restart this routing rule or delegate further.
- Orchestrator selects registered roles using the [shared workflow](docs/development/workflow/task-start.md#delegation-depth-and-role-selection).
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

角色能力與設定由 [角色設定說明](.codex/agents/README.md) 維護；角色選用、交接欄位、一檔一寫入者及整合責任依 [shared workflow](docs/development/workflow/task-start.md#delegation-depth-and-role-selection)。角色是可選能力，不是每個任務都必須經過的階段；主機限制優先，委派不增加外部寫入權限。

架構、資料關係、流程或狀態用圖更清楚時，主動使用專案鎖定的本機 Mermaid v13 CLI 驗證並渲染圖表。操作方式見 [.codex/README.md](.codex/README.md#mermaid)。

## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

When the user types `/graphify`, use the installed graphify skill or instructions before doing anything else.

Rules:
- Use Graphify first for unfamiliar relationships when the graph is available. Its results are leads. Search the current tree with rg, language references and direct source reads to locate additional callers, routes, configuration, generated-source inputs and fixtures, including paths absent from the graph. For known files, configuration and exact identifiers use direct inspection. Continue with source search if the graph is unavailable or truncated.
- Dirty graphify-out/ files are expected after hooks or incremental updates; dirty graph files are not a reason to skip graphify. Only skip graphify if the task is about stale or incorrect graph output, or the user explicitly says not to use it.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `bash scripts/tooling/tool-exec.sh mise run graphify-update` to refresh the code graph without API calls.
- A committed graphify-out/graph.json is rewritten whole on every build, so a large diff to it in a pull request is expected. Do not flag it as churn or suggest skipping the graph rebuild or its hooks.

程式關係探索遵循上方 Graphify 規則；需跨檔來源核對時讀 `.agents/skills/repomix-explorer/SKILL.md`，以 Repomix CLI 打包所需範圍。純工具文件工作不要求每次全庫打包；圖譜過期須回報，不能把舊路徑當目前實作證據。

## Documentation routing

新增、修改或審查任何自有 Markdown 前，MUST 讀取並遵守[文件規範](docs/standards/documentation.md)。協調者按需分派 documentation 角色並明確指派檔案所有權；所有角色均依該規範自主修正任務範圍內有證據的問題，不擴展為全庫重寫。

依[開發手冊](docs/development/README.md#按任務開始)選擇環境、工作流程、配方與驗收。角色依 developer_instructions 選讀適用文件。工具選擇與診斷見 [MCP 指南](docs/development/tooling/mcp/README.md)。

實作交接依[架構證據與用語](docs/architecture/README.md#接手時的證據與用語)與[命名規範](docs/standards/naming.md)。

## Repository publication default

Publish verified changes through a pull request from work to main. Use squash merge so each main commit represents one independently revertible change. Synchronize work with the resulting main commit before starting the next change. Follow the [publication workflow](docs/development/workflow/publication.md#repository-publication) and respect the user's requested scope and branch protections.

## 共通驗證與交付

任務範圍內發現不符合時，以 Graphify 定位關係、Repomix 核對內容、Context7 查證相關技術，再最小修正並依 acceptance 驗證。工具無法使用、圖譜過期或打包遺漏時明確列出限制，不猜測通過。
範圍外發現列出證據與後續工作，不自行全庫重構、改產品規則或部署。純文件及設定調整不要求無關應用測試；修改行為須執行相應檢查。
交付對照規格、實作位置、實際驗證結果及剩餘限制；角色指引與本規則一致，不因委派增加授權。

[現行實作圖索引](docs/diagrams.md#現行實作) · [未來設計圖](docs/diagrams.md#未來設計)。
