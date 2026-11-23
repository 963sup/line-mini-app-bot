# 實作與審查

已確認工作範圍後，從使用者入口追到適用契約與目前程式。修改前讀取目標路徑上的巢狀指引，按需要選用[操作配方](../recipes/README.md)。

核對既有呼叫者及資料語意，實作可從正常入口驗收的完整變更。交付前按[驗收矩陣](../tooling/testing/commands.md)選檢查，並記錄[證據限制](../verification/evidence.md)。

## 完整實作就緒條件

核對本次操作的 API、schema、狀態機與驗收案例，逐一確認角色及各分支的結果。涉及[平台交付提案的 P01–P07](../../proposals/platform/delivery.md#產品決策閘門)時，只追蹤本功能依賴的決策，現行行為以[領域規格](../../domains/README.md)為準。

若要新增到場判定或定位歷史，先依[定位契約](../../integrations/geolocation/browser-api.md)確認 P08 的目的、精度與新鮮度、替代流程、保存期及同意。未決前維持既有功能範圍。

## Deployment and browser location routing

部署與瀏覽器定位任務分別依[部署指引](../../operations/AGENTS.md)與[定位指引](../../../frontend/miniapp/src/shared/location/AGENTS.md)載入適用決策與操作文件；實際驗收依 [acceptance](../verification/evidence.md#cloud-run-and-browser-location)。本節不另外維護部署政策。

## Mandatory documentation verification

All roles MUST follow the [root source-verification policy](../../../AGENTS.md#mandatory-documentation-verification)
and the [Markdown policy](../../standards/documentation.md). Reuse applicable
evidence; report unavailable sources and leave affected decisions pending.
Documentation evidence does not replace implementation checks or authorize
additional writes.

## Business Operation compliance

業務開發、審查及修正必須先讀 [操作模型](../../contracts/business-operations.md)。任務分派附上操作契約、規格來源、既有實作、檔案所有權與驗收條件。

## 實作與自我審查順序

1. 先唯讀定位入口、呼叫者與操作契約，核對當前原始碼；圖譜及來源包的使用方式依[工具指引](../tooling/agents/README.md)。
2. 區分缺陷、工具盲點與待確認需求；記錄規格、根因、影響及驗收條件，只修改本次授權範圍並保留既有變更。
3. 從正常入口完成使用案例及各分支；遇到不明行為依[除錯流程](debugging.md)定位。
4. 唯讀自我審查契約與實作，再依[驗收矩陣](../tooling/testing/commands.md)執行相關檢查。自我審查不等於獨立審查，不把工具成功等同業務正確，也不無限重跑全庫檢查。
5. 程式修改後依[根指引](../../../AGENTS.md)更新 Graphify 與受影響來源包；沿用仍有效的[既有證據](handoff.md#minimal-handoff-and-evidence-reuse)。
6. 依[六欄交接](handoff.md#minimal-handoff-and-evidence-reuse)交付發現、規格依據、修改位置、實際檢查與剩餘限制；範圍外事項另列。

目前代理直接完成上述工作，不啟動子代理。此流程不增加 push、部署或資料遷移授權。

[開發流程索引](README.md)
