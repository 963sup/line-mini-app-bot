# api 工作指引

適用於本目錄及其子目錄；先遵循 [根目錄 AGENTS.md](../AGENTS.md)，再套用較近的子目錄補充。使用者及主機指令優先；本檔不增加外部寫入授權。

## 職責與邊界

操作契約前必讀 [業務實體、集合與 API 命名](../docs/development/naming.md#業務實體集合與-api-命名)。模型型別、集合、查詢與路徑按其語意命名；保留既有單例及公開契約相容例外，不手改生成名稱。

contract/openapi.yaml 僅組合公開契約；能力歸屬與引用規則遵循 [契約能力索引](contract/README.md)。各能力維護自己的操作與 DTO，共用錯誤／認證放 common/，能力檔不得反向引用根入口。client/、codegen/、compatibility/ 維持共用工具責任，不按能力複製。

先修改契約，再由根目錄 pnpm api:generate 更新 Go/TS 產生碼；不手改 generated 檔或任意刷新 baseline。業務核心不依賴 HTTP DTO。

## 驗證

從根目錄執行 bash scripts/tooling/tool-exec.sh pnpm api:check。歷史比較依 README.md 指定 base revision；缺少 base 不宣稱通過。

命令除另有標示外均在儲存庫根目錄執行。閱讀 [本目錄 README](README.md) 與 [文件索引](../docs/README.md)；回報實際執行結果與限制。

## 文件閱讀路由

先由 [本模組相關文件](README.md#相關文件) 與 [模組閱讀入口](../docs/README.md#按模組閱讀) 選擇文件。局部用途與命令維護在本模組 README；跨模組架構、功能與營運文件依 docs 主題分類保存。新增或搬移相關文件時，同步更新兩層索引與連結。

## Business Operation

開始相關任務前，必須閱讀 [操作模型](../docs/architecture/business-operation.md) 與 [符合性流程](../docs/agents/workflow.md#business-operation-compliance)。

操作契約須對應 OpenAPI 的輸入、回應及既有公開失敗語義；客戶端不能自行宣告可信 Actor、角色或審核權。先改契約再生成，不手改產生碼，不為模型強迫新增端點或破壞相容性。
