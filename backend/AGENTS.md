# backend 工作指引

## 用途與範圍

保存 Go 後端指令、套件與模組設定。

## 架構與責任邊界

依目錄用途提供後端入口或共用能力；服務接線由 server 負責，資源授權與使用案例歸各業務領域。

## 閱讀與修改入口

先查閱[package.json](package.json)、[turbo.json](turbo.json)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[文件索引](../docs/README.md) · [模組閱讀入口](../docs/README.md#按模組閱讀)


適用於本目錄及其子目錄；先遵循 [根目錄 AGENTS.md](../AGENTS.md)，再套用較近的子目錄補充。使用者及主機指令優先；本檔不增加外部寫入授權。

## 職責與邊界

cmd/ 保存 server、richmenu、firestorecheck 入口；internal/business/ 按 attendance、summary、weather 及其子功能分工；identity 等共用能力保留於 internal/。httpapi/ 保存契約接線與產生碼。

業務核心、HTTP/LINE adapter 與 Firestore adapter 分層；共享組裝在 internal/server。修改 HTTP 格式先走 ../api 契約流程，不手改產生碼。憑證不得寫入 .env.example。

## 驗證

使用根目錄 scripts/tooling/tool-exec.sh；Go 修改依 ../docs/development/acceptance.md 執行 fmt-check、lint、check、build 與相關既有測試。純文件修改不冒稱應用測試已執行。

命令除另有標示外均在儲存庫根目錄執行。閱讀 [本目錄 README](README.md) 與 [文件索引](../docs/README.md)；回報實際執行結果與限制。

## 文件閱讀路由

先由 [本模組相關文件](README.md#相關文件) 與 [模組閱讀入口](../docs/README.md#按模組閱讀) 選擇文件。局部用途與命令維護在本模組 README；跨模組架構、功能與營運文件依 docs 主題分類保存。索引更新依[文件維護指引](../docs/AGENTS.md#主題子目錄)。

## Business Operation

開始相關任務前，必須閱讀 [操作模型](../docs/architecture/contracts/business-operations.md) 與 [符合性流程](../docs/development/workflow.md#business-operation-compliance)。

後端由可信身分取得 Actor，依資源歸屬與範圍檢查 Authority／Policy，使用案例協調領域規則與結果。涉及一致性的授權、版本及狀態檢查與交易協調；定義重送與副作用失敗語義。不建立無實際需求的通用引擎或 access 模組。

[現行實作圖：格式與資料相容](../docs/architecture/runtime/backend-boundaries.md#格式與資料相容) · [未來設計圖](../docs/diagrams.md#未來設計)

## 前端操作與業務契約

Web 業務操作遵循 **Frontend → API → Backend**。每個 API 操作以一個可驗證的業務結果為單位，優先沿用能完整表達該結果的既有契約。

| 層級 | 責任 |
| --- | --- |
| Frontend | 收集輸入，處理瀏覽器互動與必要的本地狀態，透過產生的 API client 送出操作，依回應呈現載入、成功、拒絕或待確認狀態。 |
| API | 在 OpenAPI 定義操作的輸入、成功結果與可辨識的錯誤。對可能重送或併發的寫入，明定回應遺失後的重試方式與版本衝突語意。契約變更後重新產生 Go 與 TypeScript 程式碼。 |
| Backend | Go 依可信身分與目前資料執行授權、權威驗證、業務計算與狀態轉換，維護資料一致性，回傳操作結果。 |

審查時，若前端必須掃描清單或拼接多個查詢才能推斷操作是否成功，應回頭檢查契約。例如「加入專案引用」的回應遺失時，UI 應顯示結果待確認，依契約重試或查詢原操作，由 Go 回覆結果。清單反映的是目前狀態，可能已被後續操作改變。

新增、修改或審查業務模型、API 契約、資源路徑、collection 或查詢欄位前，所有角色必讀 [命名規範](../docs/standards/naming.md#業務實體集合與-api-命名)，遵守單一業務實體用單數、集合與 REST 資源路徑用複數的規則及既有相容例外。完整規範只維護於該文件；不得直接重新命名既有公開契約或保存資料。

後端執行授權與業務規則，各領域保留獨立狀態模型，共用必要機制，避免不必要抽象。 定位資料的授權與驗證依[定位規格](../docs/architecture/integrations/web-geolocation.md)。
