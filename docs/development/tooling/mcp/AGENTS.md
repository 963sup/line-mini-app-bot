# MCP 與工具查證指引

## 用途與範圍

保存各 MCP 工具的連線、用途與驗證說明。

## 架構與責任邊界

文件描述能力與診斷方式，實際連線配置由專案的 Codex 設定維護。

## 閱讀與修改入口

先查閱[graphify.md](graphify.md)、[repomix.md](repomix.md)、[context7.md](context7.md)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[工具文件指引](../AGENTS.md) · [本目錄索引](README.md)


遵循[工具文件指引](../AGENTS.md)。各工具的操作入口見[本目錄索引](README.md)。

## 官方來源

- For any LINE question, issue or implementation uncertainty, consult the current official source at https://github.com/line/line-developers-docs-source first. Verify its current default branch/revision; do not treat an old snapshot as current. If coverage is incomplete, follow the official LINE reference linked there and state the limitation. Do not introduce deprecated features or historical compatibility flows into new work.
- For Google Cloud questions, use the installed google/skills skills and Google Developer Knowledge MCP first. If unavailable, use current official Google sources and report the lookup limitation. Validate exact gcloud syntax with leaf-level help before executing commands.

Codex 問題先查 OpenAI Docs MCP。SDK 與版本查證依[全庫查證規則](../../../../AGENTS.md#mandatory-documentation-verification)。

## gcloud sign-in routing

For gcloud login, follow [the short sign-in workflow](google.md#gcloud-sign-in). Keep one orchestrator executor for the whole OAuth process; the top-level agent relays the clickable URL and user code to that same executor. Reuse valid credentials first; otherwise provide the headless login URL immediately. Do not start browser/desktop discovery or additional specialists unless the user asks. Keep CLI login separate from ADC and resource changes.
