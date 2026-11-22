# 開發手冊

先讀[根目錄指引](../../AGENTS.md)，再按目前的任務選入口。修改前讀取目標路徑上的巢狀 `AGENTS.md`，交付前按變更類型選檢查。

## 按任務開始

| 目前要做什麼 | 先讀 | 下一步 |
| --- | --- | --- |
| 準備本機或新的 Codex Cloud 環境 | [開發環境](environment/README.md) | 啟動所需服務或執行[新環境驗收](environment/codex-cloud.md#新環境驗收) |
| 開始、接手或界定工作範圍 | [開始任務](workflow/task-start.md) | 找到工作樹、適用指引與完成條件 |
| 修改 HTTP API | [API 操作配方](tooling/api/change-api.md) | 從契約追到生成碼與呼叫端 |
| 修改領域行為 | [領域操作配方](recipes/change-domain.md) | 從規格追到授權、寫入與既有測試 |
| 修改 MINI App 頁面 | [前端操作配方](recipes/change-frontend.md) | 從頁面入口追到 API 與瀏覽器驗收 |
| 分析錯誤或部署紅燈 | [系統化除錯](workflow/debugging.md)或[部署診斷](recipes/diagnose-deployment.md) | 重現錯誤或定位雲端失敗階段 |
| 選檢查、整理證據 | [驗收矩陣](tooling/testing/commands.md#變更驗收矩陣) | 按[證據限制](verification/evidence.md)記錄結果 |
| 交接工作或推送已驗證的階段 | [交接規格](workflow/handoff.md#sidebar-handoff)或[發布流程](workflow/publication.md#repository-publication) | 記錄可重現狀態或核對已驗證的提交 |
| 查 CLI、MCP、技能或來源分析 | [工具索引](tooling/README.md) | 選對應工具專頁 |

[開發流程索引](workflow/README.md) · [驗收索引](verification/README.md) · [操作配方索引](recipes/README.md)

舊書籤可透過[原流程](workflow.md)、[原驗收](acceptance.md)與[原除錯](debugging.md)的章節連到新位置。

## 歸屬與維護

開發手冊說明怎麼工作。[領域規格](../domains/README.md)定義應有行為，[架構文件](../architecture/README.md)定義邊界與契約，[腳本索引](../../scripts/README.md)指向可執行入口。工具版本與任務以 [mise.toml](../../mise.toml) 為準。

[本目錄指引](AGENTS.md) · [父層指引](../AGENTS.md) · [文件總索引](../README.md)
