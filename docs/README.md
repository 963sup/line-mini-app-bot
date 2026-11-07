# Project documentation

These documents describe the development workflow and intended application.
Application design is a target, not evidence that Go code or deployment exists.
Read relevant pages on demand; Codex does not automatically load all docs.

## 快速入口

| 你現在要做什麼？ | 從這裡開始 |
| --- | --- |
| 理解 LINE／MINI App／Go／Firestore 如何協作 | [應用架構](architecture/overview.md) |
| 決定模組、介面與業務規則放哪裡 | [開發原則與決策](architecture/principles.md) |
| 開始寫 Go | [環境設定](development/setup.md) → [Go 慣例](development/go.md) → [測試](development/testing.md) |

## 文件索引

- [Agent workflow](agents/workflow.md) and [acceptance](agents/acceptance.md)
- [mise](tooling/mise.md) and [MCP services](tooling/mcp/README.md)
- [Environment setup](development/setup.md), [Go conventions](development/go.md), [tests](development/testing.md)
- [應用架構與流程](architecture/overview.md)、[開發原則與決策](architecture/principles.md)、[決策紀錄](architecture/decisions/README.md)
- [LINE 能力矩陣：一對一、群組、Rich Menu、Flex、MINI App](integrations/line.md) and [Firestore](integrations/firestore.md)
- [Cloud Run](operations/cloud-run.md) and [troubleshooting](operations/troubleshooting.md)
- [Fly.io deployment from main](operations/fly.md)

Configuration sources: `.codex/config.toml` owns roles/MCP connections;
`mise.toml` owns tool versions, non-secret environment and tasks;
`scripts/` implements tasks. Do not duplicate secrets or version pins in docs.

LINE 設定命名與機密注入：[環境變數規格](integrations/line-environment.md)。
