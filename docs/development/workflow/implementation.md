# 實作與審查

已確認工作範圍後，從使用者入口追到適用契約與目前程式。修改前讀取目標路徑上的巢狀指引，依變更選讀[領域行為](#修改領域行為)、[前端操作](#修改-mini-app-前端)或 [API 契約](../tooling/api/change-api.md)。

核對既有呼叫者及資料語意，實作可從正常入口驗收的完整變更。交付前按[驗收矩陣](../tooling/testing/commands.md)選檢查，並記錄[證據限制](../verification/evidence.md)。

## 完整實作就緒條件

核對本次操作的 API、schema、狀態機與驗收案例，逐一確認角色及各分支的結果。涉及[平台交付提案的 P01–P07](../../proposals/platform/delivery.md#產品決策閘門)時，只追蹤本功能依賴的決策，現行行為以[領域規格](../../domains/)為準。

若要新增到場判定或定位歷史，先依[定位契約](../../integrations/geolocation/browser-api.md)確認 P08 的目的、精度與新鮮度、替代流程、保存期及同意。未決前維持既有功能範圍。

## Deployment and browser location routing

部署與瀏覽器定位任務分別依[部署指引](../../operations/cloud-run.md)、[正式環境故障排查](../../operations/troubleshooting.md)與[定位指引](../../../frontend/AGENTS.md)載入適用決策與操作文件；實際驗收依 [acceptance](../verification/evidence.md#cloud-run-and-browser-location)。本節不另外維護部署政策。

## Mandatory documentation verification

All roles MUST follow the [root source-verification policy](../../../AGENTS.md#mandatory-documentation-verification)
and the [Markdown policy](../../standards/documentation.md). Reuse applicable
evidence; report unavailable sources and leave affected decisions pending.
Documentation evidence does not replace implementation checks or authorize
additional writes.

## Business Operation compliance

業務開發、審查及修正必須先讀 [操作模型](../../contracts/business-operations.md)。任務分派附上操作契約、規格來源、既有實作、檔案所有權與驗收條件。

## 修改領域行為

授權、狀態轉換、重送、併發或資料寫入規則有變時使用。先取得需求與預期的成功、拒絕和失敗狀態。

1. 從[領域規格](../../domains/)選功能，依[業務操作模型](../../contracts/business-operations.md)定位 Actor、Resource、Action 與 Authority。
2. 從使用者入口追到 use case、聚合或 policy，再核對 port、adapter 及交易邊界。邊界存在疑問時，從 [DDD 問題索引](../../references/ddd/design-guide.md#按問題查閱)按題選讀。
3. 依[業務符合性流程](#business-operation-compliance)同步規格、判斷與實作。影響 API 時，也走[契約配方](../tooling/api/change-api.md)。
4. 依[驗收矩陣](../tooling/testing/commands.md#變更驗收矩陣)執行既有適用案例，記錄能支持哪些授權、重送、交易或併發結論。

完成時指出改動的權威規格、Go 符號、寫入位置與已驗證的結果。若缺少整合證據，用補卡範例的方式寫出仍須核對的交易或併發行為。[補卡審核的規格到程式範例](../../references/ddd/design-guide.md#把概念用到現行程式補正審核)

## 修改 MINI App 前端

修改頁面、路由、互動、表單或元件時使用。先找出使用者進入的路由、需要的身分和前置資料。

1. 從 [MINI App 目錄導覽](../../../frontend/miniapp/)找頁面、業務畫面及巢狀指引。元件另核對 [UI 指引](../../../frontend/AGENTS.md)。
2. 從畫面讀寫或變更狀態的動作追到 API 及身分來源。涉及取消、過期回應或登入切換時，對照 [React 生命週期](../../architecture/runtime/react-lifecycle.md)。
3. 用[環境啟動表](../environment/codex-cloud.md#本機服務啟動)開啟目標頁面，驗收主要操作與本次變更涉及的載入、錯誤、空資料或拒絕情境。
4. 按[前端驗收矩陣](../tooling/testing/commands.md#變更驗收矩陣)執行既有類型檢查與測試。路由、資源或靜態輸出有改時再建置。

完成時記錄路由、重現步驟、瀏覽器結果及既有檢查。需驗證 LINE 或定位時，分開記錄實際裝置與瀏覽器的結果。[裝置與定位證據](../verification/evidence.md#cloud-run-and-browser-location)

## 實作與自我審查順序

1. 先唯讀定位入口、呼叫者與操作契約，核對當前原始碼；圖譜及來源包的使用方式依[工具指引](../tooling/agents/)。
2. 區分缺陷、工具盲點與待確認需求；記錄規格、根因、影響及驗收條件，只修改本次授權範圍並保留既有變更。
3. 從正常入口完成使用案例及各分支；遇到不明行為依[除錯流程](debugging.md)定位。
4. 唯讀自我審查契約與實作，再依[驗收矩陣](../tooling/testing/commands.md)執行相關檢查。自我審查不等於獨立審查，不把工具成功等同業務正確，也不無限重跑全庫檢查。
5. 程式修改後依[根指引](../../../AGENTS.md)更新 Graphify 與受影響來源包；沿用仍有效的[既有證據](handoff.md#minimal-handoff-and-evidence-reuse)。
6. 依[六欄交接](handoff.md#minimal-handoff-and-evidence-reuse)交付發現、規格依據、修改位置、實際檢查與剩餘限制；範圍外事項另列。

目前代理直接完成上述工作，不啟動子代理。此流程不增加 push、部署或資料遷移授權。

[開發流程索引](./)
