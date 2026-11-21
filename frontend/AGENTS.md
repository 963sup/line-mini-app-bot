# frontend 工作指引

## 閱讀與修改入口

先查閱[miniapp](miniapp)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[文件索引](../docs/README.md) · [模組閱讀入口](../docs/README.md#按模組閱讀)

適用於本目錄及其子目錄；先遵循 [根目錄 AGENTS.md](../AGENTS.md)，再套用較近的子目錄補充。使用者及主機指令優先；本檔不增加外部寫入授權。

## 職責與邊界

miniapp/ 為 Next.js／React workspace；src/pages/ 提供路由，business/attendance/ 提供出勤畫面，identity/、platform/line/、shared/、shell/ 分別處理身分、平台、共用與外框。

進入 miniapp/ 先讀其 AGENTS.md 與安裝版本的 Next.js 指引。HTTP 呼叫透過 @repo/api-client；位置由瀏覽器取得，後端驗證。不手改建置輸出。

## 驗證

修改互動時，核對使用此模組的頁面與 miniapp 測試。

從根目錄透過 tool-exec.sh 執行 pnpm --filter @repo/miniapp check 與相關既有 test；需要靜態輸出時執行 build。API 變更另走 ../api/README.md。

命令除另有標示外均在儲存庫根目錄執行。閱讀 [本目錄 README](README.md) 與 [文件索引](../docs/README.md)；回報實際執行結果與限制。

## 文件閱讀路由

先由 [本模組相關文件](README.md#相關文件) 與 [模組閱讀入口](../docs/README.md#按模組閱讀) 選擇文件。局部用途與命令維護在本模組 README；跨模組架構、功能與營運文件依 docs 主題分類保存。索引更新依[文件維護指引](../docs/AGENTS.md#主題子目錄)。

## Business Operation

開始相關任務前，必須閱讀 [操作模型](../docs/architecture/contracts/business-operations.md) 與 [符合性流程](../docs/development/workflow.md#business-operation-compliance)。

畫面呈現操作可用性、確認與結果；前端隱藏按鈕不能取代後端授權。使用 API 契約處理拒絕、過期版本與失敗，結果未知時遵循既有重試契約，不自行宣告權限或成功。

[現行實作圖：頁面與身分生命週期](../docs/architecture/runtime/react-lifecycle.md#頁面與身分生命週期) · [未來設計圖](../docs/diagrams.md#未來設計)

## 前端操作與業務契約

依[前端、API 與後端責任及重送範例](../docs/architecture/contracts/business-operations.md#前端操作與業務契約)審查操作結果。

## shadcn 任務入口

凡前端 UI 任務，主動先讀 [官方元件保護與 Skill／MCP 工作流程](miniapp/src/shared/ui/AGENTS.md)，再依 [本機參考庫](../docs/development/tooling/shadcn-reference/README.md) 檢查新鮮度、搜尋完整 Components／AI Elements 索引及相關文件，不需使用者逐次指定 shadcn。完整元件索引只由該參考庫生成；MCP 未曝光時依 UI 指引的主機曝光診斷流程處理。
元件選擇、按需新增與 UI 驗收遵循該檔的「需求驅動的選用與驗收」；主動完成當前任務，不以已安裝清單限制候選，也不擴大為未要求的全站改版。

涉及 shadcn/ui、components.json、registry、preset 或其 MCP 的任務，開始前必讀 [.agents/skills/shadcn/SKILL.md](../.agents/skills/shadcn/SKILL.md)，再按需讀其 references；使用方式與驗證見 [shadcn 工具指南](../docs/development/tooling/mcp/shadcn.md)。所有角色都適用。

涉及統計、報表或儀表板時，先查 [shadcn 官方圖表範例](../docs/development/tooling/mcp/shadcn.md#官方圖表範例)，按資料呈現需求選用圖表並重用共用 Chart 元件。
上游 skill 保留原樣；其 `npx shadcn@latest` 等範例在本專案一律改用 `bash scripts/tooling/tool-exec.sh shadcn`，並將工作目錄指向 `frontend/miniapp`。先讀實際元件文件、檢查已安裝元件，更新前使用 `--dry-run`／`--diff`；不因 skill 範例自動重建專案或安裝所有元件。

## shadcn 元件保護

Codex 預設不得直接修改 `frontend/miniapp/src/shared/ui/*.tsx`、`shared/hooks/use-mobile.ts` 或 `src/shell/sidebar-07/*.tsx` 的官方樣式、結構與互動。不得為配合單一頁面需求更改 primitives，亦不得用全域 CSS 覆蓋其外觀。
業務需求優先透過公開 props、variants、資料與事件，由 `shell/AppFrame.tsx`、`shell/Shell.tsx` 或業務層組合；sidebar-07 保留既有真實導覽、身分與登入操作，不換回示範帳號、假連結或訂閱／帳單功能。
只有使用者明確要求修改、還原或更新上述元件時才可更動，且授權只適用當次任務；先用固定版 CLI `add --dry-run`／`--diff` 核對，再最小修改並執行前端檢查。不要自行執行 `add sidebar-07 --overwrite`：它會產生另一套 shared 元件及 dashboard，無法直接更新目前 shell 中的整合。
必要的 alias、client directive 與已記錄的可及性修正保留。不得藉套用 upstream 移除 `Progress` 傳給 Root 的 `value`，使進度值無法被輔助工具讀取。上游 skill 保留原樣，本專案保護規則優先於其一般客製化建議。

新增、修改或審查業務模型、API 契約、資源路徑、collection 或查詢欄位前，所有角色必讀 [命名規範](../docs/standards/naming.md#業務實體集合與-api-命名)，遵守單一業務實體用單數、集合與 REST 資源路徑用複數的規則及既有相容例外。完整規範只維護於該文件；不得直接重新命名既有公開契約或保存資料。
