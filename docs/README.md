# Project documentation

These documents describe the development workflow and intended application.
Application design is a target, not evidence that Go code or deployment exists.
Read relevant pages on demand; Codex does not automatically load all docs.

目前階段：先穩定基礎設定與應用骨架，部署僅使用 Fly.io。
首版僅做 [基礎出勤](features/attendance.md)：上班、下班、本人紀錄。
先完成 [基礎設施與分階段落地](development/foundation.md)，再接入 LINE／Firestore，最後開發出勤功能。
架構依 [已接受的基線](architecture/decisions/0001-baseline.md)；目前不新增測試方案、
測試案例或測試工具。既有測試文件留待後續階段參考，不列為本輪補文件工作。

## 快速入口

| 你現在要做什麼？ | 從這裡開始 |
| --- | --- |
| 穩定基礎、確認開發先後順序 | [基礎設施與分階段落地](development/foundation.md) |
| 確認首版出勤範圍與待定規則 | [基礎出勤](features/attendance.md) |
| 理解 LINE／MINI App／Go／Firestore 如何協作 | [應用架構](architecture/overview.md) |
| 決定模組、介面與業務規則放哪裡 | [開發原則與決策](architecture/principles.md) |
| 開始寫 Go | [環境設定](development/setup.md) → [Go 慣例](development/go.md)；[測試](development/testing.md)留待後續階段 |

- [Codex Cloud 設定與 MCP 驗收](development/codex-cloud.md)

## 文件索引

- [Agent workflow](agents/workflow.md) and [acceptance](agents/acceptance.md)
- [mise](tooling/mise.md) and [MCP services](tooling/mcp/README.md)
- [Environment setup](development/setup.md), [Go conventions](development/go.md), [tests](development/testing.md)
- [應用架構與流程](architecture/overview.md)、[開發原則與決策](architecture/principles.md)、[決策紀錄](architecture/decisions/README.md)
- [LINE 能力矩陣：一對一、群組、Rich Menu、Flex、MINI App](integrations/line.md) and [Firestore](integrations/firestore.md)
- [Troubleshooting](operations/troubleshooting.md)
- [Fly.io deployment from main](operations/fly.md)

Configuration sources: `.codex/config.toml` owns roles/MCP connections;
`mise.toml` owns tool versions, non-secret environment and tasks;
`scripts/` implements tasks. Do not duplicate secrets or version pins in docs.

LINE 設定命名與機密注入：[環境變數規格](integrations/line-environment.md)。
