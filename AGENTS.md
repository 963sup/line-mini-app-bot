# Repository agent guidance

## 從這裡開始

任務入口依第 18 條，目錄指引見[巢狀導覽](README.md#巢狀目錄導覽)。

- 專案不使用 Codex hooks；依任務按需使用 Repomix、Graphify 或直接查閱原始文件。

## 共通工程規則

1. **Single Source of Truth**：為每項業務規則、資料模型、API 契約、設定與文件指定唯一權威來源；其他位置以引用或自動生成方式保持同步。
2. **Backend → API → Frontend**：將業務規則、資料存取、授權與狀態轉換交由 Backend 執行，經 API 提供結果；Frontend 僅處理呈現、互動、即時輸入回饋與必要的客戶端狀態。 Backend 統一業務計算、驗證與資料一致性檢查，API 直接提供各使用情境需要的結果。可由既有資料推導的顯示值在呈現時算出，避免新增 state 或 Effect。
3. **Contract First**：先用 OpenAPI 定義 Schema、序列化方式、請求、回應與錯誤格式，再據此實作；優先生成型別與 Client，避免手動維護另一份契約。在 API 邊界驗證外部資料，內部使用契約型別；區分缺值、null 與零值。沿用契約的 HTTP 狀態與穩定錯誤碼，客戶端依錯誤碼處理。清單端點依需求明訂分頁、篩選、穩定排序與回傳欄位。
4. **Standard First**：優先遵循 HTTP 語意、REST 慣例、Go 慣用寫法及既有架構；沿用成熟標準與現有工具，避免自訂協定、格式、框架或加入不必要的依賴。
5. **Explicit Boundaries**：依 DDD Bounded Context 定義領域模型、職責、資料所有權、模組介面與依賴方向；跨領域操作須經由明確介面，避免存取內部實作或形成循環依賴。命名沿用領域共通語言；由組裝入口提供外部依賴，讓領域規則可獨立驗證。
6. **Minimal Abstraction**：遵循 KISS、YAGNI 與 DRY，優先採用直接、明確的實作；確認有實際共通需求後才提取抽象。 提取前先評估正確性、可靠性、可讀性、認知負擔與維護成本。
7. **Consistent Conventions**：相同語意沿用相同的命名、目錄、格式、型別、識別碼、時間、序列化、驗證與錯誤慣例；查詢、分頁、排序、過濾、狀態及資源生命週期亦遵循既有約定。命名與套件配置遵循[命名規範](docs/standards/naming.md)。公開資源 ID 維持穩定；時間點依契約明訂格式與時區（如 RFC 3339 UTC），業務日期以領域時區判定。
8. **Explicit Business Behavior**：由負責的領域統一定義並執行業務操作、權限決策、狀態機、合法轉換與資料驗證，明訂不變條件、前置條件、副作用與錯誤語意。對建立、更新、封存與刪除等支援的操作，定義允許的起始狀態、結束狀態及關聯資料清理責任。
9. **Reliable Operations**：依操作需求界定交易與一致性要求，處理併發、冪等性、逾時、取消、去重及故障恢復；可重試的操作須定義退避策略與停止條件，避免重複副作用。涉及競爭寫入時，以交易或版本條件保護不變條件；有副作用的重送沿用同一操作識別。外部依賴持續失敗時，依影響評估是否需要斷路。
10. **Security by Default**：依信任邊界採用安全預設值與最小權限，統一處理身分認證、授權、Session、Token、輸入驗證、敏感資料及機密管理；安全決策由後端執行。依[身分契約](docs/contracts/identity.md)驗證 Session 或 Token，再由後端按操作、資源與成員關係判斷權限。
11. **Event & Integration Consistency**：界定同步與非同步整合邊界，統一事件命名、Schema、版本、發布、消費與處理語意；依交付保證處理順序、重複投遞、重試與相容性。跨系統交換事件時先核對接收方支援的格式與版本，再決定是否採用 CloudEvents。
12. **Observable Operations**：統一結構化日誌、追蹤、指標、Correlation ID 與錯誤分類；優先使用 OpenTelemetry，讓關鍵操作、效能與失敗可量測，跨服務請求可追蹤與診斷。使同一操作的日誌與追蹤可透過識別碼關聯；指標標籤使用有限的分類，避免放入個別資源 ID。
13. **Compatibility & Lifecycle**：統一版本策略、API 相容性、資料遷移、設定管理與資源建立及刪除流程；優先採用向後相容的變更與漸進式遷移，保護既有使用者、資料及整合。需跨版本部署的資料變更先擴充相容結構，完成資料遷移並確認舊讀寫端退場後，再移除舊結構。對外發布版本依其相容性政策標示變更。
14. **Automated Enforcement**：優先用格式化、Lint、型別檢查、Schema 驗證、程式碼生成、依賴檢查、測試與 CI 自動執行可機械化的工程檢查。契約變更後重新生成並檢查差異，確保提交的產物能由來源重建。
15. **Minimal Change & Continuous Validation**：先分析依賴、根因與影響範圍，再以最小必要變更修正；每次變更依影響範圍驗證型別、契約、測試結果、安全性、相容性與核心行為。
16. **Continuous Simplification**：持續刪除或合併重複、衝突、過時、無效及不必要的程式碼、抽象、依賴、設定、文件與流程；以實際量測比較改善前後的差異，並以回歸驗證確認行為。
17. **Task Scope & Acceptance**：從使用者描述釐清真正目標、限制與驗收標準。依第一性原理選擇能達成目標的最小可靠方案與合適工具。按影響與風險排序，驗收失敗時回查根因，達標或遇到無法自行排除的阻礙時結束。範圍外發現另列證據與後續工作。
18. **Applicable Instructions**：開始每個任務前，MUST 閱讀[環境與能力入口](docs/development/environment/README.md#依任務選擇能力)，依任務選擇適用能力，並讀完對應指引後才開始實質操作。不得只讀能力名稱或憑記憶跳過；同一任務可重用已讀且未變更的指引，任務範圍改變時重新選擇。修改前，沿目標檔案的目錄路徑讀取所有適用的 AGENTS.md。在主機權限與使用者指定的任務範圍內，根 AGENTS.md 定義共通規則，目標路徑上較近的 AGENTS.md 優先處理局部事項。
19. **Single-Agent Execution**：由目前代理承擔規劃、探索、設計、實作、自我審查、測試與文件維護，不啟動子代理或委派給 orchestrator。保留既有變更與上游技能，如實標示自我審查；工具與能力不增加外部寫入授權。
20. **Source Verification**：本地行為以目前 checkout 的原始碼與實測為準，外部行為查對應版本的官方文件。優先查證會影響決策的假設與失敗風險。版本、來源與適用條件仍相符時，沿用已驗證的證據。文件查證不能取代行為驗收。
21. **Targeted Exploration**：從具體問題開始探索。陌生依賴關係優先用可用的 Graphify 定位，已知識別字直接查原始碼。以目前工作樹核對圖譜與 Repomix 線索，證據足以回答問題時停止擴大搜尋。
22. **Pinned Toolchain**：優先使用現有工具，以 [mise.toml](mise.toml) 為版本來源，透過專案 wrapper 或既有安裝任務執行。缺少工具時核對宣告版本與安裝狀態，不自行升級、改鎖檔或安裝未宣告套件。建置使用版本庫內的工具與依賴鎖定資料，讓乾淨 checkout 可用相同輸入重建。
23. **Documentation Ownership**：以 README 說明目錄用途、入口與規格。只有上層未涵蓋局部規則時才新增 AGENTS.md，寫明範圍、上層指引、局部職責與驗證方式；行為變更時同步核對入口、連結與命令。API 文件引用 OpenAPI；影響多個模組或日後難以逆轉的架構決策，在既有決策文件或 ADR 記錄理由與取捨。
24. **Regression Verification**：修復後重跑原始案例。新增或搬移檔案時，按[驗收矩陣](docs/development/tooling/testing/commands.md#變更驗收矩陣)驗證行為。跨目錄搬移須核對來源、目的地與呼叫端。涉及 Firestore 交易、併發、重送或持久資料相容性時，依[整合驗收流程](docs/development/tooling/testing/commands.md#firestore-integration)主動檢查、準備並重跑 emulator 測試。依風險選擇驗證層級：領域規則用單元測試，儲存與服務邊界用整合或契約測試，關鍵使用者流程用端到端測試。
25. **Failure Isolation**：工具無法使用、圖譜過期或打包遺漏時，改用可用來源查證並記錄限制。只暫緩依賴缺失證據的決定，繼續可獨立完成的工作。
26. **Verified Delivery**：交付時對照規格與驗收條件，列出實作位置、採用的能力、實際檢查、結果與適用版本，區分已完成的變更、已驗證的行為與剩餘限制。
27. **Verified Publication**：預設在 `main` 開發，每筆提交須是完整、已驗證且可獨立審查的里程碑。僅在授權範圍內推送，並遵守發布流程與分支保護。提交前審查暫存差異，確認符合驗收條件，並以提交訊息說明具體變更。
28. **Explicit Handoff**：使用者要求「交接」、「接手」或「交給另一個側欄聊天」時，直接採用[側欄交接流程](docs/development/workflow/handoff.md#sidebar-handoff)，輸出六欄摘要及接手所需的來源。
29. **Configuration Boundaries**：為環境設定定義來源、優先順序、必要值與安全預設；在載入邊界解析並驗證型別。缺少必要值時清楚指出設定名稱，診斷輸出不得包含機密值。
30. **UI & Form Consistency**：依[Mini App UI 規範](docs/standards/miniapp-ui.md)重用元件與互動慣例。表單欄位依 API 契約映射，區分未填、清空與未修改；即時驗證與後端欄位錯誤以一致方式呈現，送出失敗時保留可修正的輸入。
31. **Cache Consistency**：引入快取前先定義可容許的陳舊時間、鍵的資源與授權範圍、有效期限及失效條件。驗證寫入或權限變更後的讀取，以及快取未命中時的行為。
32. **Deployment & Recovery**：部署前核對授權、目標環境與已驗證的產物；部署後依[發布驗收](docs/operations/release-acceptance.md)檢查健康狀態與核心流程。事先確認回退版本與資料相容性，依驗收結果決定繼續發布或復原。

## Backend → API → Frontend

- 優先沿用 Next.js、React 與既有元件。依路由與部署模式選擇元件和資料取得方式。現行 Mini App 使用 Pages Router 與靜態匯出；評估 Server Components 或 Client Components 時須先確認路由支援。
- 發現責任錯置時，先核對呼叫端、依賴與影響，再依 Backend → API → Frontend 順序搬移、合併或刪除重複實作，同步受影響的契約與測試。維持既有功能、API 相容性、安全性與效能。無法確認變更安全性時，先回報具體風險。

## 巢狀文件維護標準

- **README.md** 有需要時補上實際架構、資料流、執行方式和測試命令。
- **AGENTS.md** 涉及授權、資料相容、交易或產生碼時，指出責任歸屬及需要同步的項目。
- 明確區分現行結構、必須遵守的規則和按需採用的參考範本。只有實際需要的章節或子目錄才建立。驗證需特殊環境時，列明前置條件。

<a id="cli-discovery"></a>

## 按需安裝工具

- 新環境先依 [Cloud 環境設定](docs/development/environment/codex-cloud.md)執行 `scripts/bootstrap.sh`，準備 mise、系統 Python 與網路。
- 查找或執行 CLI 時，從根目錄用 `bash scripts/tooling/tool-exec.sh mise ls --installed` 核對已安裝版本。wrapper 找不到 mise 時，檢查 `/workspace/tooling/bin/mise` 與 `/workspace/tooling/mise-data/installs`。
- 執行單一工具用 `bash scripts/tooling/tool-exec.sh <命令> [參數]`，wrapper 會依宣告版本準備工具；`mise run` 可能準備整個任務環境。複合命令須明列已宣告的工具，例如 `bash scripts/tooling/tool-exec.sh --tool node --tool npm:repomix bash -c 'repomix --version'`。
- 缺少 Graphify 時，執行 `mise run install-graphify`。
- 聊天可呼叫的 MCP 須另行確認。缺少服務、呼叫失敗或需要整體驗收時，依[共用排查流程](docs/development/tooling/agents/README.md#project-loading-check)處理，並繼續可獨立完成的工作。

依據：mise 2026.11.3 `exec --help` 與 [官方設定](https://mise.jdx.dev/configuration/settings.html#exec_auto_install)。

## Mandatory documentation verification

- **外部行為**：依[工具與登入指引](docs/development/tooling/agents/AGENTS.md)選官方來源；使用 Context7 時先解析 library ID，再以專案依賴版本查詢。遇到配額限制時走[備援流程](docs/development/tooling/agents/context7.md#anonymous-fallback)。

## 單一代理與能力選用

依[能力入口](docs/development/workflow/capabilities/README.md)按任務讀取專業指引，工作流程見[開始任務](docs/development/workflow/task-start.md)。按驗收所需控制工具呼叫、上下文與中間產物，避免重複分析及拆分工作。

架構、資料關係、流程或狀態用圖更清楚時，主動使用專案鎖定的本機 Mermaid v13 CLI 驗證並渲染圖表。操作方式見 [Mermaid 工具指南](docs/development/tooling/shared/mermaid.md)。

## graphify

The code graph lives under `graphify-out/`. For Markdown headings and links, use the separate [docs-only graph](docs/development/tooling/shared/graphify.md#docs-only-graph).

完成第 18 條的任務起始閱讀後，若使用者指定 `/graphify`，在執行圖譜操作前先讀取已安裝的 Graphify 技能或操作指引。

Rules:
- 依第 21 條探索依賴時，用 rg、語言參照與原始碼補查呼叫端、路由、設定、產生碼輸入與 fixture，涵蓋圖譜遺漏的路徑。
- Local graph changes after incremental updates are expected and do not prevent queries.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `bash scripts/tooling/tool-exec.sh mise run graphify-update` to refresh the code graph without API calls.
- 程式碼圖譜由本機產生並由 Git 忽略。唯讀查詢若發現圖譜過期，且判斷需要最新依賴關係，再更新目前 checkout 的圖譜。

跨檔查詢依[探索指南](docs/development/workflow/capabilities/explorer.md)選擇既有參考技能或限定範圍的打包。圖譜過期時回報限制，並用目前來源核對。

## Documentation routing

新增、修改或審查任何自有 Markdown 前，MUST 讀取並遵守[文件規範](docs/standards/documentation.md)。目前代理按需採用文件能力指引並確認檔案所有權，自主修正任務範圍內有證據的問題，不擴展為全庫重寫。

依[開發手冊](docs/development/README.md#按任務開始)選擇環境、工作流程、配方與驗收。工具選擇與診斷見 [MCP 指南](docs/development/tooling/agents/README.md)。

實作交接依[交接與證據重用](docs/development/workflow/handoff.md#minimal-handoff-and-evidence-reuse)與[命名規範](docs/standards/naming.md)。

## Repository publication default

第 27 條的操作步驟見[publication workflow](docs/development/workflow/publication.md#repository-publication)。

## 工程決策與交付要求

- 工程任務開始前，閱讀並遵循[工程原則與決策](docs/standards/engineering.md)。
- 涉及介面、互動或使用者流程時，另閱讀並遵循[產品與體驗設計原則](docs/standards/design.md)。
- 先核對領域模型與關鍵依賴，按影響程度與風險排序，將有共同根因的問題集中修正。

## 共通驗證與交付

純文件及設定調整不要求無關應用測試；修改行為須執行相應檢查。
