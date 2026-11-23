# backend 工作指引

## 閱讀與修改入口

先查閱[package.json](package.json)、[turbo.json](turbo.json)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[文件索引](../docs/README.md) · [模組閱讀入口](../docs/README.md#按模組閱讀)

適用於本目錄及其子目錄；先遵循 [根目錄 AGENTS.md](../AGENTS.md)，再套用較近的子目錄補充。使用者及主機指令優先；本檔不增加外部寫入授權。

## 職責與邊界

唯一可執行入口為 [main.go](main.go)。[business](internal/business) 依業務領域分工，[bootstrap](internal/bootstrap) 組裝服務，[identity](internal/identity) 處理身分與 Session，[httpapi](internal/transport/http/api) 保存 OpenAPI 契約接線與生成碼。

業務核心、HTTP/LINE adapter 與 Firestore adapter 分層；共享組裝在 internal/bootstrap。修改 HTTP 格式先走 ../api 契約流程，不手改產生碼。憑證不得寫入 .env.example。

## Domain & Reliability Gate

適用於 Go 後端的新增、修改與重構。依既有領域指引實作，沿用下方驗證流程。

### 業務與架構

- MUST 先釐清業務目標、所屬 Bounded Context、狀態、資料關係及現有呼叫鏈。各領域擁有自己的資料與狀態，跨領域透過明確契約協作，禁止循環依賴。
- MUST 由所屬領域掌握業務規則、權威計算與狀態轉換，由使用案例協調授權與一致性。從可信 Session 或已驗證的 LINE 身分取得 Actor，按資源、動作、政策及目前狀態決策。MUST NOT 信任客戶端宣告的角色或權限。
- MUST 維持 Domain、Use Case、Port、Adapter 與 Bootstrap 的依賴方向。HTTP DTO、Firestore 儲存格式、LINE SDK 與框架型別留在外層。Handler 與 Adapter 不得重算領域結果或另訂授權規則。
- MUST 對可能併發、重送或部分成功的操作定義交易邊界、冪等條件及失敗語意。變更公開行為或持久化格式前，核對消費端、歷史文件的讀寫相容性及遷移風險。

### 最小化與驗收

- SHOULD 優先刪除、搬移、合併或復用既有 Go 實作。MUST NOT 僅為單一實作、假設需求或形式一致新增介面、Service、Factory、Manager、Registry 或通用引擎，也不得為共用而合併不同領域規則。
- MUST 從原始症狀追查呼叫端，集中修正根因。涉及 HTTP 格式時依 [API 工作流程](../api/README.md)修改契約並重新生成。
- MUST 依下方驗證要求覆蓋成功、拒絕、失敗及關鍵邊界。涉及 Firestore 交易、併發、重送或持久資料相容性時，執行[既有 Emulator 整合驗收](../docs/development/tooling/testing/commands.md#firestore-integration)。

**完成條件**為業務權威唯一、依賴方向正確、授權與一致性經驗證，且適用檢查通過。驗收失敗時繼續修正，受阻時列明無法驗證的情境。

## 驗證

修改套件行為時，檢查呼叫端與受影響的 Go 測試。

使用根目錄 scripts/tooling/tool-exec.sh；Go 修改依[變更驗收矩陣](../docs/development/tooling/testing/commands.md#變更驗收矩陣)執行 fmt-check、lint、check、build 與相關既有測試。純文件修改不冒稱應用測試已執行。

命令除另有標示外均在儲存庫根目錄執行。閱讀 [本目錄 README](README.md) 與 [文件索引](../docs/README.md)；回報實際執行結果與限制。

## 文件閱讀路由

先由 [本模組相關文件](README.md#相關文件) 與 [模組閱讀入口](../docs/README.md#按模組閱讀) 選擇文件。局部用途與命令維護在本模組 README；跨模組架構、功能與營運文件依 docs 主題分類保存。索引更新依[文件維護指引](../docs/AGENTS.md#主題子目錄)。

## Business Operation

開始相關任務前，必須閱讀 [操作模型](../docs/contracts/business-operations.md) 與 [符合性流程](../docs/development/workflow/implementation.md#business-operation-compliance)。

後端由可信身分取得 Actor，依資源歸屬與範圍檢查 Authority／Policy，使用案例協調領域規則與結果。涉及一致性的授權、版本及狀態檢查與交易協調；定義重送與副作用失敗語義。不建立無實際需求的通用引擎或 access 模組。

[現行實作圖：格式與資料相容](../docs/architecture/runtime/backend-boundaries.md#格式與資料相容)

## 前端操作與業務契約

依[前端、API 與後端責任及重送範例](../docs/contracts/business-operations.md#前端操作與業務契約)審查操作結果。

新增、修改或審查業務模型、API 契約、資源路徑、collection 或查詢欄位前，所有角色必讀 [命名規範](../docs/standards/naming.md#業務實體集合與-api-命名)，遵守單一業務實體用單數、集合與 REST 資源路徑用複數的規則及既有相容例外。完整規範只維護於該文件；不得直接重新命名既有公開契約或保存資料。

後端執行授權與業務規則，各領域保留獨立狀態模型，共用必要機制，避免不必要抽象。 定位資料的授權與驗證依[定位規格](../docs/integrations/geolocation/browser-api.md)。
