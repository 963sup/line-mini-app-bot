# api 工作指引

本目錄沿用[根目錄指引](../AGENTS.md)。目錄分工、生成與驗證命令見[API 工作流程](README.md)。

## 相關規格

[業務實體、集合與 API 命名](../docs/standards/naming.md#業務實體集合與-api-命名) · [文件索引](../docs/README.md)

適用於本目錄及其子目錄；先遵循 [根目錄 AGENTS.md](../AGENTS.md)，再套用較近的子目錄補充。使用者及主機指令優先；本檔不增加外部寫入授權。

## 巢狀指引

進入子目錄前讀取對應指引：[contract](contract/AGENTS.md)、[client](client/AGENTS.md)、[codegen](codegen/AGENTS.md)。contract 下的能力目錄再繼承 contract/AGENTS.md；各層只補充自己的責任，不複製上層規則。完整入口見 [目錄導覽](README.md#巢狀指引)。

## 職責與邊界

操作契約前必讀 [業務實體、集合與 API 命名](../docs/standards/naming.md#業務實體集合與-api-命名)。模型型別、集合、查詢與路徑按其語意命名；保留既有單例及公開契約相容例外，不手改生成名稱。

contract/openapi.yaml 僅組合公開契約；能力歸屬與引用規則遵循 [契約能力索引](contract/README.md)。各能力維護自己的操作與 DTO，共用錯誤／認證放 common/，能力檔不得反向引用根入口。client/ 與 codegen/ 維持共用工具責任，不按能力複製。

先修改契約，再由根目錄 pnpm api:generate 更新 Go/TS 產生碼；不手改 generated 檔。業務核心不依賴 HTTP DTO。

## 驗證

從根目錄執行 bash scripts/tooling/tool-exec.sh pnpm api:check。歷史相容性比較須明列 base revision 及結果；檢查範圍見 [API 工作流程](README.md#目錄導覽與維護)。

命令除另有標示外均在儲存庫根目錄執行。閱讀 [本目錄 README](README.md) 與 [文件索引](../docs/README.md)；回報實際執行結果與限制。

## 文件閱讀路由

先由 [本模組相關文件](README.md#相關文件) 與 [模組閱讀入口](../docs/README.md#按模組閱讀) 選擇文件。局部用途與命令維護在本模組 README；跨模組架構、功能與營運文件依 docs 主題分類保存。索引更新依[文件維護指引](../docs/AGENTS.md#主題子目錄)。

## Business Operation

開始相關任務前，必須閱讀 [操作模型](../docs/contracts/business-operations.md) 與 [符合性流程](../docs/development/workflow.md#business-operation-compliance)。

操作契約須對應 OpenAPI 的輸入、回應及既有公開失敗語義；客戶端不能自行宣告可信 Actor、角色或審核權。先改契約再生成，不手改產生碼，不為模型強迫新增端點或破壞相容性。

[現行實作圖：基礎入口](../docs/contracts/http.md#基礎入口)

## 前端操作與業務契約

依[前端、API 與後端責任及重送範例](../docs/contracts/business-operations.md#前端操作與業務契約)審查操作結果。

新增、修改或審查業務模型、API 契約、資源路徑、collection 或查詢欄位前，所有角色必讀 [命名規範](../docs/standards/naming.md#業務實體集合與-api-命名)，遵守單一業務實體用單數、集合與 REST 資源路徑用複數的規則及既有相容例外。完整規範只維護於該文件；不得直接重新命名既有公開契約或保存資料。
