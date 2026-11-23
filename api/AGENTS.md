# api 工作指引

目錄分工、生成與驗證命令見[API 工作流程](README.md)。進入 [contract](contract/AGENTS.md)、[codegen](codegen/AGENTS.md) 或 [client](client/AGENTS.md) 修改檔案前，先讀該目錄的工作指引。

## 相關規格

[業務實體、集合與 API 命名](../docs/standards/naming.md#業務實體集合與-api-命名) · [文件索引](../docs/README.md)

適用於本目錄及其子目錄；先遵循 [根目錄 AGENTS.md](../AGENTS.md)，再套用較近的子目錄補充。使用者及主機指令優先；本檔不增加外部寫入授權。

## 巢狀指引

進入子目錄前讀取對應指引：[contract](contract/AGENTS.md)、[client](client/AGENTS.md)、[codegen](codegen/AGENTS.md)。contract 下的能力目錄再繼承 contract/AGENTS.md；各層只補充自己的責任，不複製上層規則。完整入口見 [目錄導覽](README.md#巢狀指引)。

## 職責與邊界

操作契約前必讀 [業務實體、集合與 API 命名](../docs/standards/naming.md#業務實體集合與-api-命名)。模型型別、集合、查詢與路徑按其語意命名；保留既有單例及公開契約相容例外，不手改生成名稱。

contract/openapi.yaml 僅組合公開契約；能力歸屬與引用規則遵循 [契約能力索引](contract/README.md)。各能力維護自己的操作與 DTO，共用錯誤／認證放 common/，能力檔不得反向引用根入口。client/ 與 codegen/ 維持共用工具責任，不按能力複製。

先修改契約，再由根目錄 pnpm api:generate 更新 Go/TS 產生碼；不手改 generated 檔。業務核心不依賴 HTTP DTO。

## Contract & Compatibility Gate

適用於公開與內部 HTTP 契約、DTO、Client 及生成流程。依 contract、client 與 codegen 的巢狀指引執行。

### 契約與相容性

- MUST 以 OpenAPI 為 HTTP 契約的唯一權威來源。先在 [capabilities.json](contract/capabilities.json) 核對能力歸屬、資源語意、資料所有權及消費端，區分公開 HTTP、內部 Worker HTTP 與非 HTTP 事件。
- MUST 定義操作的 Request、Response、狀態碼、認證、錯誤與必要的重送語意，沿用共用錯誤及認證定義。API 描述可觀察的行為，業務規則、狀態轉換與授權決策由 Backend 執行。客戶端傳入的 Actor、角色、權限或判斷結果不得作為可信授權依據。
- MUST 維持公開路徑、欄位、型別、錯誤及行為相容。破壞性變更須先確認需求、受影響呼叫端與遷移策略。

### 最小化與驗收

- SHOULD 優先擴充語意相符的操作，回傳消費端需要的資料，減少傳輸與重複組裝。MUST NOT 為單一畫面破壞資源語意或其他消費端相容性，也不得強行共用語意不同的 Schema。
- MUST NOT 為單一欄位或假設需求堆疊 DTO、Wrapper、Schema 或轉換器。沿用現有生成設定與共用 Client，避免重複端點、平行契約或額外 Generator。
- MUST 先確認 Backend 業務行為，再依 Contract → Generate → Backend Adapter → Frontend Consumer 同步。檢視 Go、TypeScript 產生碼差異與所有受影響消費端，禁止手改生成檔或複製傳輸型別。
- MUST 依下方驗證流程執行 `pnpm api:check`，核對操作歸屬、Schema 引用及生成結果，排除循環引用、無效定義與契約漂移。歷史相容性須另指定 base revision 比較。對變更操作驗證成功、拒絕、錯誤及適用的重送情境。

**完成條件**為操作語意明確、契約與兩端實作一致、無重複定義或無效端點，且適用相容性檢查通過。無法驗證時列出缺口，禁止放寬契約或更換比較基準來通過驗收。

## 驗證

從根目錄執行 bash scripts/tooling/tool-exec.sh pnpm api:check。歷史相容性比較須明列 base revision 及結果；檢查範圍見 [API 工作流程](README.md#目錄導覽與維護)。

命令除另有標示外均在儲存庫根目錄執行。閱讀 [本目錄 README](README.md) 與 [文件索引](../docs/README.md)；回報實際執行結果與限制。

## 文件閱讀路由

先由 [本模組相關文件](README.md#相關文件) 與 [模組閱讀入口](../docs/README.md#按模組閱讀) 選擇文件。局部用途與命令維護在本模組 README；跨模組架構、功能與營運文件依 docs 主題分類保存。索引更新依[文件維護指引](../docs/AGENTS.md#主題子目錄)。

## Business Operation

開始相關任務前，必須閱讀 [操作模型](../docs/contracts/business-operations.md) 與 [符合性流程](../docs/development/workflow/implementation.md#business-operation-compliance)。

操作契約須對應 OpenAPI 的輸入、回應及既有公開失敗語義；客戶端不能自行宣告可信 Actor、角色或審核權。先改契約再生成，不手改產生碼，不為模型強迫新增端點或破壞相容性。

[現行實作圖：基礎入口](../docs/contracts/http.md#基礎入口)

## 前端操作與業務契約

依[前端、API 與後端責任及重送範例](../docs/contracts/business-operations.md#前端操作與業務契約)審查操作結果。

新增、修改或審查業務模型、API 契約、資源路徑、collection 或查詢欄位前，所有角色必讀 [命名規範](../docs/standards/naming.md#業務實體集合與-api-命名)，遵守單一業務實體用單數、集合與 REST 資源路徑用複數的規則及既有相容例外。完整規範只維護於該文件；不得直接重新命名既有公開契約或保存資料。
