# Repository agent guidance

## 從這裡開始

從使用者描述判斷要交付的結果，自行選用合適工具。要求修復時，追到原因、修正並用原始案例驗收；只有完成或遇到無法自行排除的阻礙才結束。

- 開始每個任務前，MUST 閱讀[環境與能力入口](docs/development/environment/README.md#依任務選擇能力)，依任務選擇適用能力，並讀完對應指引後才開始實質操作。不得只讀能力名稱或憑記憶跳過；同一任務可重用已讀且未變更的指引，任務範圍改變時重新選擇。交付時說明採用的能力與實際驗證結果。
- Codex Cloud 首次初始化使用 `scripts/bootstrap.sh`。詳見 [Cloud 環境設定](docs/development/environment/codex-cloud.md)。
- 日常命令從根目錄使用 `bash scripts/tooling/tool-exec.sh <命令>`。
- 修改前遵循從根目錄到目標目錄的 AGENTS.md；需要查入口、操作方式或領域背景時，再讀最近的 README.md。跨目錄搬移時核對來源、目的地與呼叫端，保留既有變更與上游技能。[巢狀導覽](README.md#巢狀目錄導覽)
- 新增或搬移檔案時，按[驗收矩陣](docs/development/tooling/testing/commands.md#變更驗收矩陣)驗證行為。
- 專案不使用 Codex hooks；依任務按需使用 Repomix、Graphify 或直接查閱原始文件。
- 使用者要求「交接」、「接手」或「交給另一個側欄聊天」時，直接採用[側欄交接流程](docs/development/workflow/handoff.md#sidebar-handoff)，輸出六欄摘要及接手所需的來源。

## Backend → API → Frontend

- Backend 是業務邏輯、資料規則、權限決策與狀態轉換的唯一權威來源。
- API 是前後端唯一契約邊界，統一資料結構與介面定義。
- Frontend 保持最薄，只負責 UI 呈現、使用者互動、必要的前端驗證與 UI 狀態。
- 禁止 Frontend 重複實作 Backend 業務邏輯、權限決策或資料規則。
- 發現違反上述原則時，主動檢查依賴與影響範圍，將邏輯移至正確層級，移除重複實作並更新契約與測試。
- 優先搬移、共用或刪除既有程式碼，避免複製、額外抽象與功能膨脹。
- 保持既有行為與 API 相容性；無法確認安全性的變更，先報告而不擅自修改。

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

- 本機專案 CLI 必須透過 `bash scripts/tooling/tool-exec.sh <command> [args...]` 執行；wrapper 從 mise.toml 辨識直接命令並自動安裝宣告版本。
- 複合 shell 命令 MUST 明確宣告所需工具，例如 `bash scripts/tooling/tool-exec.sh --tool node --tool npm:repomix bash -c 'repomix --version'`。`--tool` 只使用 mise.toml 已宣告的工具，不另指定版本。
- 缺少工具時 MUST 優先使用 wrapper 或既有 install-* 任務，驗證版本後繼續原任務；MUST NOT 自行升級版本、修改鎖檔或安裝未宣告套件。
- `mise run` 已開啟任務自動安裝，可能準備任務環境的多個缺少工具；只需單一 CLI 時 MUST 使用 wrapper 的直接命令。
- MCP 缺少、呼叫失敗或需要整體驗收時，依[共用 MCP 排查流程](docs/development/tooling/agents/README.md#project-loading-check)處理並繼續可獨立完成的工作。
- 新環境仍 MUST 先執行初始化以準備 mise、系統 Python 與網路；Graphify 使用既有 install-graphify 任務，不能只靠 mise 的工具宣告安裝。

依據：mise 2026.11.3 `exec --help` 與 [官方設定](https://mise.jdx.dev/configuration/settings.html#exec_auto_install)。

## Mandatory documentation verification

- **本地行為**：直接檢查目前 checkout 的實作，用相關命令或測試驗證。若判斷取決於外部套件的行為，再查對應版本的文件。
- **外部行為**：依[工具與登入指引](docs/development/tooling/agents/AGENTS.md)選官方來源；使用 Context7 時先解析 library ID，再以專案依賴版本查詢。遇到配額限制時走[備援流程](docs/development/tooling/agents/context7.md#anonymous-fallback)。
- 回報實際證據、適用版本及未確認處。來源不足時暫緩依賴該證據的決定，繼續可獨立完成的工作。文件查證不增加外部寫入授權，也不能取代行為驗收。

## 單一代理與能力選用

由目前代理承擔規劃、探索、設計、實作、自我審查、測試與文件維護，不啟動子代理或委派給 orchestrator。依[能力入口](docs/development/workflow/capabilities/README.md)按任務讀取專業指引，工作流程見[開始任務](docs/development/workflow/task-start.md)。保留既有變更；自我審查不等於獨立審查。工具與能力不增加外部寫入權限。

架構、資料關係、流程或狀態用圖更清楚時，主動使用專案鎖定的本機 Mermaid v13 CLI 驗證並渲染圖表。操作方式見 [Mermaid 工具指南](docs/development/tooling/shared/mermaid.md)。

## graphify

The code graph lives under `graphify-out/`. For Markdown headings and links, use the separate [docs-only graph](docs/development/tooling/shared/graphify.md#docs-only-graph).

When the user types `/graphify`, use the installed graphify skill or instructions before doing anything else.

Rules:
- Use Graphify first for unfamiliar relationships when the graph is available. Its results are leads. Search the current tree with rg, language references and direct source reads to locate additional callers, routes, configuration, generated-source inputs and fixtures, including paths absent from the graph. For known files, configuration and exact identifiers use direct inspection. Continue with source search if the graph is unavailable or truncated.
- Local graph changes after incremental updates are expected and do not prevent queries.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `bash scripts/tooling/tool-exec.sh mise run graphify-update` to refresh the code graph without API calls.
- The code graph is generated locally and ignored by Git. Refresh it in the current checkout when the task needs updated code relationships.

程式關係探索遵循上方 Graphify 規則；需跨檔來源核對時讀 `.agents/skills/repomix-explorer/SKILL.md`，以 Repomix CLI 打包所需範圍。純工具文件工作不要求每次全庫打包；圖譜過期須回報，不能把舊路徑當目前實作證據。

## Documentation routing

新增、修改或審查任何自有 Markdown 前，MUST 讀取並遵守[文件規範](docs/standards/documentation.md)。目前代理按需採用文件能力指引並確認檔案所有權，自主修正任務範圍內有證據的問題，不擴展為全庫重寫。

依[開發手冊](docs/development/README.md#按任務開始)選擇環境、工作流程、配方與驗收。工具選擇與診斷見 [MCP 指南](docs/development/tooling/agents/README.md)。

實作交接依[交接與證據重用](docs/development/workflow/handoff.md#minimal-handoff-and-evidence-reuse)與[命名規範](docs/standards/naming.md)。

## Repository publication default

Develop directly on `main`. Every commit must be a complete, independently reviewable, verified milestone. Push only completed milestones within the authorized scope. Follow the [publication workflow](docs/development/workflow/publication.md#repository-publication) and respect branch protections.

## 工程決策與交付要求

- 工程任務開始前，閱讀並遵循[工程原則與決策](docs/standards/engineering.md)。
- 涉及介面、互動或使用者流程時，另閱讀並遵循[產品與體驗設計原則](docs/standards/design.md)。

每次交付說明解決的問題、實際執行的驗證與結果，以及仍未驗證的項目。每次在 `main` 提交或推送仍須符合[階段交付門檻](docs/development/workflow/publication.md#repository-publication)。

## 共通驗證與交付

任務範圍內發現不符合時，依上方探索與查證規則選用工具，找出並修正任務範圍內的根因，控制變更範圍，再依 acceptance 驗證。工具無法使用、圖譜過期或打包遺漏時明確列出限制，不猜測通過。
範圍外發現列出證據與後續工作，不自行全庫重構、改產品規則或部署。純文件及設定調整不要求無關應用測試；修改行為須執行相應檢查。
涉及 Firestore 交易、併發、重送或持久資料相容性時，依[整合驗收流程](docs/development/tooling/testing/commands.md#firestore-integration)主動檢查、準備並重跑 emulator 測試。
交付對照規格、實作位置、實際驗證結果及剩餘限制。
