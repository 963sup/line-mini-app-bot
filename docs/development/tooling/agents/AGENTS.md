# MCP 與工具查證指引

遵循[工具文件指引](../AGENTS.md)。各工具的操作入口見[本目錄索引](README.md)。

## 官方來源

- For any LINE question, issue or implementation uncertainty, consult the current official source at https://github.com/line/line-developers-docs-source first. Verify its current default branch/revision; do not treat an old snapshot as current. If coverage is incomplete, follow the official LINE reference linked there and state the limitation. Do not introduce deprecated features or historical compatibility flows into new work.
- For Google Cloud questions, use the installed google/skills skills and Google Developer Knowledge MCP first. If unavailable, use current official Google sources and report the lookup limitation. Validate exact gcloud syntax with leaf-level help before executing commands.

Codex 問題先查 OpenAI Docs MCP。SDK 與版本查證依[全庫查證規則](../../../../AGENTS.md#mandatory-documentation-verification)。

## gcloud sign-in routing

For gcloud login, follow [the short sign-in workflow](../shared/gcloud.md#gcloud-sign-in). Keep the current agent as the executor for the whole OAuth process; it presents the clickable URL and returns the user code to the same process. Reuse valid credentials first; otherwise provide the headless login URL immediately. Do not start browser/desktop discovery or additional specialists unless the user asks. Keep CLI login separate from ADC and resource changes.
