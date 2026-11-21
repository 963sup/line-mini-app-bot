# VS Code 工作區設定

從儲存庫根目錄開啟 VS Code。執行「Tasks: Run Task」可選擇以下[工作區任務](tasks.json)。

| 任務 | 用途 |
| --- | --- |
| Go: setup | 以 core 模式初始化本機開發工具 |
| Environment: doctor | 檢查開發環境 |
| Go: check application | 驗證 Go 模組、執行 vet 與 race 測試 |
| Repository: check configuration | 檢查儲存庫設定 |

Codex Cloud 的 Install script 使用 `scripts/bootstrap.sh`，預設為 cloud 模式。[Cloud 初始化說明](../.codex/README.md#codex-環境初始化)

[編輯器設定](settings.json) · [推薦擴充套件](extensions.json) · [本層維護指引](AGENTS.md)
