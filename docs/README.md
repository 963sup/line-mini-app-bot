# 專案文件索引

依工作模組選擇閱讀入口；文件維持按架構、功能、開發、整合與營運分類。
模組 README 保存局部命令與維護方式，跨模組規格集中於 docs，透過連結共用。
本機實作不代表正式環境已部署或完成渠道驗收。

## 按模組閱讀

| 工作模組 | 目錄入口 | 相關文件 |
| --- | --- | --- |
| Backend | [Go 後端](../backend/README.md) | [功能邊界](architecture/backend-boundaries.md)、[依賴規則](architecture/dependencies.md)、[Go 慣例](development/go.md)、[可靠性與資料](architecture/reliability.md) |
| API | [HTTP API 契約與生成流程](../api/README.md) | [HTTP 契約](architecture/http-contracts.md)、[資料契約](architecture/data-contracts.md)、[查詢與更正](features/attendance-queries.md) |
| Frontend | [MINI App 前端](../frontend/README.md) | [React 架構與生命週期](architecture/react-frontend.md)、[Web Geolocation](integrations/web-geolocation.md)、[查詢與更正](features/attendance-queries.md) |
| Scripts | [腳本分類與命令](../scripts/README.md) | [mise](tooling/mise.md)、[分析範圍](tooling/analysis-scope.md)、[路徑整理](development/path-reorganization.md) |
| Deploy | [部署設定與索引](../deploy/README.md) | [Cloud Run](operations/cloud-run.md)、[發布驗收](operations/release-acceptance.md)、[故障排查](operations/troubleshooting.md) |

## 功能與跨模組架構

| 需求 | 文件 |
| --- | --- |
| 群組請假入口與本人登記紀錄 | [群組請假登記](features/leave.md) |
| 群組人數看板、本人漏下班檢查 | [出勤看板與漏下班](features/attendance-board.md) |
| 群組定位打卡、本人紀錄 | [出勤總覽](features/attendance.md) |
| 月摘要、明細分頁、更正與審核 | [查詢與更正](features/attendance-queries.md) |
| 事件重建、休息與工時計算 | [工時計算](features/worktime.md) |
| 可選位置天氣、群聊摘要 | [天氣](features/weather.md)、[摘要](features/summary.md) |
| 理解各模組如何協作 | [架構總覽](architecture/overview.md)、[目前目錄](architecture/layout.md)、[開發原則](architecture/principles.md) |
| HTTP、資料與工作生命週期 | [HTTP 契約](architecture/http-contracts.md)、[資料契約](architecture/data-contracts.md)、[狀態機](architecture/state-machines.md) |
| 交付順序與架構決策 | [交付計畫](architecture/delivery.md)、[ADR 索引](architecture/decisions/README.md)、[Monorepo 決策](architecture/decisions/0004-monorepo.md) |

## 開發與接手

先讀 [已接受的基線](architecture/decisions/0002-baseline.md) 與 [分階段落地](development/foundation.md)，
再按任務閱讀 [接手索引](architecture/codex-readiness.md)、[實作契約](architecture/implementation-contract.md) 與 [命名規則](development/naming.md)。
[未來平台規劃](architecture/future-platform.md) 為後續參考，不構成目前必做範圍。

- 環境：[設定](development/setup.md)、[Cloud 開發環境](development/cloud-environment.md)、[Codex Cloud](development/codex-cloud.md)。
- LINE：[接入試作](development/line-prototype.md)、[能力矩陣](integrations/line.md)、[呈現選型](integrations/line-presentation-options.md)、[環境變數](integrations/line-environment.md)。
- 身分登入：[Go browser Session 與滑動續期](architecture/browser-session.md)。
- 資料整合：[Firestore](integrations/firestore.md)。
- 工作流程：[代理協作](agents/workflow.md)、[驗收](agents/acceptance.md)、[系統化除錯](agents/debugging.md)、[測試階段說明](development/testing.md)。
- 工具：[MCP 索引](tooling/mcp/README.md)、[Skills](tooling/skills.md)、[Graphify／Repomix 交叉比對](tooling/graphify-repomix-crosscheck.md)。

## 強制查證規則（全儲存庫、所有角色）

LINE 問題先查 [官方文件來源庫](https://github.com/line/line-developers-docs-source) 的最新版本；其他技術問題或任何實作不確定一律先查 Context7。
不可用記憶代替查證；來源不可用或不足時明確回報，受影響決策保持待確認。完整規則見根目錄 AGENTS.md。

專案技能清單與來源：[Skills](tooling/skills.md)。

## 文件歸屬與維護

| 位置 | 保存內容 |
| --- | --- |
| [architecture/](architecture/README.md) | 跨模組架構、契約設計與決策 |
| [features/](features/README.md) | 完整功能流程與業務規則 |
| [development/](development/README.md) | 開發環境、慣例與開發流程 |
| [integrations/](integrations/README.md) | 外部服務與瀏覽器能力整合 |
| [operations/](operations/README.md) | 部署、發布驗收與故障排查 |
| [tooling/](tooling/README.md) | 開發工具、分析範圍與工具報告 |
| [agents/](agents/README.md) | 代理協作、驗收與除錯流程 |
| 各模組 `README.md` | 局部用途、命令、修改邊界與相關文件連結 |

新增文件依主題歸屬，並加入本索引與相關模組 README；跨模組內容維持單一來源，避免複製成多份 backend／frontend 規格。
以已接受 ADR 與實際來源為準，區分已實作、待驗收與未來規劃。
`.codex/config.toml` 管理角色與 MCP 連線，`mise.toml` 管理版本、非機密環境與任務，`scripts/` 實作任務；文件不重複設定值或機密。

修改後驗證本機連結、標題錨點與來源路徑，執行 `git diff --check`；相關設定修改另執行設定檢查。
圖譜與打包成功不證明語意正確或整合已驗收。

[本目錄指引](AGENTS.md) · [根目錄說明](../README.md) · [根目錄指引](../AGENTS.md)

- 業務操作：[Business Operation 模型與操作契約](architecture/business-operation.md)、[持續符合性流程](agents/workflow.md#business-operation-compliance)。

- [Business 模組、契約與驗證對照](development/business-refactor.md)
