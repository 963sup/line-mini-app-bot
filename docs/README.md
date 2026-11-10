# Project documentation

本索引先列目前程式能力；本機實作不代表正式環境已部署或完成渠道驗收。

| 目前能力 | 文件入口 |
| --- | --- |
| 群組發起定位打卡、本人紀錄 | [出勤總覽](features/attendance.md) |
| 月摘要、明細分頁、更正與審核 | [查詢與更正](features/attendance-queries.md) |
| 事件重建、休息與工時計算 | [工時計算](features/worktime.md) |
| 可選位置天氣與群聊摘要 | [天氣](features/weather.md)、[摘要](features/summary.md) |
| 原始碼目錄與模組邊界 | [目前檔案樹](architecture/layout.md)、[依賴規則](architecture/dependencies.md) |

部署目標為 Cloud Run；環境與上線條件見 [部署說明](operations/cloud-run.md)。
完整 direct/group/jobs 等 [未來平台規劃](architecture/future-platform.md) 是參考，按實際需求建立。
開發規範仍依 [基線決策](architecture/decisions/0002-baseline.md)；保留既有檢查，不新增測試工程。

## Codex 首次接手

先讀 [已接受的基線](architecture/decisions/0002-baseline.md) 與 [分階段落地](development/foundation.md)，
再按任務讀 [接手索引與完整度](architecture/codex-readiness.md) 與 [實作契約與待決策](architecture/implementation-contract.md)。
平台擴充契約保留為後續參考，不構成首版必做項目。
符號與檔案依 [命名規則](development/naming.md)；入口依 [HTTP 契約](architecture/http-contracts.md)，資料依 [資料契約](architecture/data-contracts.md)，工作依 [狀態機](architecture/state-machines.md)。

## 快速入口

群組功能：[位置天氣](features/weather.md)、[基礎群聊摘要](features/summary.md)（含模型與任務佇列啟用設定）。

| 你現在要做什麼？ | 從這裡開始 |
| --- | --- |
| 穩定基礎、確認開發先後順序 | [基礎設施與分階段落地](development/foundation.md) |
| 接入 Messaging API／MINI App／Rich Menu | [LINE 接入能力與操作](development/line-prototype.md)；[呈現方式選型](integrations/line-presentation-options.md) |
| 使用 Go + Next.js／React 取得位置 | [Web Geolocation 示範與契約](integrations/web-geolocation.md) |
| 確認首版出勤範圍與待定規則 | [基礎出勤](features/attendance.md) |
| 理解 LINE／MINI App／Go／Firestore 如何協作 | [應用架構](architecture/overview.md) |
| 決定模組、介面與業務規則放哪裡 | [開發原則與決策](architecture/principles.md) |
| 開始寫 Go | [環境設定](development/setup.md) → [Go 慣例](development/go.md)；[測試](development/testing.md)留待後續階段 |

- [Codex Cloud 設定與 MCP 驗收](development/codex-cloud.md)

## Codex 架構閱讀路由

先讀 [架構總覽](architecture/overview.md)，按任務讀下列頁面，無需每次載入全部 docs。

| 任務 | 文件 |
| --- | --- |
| 建立套件、規劃第二／第三層目錄 | [目錄規劃](architecture/layout.md) |
| 分析依賴、Ports & Adapters、圖譜 | [依賴規則](architecture/dependencies.md) |
| Webhook、Firestore、Garr、背景工作 | [可靠性與資料](architecture/reliability.md) |
| 排定實作、測試與完成條件 | [交付計畫](architecture/delivery.md) |

## 文件索引

- [素材與腳本路徑整理](development/path-reorganization.md)

- [出勤查詢、Web 分頁與更正重試](features/attendance-queries.md)
- [Agent workflow](agents/workflow.md) and [acceptance](agents/acceptance.md) and [systematic debugging](agents/debugging.md)
- [mise](tooling/mise.md) and [MCP services](tooling/mcp/README.md)
- [Cloud development environment](development/cloud-environment.md)
- [Environment setup](development/setup.md), [Go conventions](development/go.md), [tests](development/testing.md)
- [應用架構與流程](architecture/overview.md)、[開發原則與決策](architecture/principles.md)、[決策紀錄](architecture/decisions/README.md)
- [LINE 能力矩陣：一對一、群組、Rich Menu、Flex、MINI App](integrations/line.md) and [Firestore](integrations/firestore.md)
- [Cloud Run 部署](operations/cloud-run.md)、[部署 ADR](architecture/decisions/0003-cloud-run-primary.md) and [troubleshooting](operations/troubleshooting.md)

Configuration sources: `.codex/config.toml` owns roles/MCP connections;
`mise.toml` owns tool versions, non-secret environment and tasks;
`scripts/` implements tasks. Do not duplicate secrets or version pins in docs.

LINE 設定命名與機密注入：[環境變數規格](integrations/line-environment.md)。

已提供獨立的 [Web Geolocation 示範](integrations/web-geolocation.md)；未接入首版基礎出勤。

## 強制查證規則（全儲存庫、所有角色）

LINE 問題先查 [官方文件來源庫](https://github.com/line/line-developers-docs-source) 的最新版本；其他技術問題或任何實作不確定一律先查 Context7。
不可用記憶代替查證；來源不可用或不足時明確回報，受影響決策保持待確認。完整規則見根目錄 AGENTS.md。

專案技能清單與來源：[Skills](tooling/skills.md)。

## 目錄導覽與維護

README.md 是文件索引；architecture/、development/、features/、integrations/、operations/、tooling/、agents/ 各自保存決策、開發、功能、整合、營運、工具與代理流程。

以已接受 ADR 與實際來源為準，區分已實作、待驗收與未來規劃。新增文件需連回索引；不在多處複製版本、連線設定、機密或完整政策。

驗證本機連結與來源路徑；工具圖譜不證明語意正確。純文件變更執行 git diff --check，相關設定變更再執行設定檢查。

[本目錄指引](AGENTS.md) · [根目錄說明](../README.md) · [根目錄指引](../AGENTS.md)

- [Graphify／Repomix 交叉比對與巢狀指引依據](tooling/graphify-repomix-crosscheck.md)
