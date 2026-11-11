# backend 工作指引

適用於本目錄及其子目錄；先遵循 [根目錄 AGENTS.md](../AGENTS.md)，再套用較近的子目錄補充。使用者及主機指令優先；本檔不增加外部寫入授權。

## 職責與邊界

cmd/ 保存 server、richmenu、firestorecheck 與 attendancemigrate 入口；internal/business/ 按 attendance、summary、weather 及其子功能分工；identity 等共用能力保留於 internal/。httpapi/ 保存契約接線與產生碼。

業務核心、HTTP/LINE adapter 與 Firestore adapter 分層；共享組裝在 internal/server。修改 HTTP 格式先走 ../api 契約流程，不手改產生碼。憑證不得寫入 .env.example。

## 驗證

使用根目錄 scripts/tooling/tool-exec.sh；Go 修改依 ../docs/agents/acceptance.md 執行 fmt-check、lint、check、build 與相關既有測試。純文件修改不冒稱應用測試已執行。

命令除另有標示外均在儲存庫根目錄執行。閱讀 [本目錄 README](README.md) 與 [文件索引](../docs/README.md)；回報實際執行結果與限制。

## 文件閱讀路由

先由 [本模組相關文件](README.md#相關文件) 與 [模組閱讀入口](../docs/README.md#按模組閱讀) 選擇文件。局部用途與命令維護在本模組 README；跨模組架構、功能與營運文件依 docs 主題分類保存。新增或搬移相關文件時，同步更新兩層索引與連結。

## Business Operation

開始相關任務前，必須閱讀 [操作模型](../docs/architecture/business-operation.md) 與 [符合性流程](../docs/agents/workflow.md#business-operation-compliance)。

後端由可信身分取得 Actor，依資源歸屬與範圍檢查 Authority／Policy，使用案例協調領域規則與結果。涉及一致性的授權、版本及狀態檢查與交易協調；定義重送與副作用失敗語義。不建立無實際需求的通用引擎或 access 模組。
