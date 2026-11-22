# 專案技能與工作流程

技能來源與雜湊由[根目錄鎖檔](../skills-lock.json)維護。從[技能與工具索引](../docs/development/tooling/agents/skills.md)選擇與任務相關的上游技能；Go 工作參考[專案流程](go-workflow.md)。Samber 技能由社群維護。

- `mise run install-graphify` 使用固定版本，以 `--platform agents --project` 安裝[Graphify 技能](skills/graphify/SKILL.md)。版本與任務見 [mise.toml](../mise.toml)，需要更新時由安裝流程產生。程式圖與文件圖的來源、位置及涵蓋範圍見 [Graphify 指南](../docs/development/tooling/agents/graphify.md)。
- 角色與 MCP 由[專案設定](../.codex/config.toml)註冊，角色用途見[角色說明](../.codex/agents/README.md)。工具是否已載入須在目標主機驗證。

## Deployment and browser location routing

部署依[Cloud Run 決策](../docs/decisions/0003-cloud-run-primary.md)與[發布指南](../docs/operations/cloud-run.md)選擇專用執行身分與最小 IAM。定位依[瀏覽器取得與後端驗證契約](../docs/integrations/geolocation/browser-api.md)，涉及到場判定或歷史需有產品決策。設計交接依[任務範本](../docs/development/workflow/handoff.md#六欄交接範本)與[命名規範](../docs/standards/naming.md)。

## Mandatory documentation verification

來源、版本及查證限制依[根目錄指引](../AGENTS.md#mandatory-documentation-verification)，LINE、Google Cloud 與 OpenAI 的查詢入口依[工具指引](../docs/development/tooling/agents/AGENTS.md#官方來源)。

## 目錄導覽與維護

保留上游技能與授權，專案特有規則寫入[本層指引](AGENTS.md)。修改來源或工具設定後，核對鎖檔、實際安裝路徑與適用的[設定檢查](../scripts/check/README.md)。
