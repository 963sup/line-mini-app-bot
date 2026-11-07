# Project documentation

These documents describe the development workflow and intended application.
Application design is a target, not evidence that Go code or deployment exists.
Read relevant pages on demand; Codex does not automatically load all docs.

## Codex 首次接手

先讀 [接手索引與完整度](architecture/codex-readiness.md)，再檢查 [實作契約與待決策](architecture/implementation-contract.md)。
符號與檔案依 [命名規則](development/naming.md)；入口依 [HTTP 契約](architecture/http-contracts.md)，資料依 [資料契約](architecture/data-contracts.md)，工作依 [狀態機](architecture/state-machines.md)。

## 快速入口

| 你現在要做什麼？ | 從這裡開始 |
| --- | --- |
| 理解 LINE／MINI App／Go／Firestore 如何協作 | [應用架構](architecture/overview.md) |
| 決定模組、介面與業務規則放哪裡 | [開發原則與決策](architecture/principles.md) |
| 開始寫 Go | [環境設定](development/setup.md) → [Go 慣例](development/go.md) → [測試](development/testing.md) |

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

- [Agent workflow](agents/workflow.md) and [acceptance](agents/acceptance.md)
- [mise](tooling/mise.md) and [MCP services](tooling/mcp/README.md)
- [Environment setup](development/setup.md), [Go conventions](development/go.md), [tests](development/testing.md)
- [應用架構與流程](architecture/overview.md)、[開發原則與決策](architecture/principles.md)、[決策紀錄](architecture/decisions/README.md)
- [LINE 能力矩陣：一對一、群組、Rich Menu、Flex、MINI App](integrations/line.md) and [Firestore](integrations/firestore.md)
- [Fly.io 部署](operations/fly.md)、[部署 ADR](architecture/decisions/0001-fly-deployment.md) and [troubleshooting](operations/troubleshooting.md)
- [Fly.io deployment from main](operations/fly.md)

Configuration sources: `.codex/config.toml` owns roles/MCP connections;
`mise.toml` owns tool versions, non-secret environment and tasks;
`scripts/` implements tasks. Do not duplicate secrets or version pins in docs.

LINE 設定命名與機密注入：[環境變數規格](integrations/line-environment.md)。

瀏覽器定位：[Web Geolocation API](integrations/web-geolocation.md)（安全 context、同意、Go 驗證、實機驗收）。

## 強制查證規則（全儲存庫、所有角色）

LINE 問題先查 [官方文件來源庫](https://github.com/line/line-developers-docs-source) 的最新版本；其他技術問題或任何實作不確定一律先查 Context7。
不可用記憶代替查證；來源不可用或不足時明確回報，受影響決策保持待確認。完整規則見根目錄 AGENTS.md。

專案技能清單與來源：[Skills](tooling/skills.md)。
