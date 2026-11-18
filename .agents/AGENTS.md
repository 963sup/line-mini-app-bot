# .agents 工作指引

適用於本目錄及其子目錄；先遵循 [根目錄 AGENTS.md](../AGENTS.md)，再套用較近的子目錄補充。使用者及主機指令優先；本檔不增加外部寫入授權。

## 職責與邊界

skills/ 保存上游技能；skills-sources.json 記錄來源，go-workflow.md 提供 Go 工作流程。角色註冊由 ../.codex/config.toml 與 ../.codex/agents/ 管理。

保留上游技能、授權與來源資訊；不要把專案政策複製進技能。角色數與可用性以配置及目前主機為準。

## 驗證

閱讀 ../docs/tooling/skills.md；涉及工具設定時執行設定檢查。此目錄被兩個分析工具排除，須直接檢查。

命令除另有標示外均在儲存庫根目錄執行。閱讀 [本目錄 README](README.md) 與 [文件索引](../docs/README.md)；回報實際執行結果與限制。
