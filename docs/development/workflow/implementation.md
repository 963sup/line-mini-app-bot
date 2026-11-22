# 實作與審查

已確認工作範圍後，從使用者入口追到適用契約與目前程式。修改前讀取目標路徑上的巢狀指引，按需要選用[操作配方](../recipes/README.md)。

核對既有呼叫者及資料語意，實作可從正常入口驗收的完整變更。交付前按[驗收矩陣](../verification/commands.md)選檢查，並記錄[證據限制](../verification/evidence.md)。

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

**開發與維運指引**

```mermaid
flowchart TB
  A["Graphify query 定位符號，必要時 path／explain"] --> B["Repomix CLI 與原文核對指引、契約、規則及入口"]
  B --> C["依查證路由核對版本與產品規格"]
  C --> D{"必需來源可用？"}
  D -->|否| H["受影響決策待確認，繼續獨立部分"]
  D -->|是| E["區分缺陷、工具盲點與待確認需求"]
  E --> F["記錄規格、符號、影響、最小修正及檢查"]
  F --> G["分派範圍內修正，依驗收矩陣交付"]
```

圖譜可能過期或漏掉動態關係，無節點時核對原始碼。核對 Repomix 打包範圍，Markdown／TOML 等遺漏內容直接讀取並記錄。產物放任務專屬的未提交目錄。

Context7 先解析 library ID 再查適用版本。LINE、Google Cloud 與 Codex 依[工具查證指引](../tooling/mcp/AGENTS.md)。產品決策以使用者與已接受規格為準。

Explorer 與 reviewer 維持唯讀並回報 orchestrator，implementer 修正，必要時由 debugger 定位。保留既有變更，範圍外另列。主機不支援角色時由主代理直接承接並說明。依[驗收矩陣](../verification/commands.md)核對契約、實作與證據。程式修改後依根指引更新 Graphify，重新打包受影響內容並沿用有效證據。

交付列出：發現 → 規格依據 → 修改位置 → 檢查結果 → 剩餘限制。不把工具成功等同業務正確，不無限重跑全庫檢查。此流程不增加 push、部署或資料遷移授權。

沿用已註冊角色即可，不新增 compliance agent。協調者可直接完成小任務；角色是責任分工，不是必須逐一啟動的階段。

[開發流程索引](README.md)
