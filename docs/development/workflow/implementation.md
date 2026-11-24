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

[領域規格](../../domains/)與[業務操作模型](../../contracts/business-operations.md)定義操作及授權；實作須維持 use case、聚合、policy、port、adapter 與交易邊界的一致性。跨層 API 依[契約指引](../tooling/api/change-api.md)同步；邊界概念按需查閱 [DDD 問題索引](../../references/ddd/design-guide.md#按問題查閱)。[驗收矩陣](../tooling/testing/commands.md#變更驗收矩陣)涵蓋適用的授權、重送、交易或併發案例。

完成時指出改動的權威規格、Go 符號、寫入位置與已驗證的結果。若缺少整合證據，用補卡範例的方式寫出仍須核對的交易或併發行為。[補卡審核的規格到程式範例](../../domains/attendance/design.md#補正審核與重送)

## 修改 MINI App 前端

修改頁面、路由、互動、表單或元件時使用。先找出使用者進入的路由、需要的身分和前置資料。

[前端指引](../../../frontend/AGENTS.md)定義 UI 限制；涉及取消、過期回應或登入切換時適用 [React 生命週期](../../architecture/runtime/react-lifecycle.md)。依[驗收矩陣](../tooling/testing/commands.md#變更驗收矩陣)涵蓋相關操作、載入、錯誤、空資料及拒絕情境；服務啟動方式見[環境指南](../environment/codex-cloud.md#本機服務啟動)。

完成時記錄路由、重現步驟、瀏覽器結果及既有檢查。需驗證 LINE 或定位時，分開記錄實際裝置與瀏覽器的結果。[裝置與定位證據](../verification/evidence.md#cloud-run-and-browser-location)

<a id="diagnostic-loop"></a>

## 除錯與根因驗證

重現原始問題並核對目前來源與操作契約；區分觀察、推論及確認的原因。無法重現時，記錄條件與限制，不將推測當作定論。修復後重跑原始案例及[適用檢查](../tooling/testing/commands.md)，回報尚未驗證的行為。

保留其他聊天的變更及敏感資料；不吞錯、不削弱驗證或斷言、不關閉檢查，也不加入無界重試。圖譜與打包內容用於定位，不能證明執行結果。

交接時依[證據重用](handoff.md#minimal-handoff-and-evidence-reuse)記錄重現、因果鏈、排除假設、修改必要性與檢查結果；推送依任務授權。

[開發流程索引](./)
