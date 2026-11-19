# api 工作指引

適用於本目錄及其子目錄；先遵循 [根目錄 AGENTS.md](../AGENTS.md)，再套用較近的子目錄補充。使用者及主機指令優先；本檔不增加外部寫入授權。

## 巢狀指引

進入子目錄前讀取對應指引：[contract](contract/AGENTS.md)、[client](client/AGENTS.md)、[codegen](codegen/AGENTS.md)。contract 下的能力目錄再繼承 contract/AGENTS.md；各層只補充自己的責任，不複製上層規則。完整入口見 [目錄導覽](README.md#巢狀指引)。既有 compatibility/ 繼續遵循本指引，不刪除或任意刷新 baseline。

## 職責與邊界

操作契約前必讀 [業務實體、集合與 API 命名](../docs/development/naming.md#業務實體集合與-api-命名)。模型型別、集合、查詢與路徑按其語意命名；保留既有單例及公開契約相容例外，不手改生成名稱。

contract/openapi.yaml 僅組合公開契約；能力歸屬與引用規則遵循 [契約能力索引](contract/README.md)。各能力維護自己的操作與 DTO，共用錯誤／認證放 common/，能力檔不得反向引用根入口。client/、codegen/、compatibility/ 維持共用工具責任，不按能力複製。

先修改契約，再由根目錄 pnpm api:generate 更新 Go/TS 產生碼；不手改 generated 檔或任意刷新 baseline。業務核心不依賴 HTTP DTO。

## 驗證

從根目錄執行 bash scripts/tooling/tool-exec.sh pnpm api:check。歷史比較依 README.md 指定 base revision；缺少 base 不宣稱通過。

命令除另有標示外均在儲存庫根目錄執行。閱讀 [本目錄 README](README.md) 與 [文件索引](../docs/README.md)；回報實際執行結果與限制。

## 文件閱讀路由

先由 [本模組相關文件](README.md#相關文件) 與 [模組閱讀入口](../docs/README.md#按模組閱讀) 選擇文件。局部用途與命令維護在本模組 README；跨模組架構、功能與營運文件依 docs 主題分類保存。索引更新依[文件維護指引](../docs/AGENTS.md#主題子目錄)。

## Business Operation

開始相關任務前，必須閱讀 [操作模型](../docs/foundation/architecture/business-operation.md) 與 [符合性流程](../docs/agents/workflow.md#business-operation-compliance)。

操作契約須對應 OpenAPI 的輸入、回應及既有公開失敗語義；客戶端不能自行宣告可信 Actor、角色或審核權。先改契約再生成，不手改產生碼，不為模型強迫新增端點或破壞相容性。

[現行實作圖：基礎入口](../docs/foundation/api/http-contracts.md#基礎入口) · [未來設計圖](../docs/diagrams.md#未來設計)

## 前端操作與業務契約

Web 業務操作遵循 **Frontend → API → Backend**。每個 API 操作以一個可驗證的業務結果為單位，優先沿用能完整表達該結果的既有契約。

| 層級 | 責任 |
| --- | --- |
| Frontend | 收集輸入，處理瀏覽器互動與必要的本地狀態，透過產生的 API client 送出操作，依回應呈現載入、成功、拒絕或待確認狀態。 |
| API | 在 OpenAPI 定義操作的輸入、成功結果與可辨識的錯誤。對可能重送或併發的寫入，明定回應遺失後的重試方式與版本衝突語意。契約變更後重新產生 Go 與 TypeScript 程式碼。 |
| Backend | Go 依可信身分與目前資料執行授權、權威驗證、業務計算與狀態轉換，維護資料一致性，回傳操作結果。 |

審查時，若前端必須掃描清單或拼接多個查詢才能推斷操作是否成功，應回頭檢查契約。例如「加入專案引用」的回應遺失時，UI 應顯示結果待確認，依契約重試或查詢原操作，由 Go 回覆結果。清單反映的是目前狀態，可能已被後續操作改變。

新增、修改或審查業務模型、API 契約、資源路徑、collection 或查詢欄位前，所有角色必讀 [命名規範](../docs/development/naming.md#業務實體集合與-api-命名)，遵守單一業務實體用單數、集合與 REST 資源路徑用複數的規則及既有相容例外。完整規範只維護於該文件；不得直接重新命名既有公開契約或保存資料。
