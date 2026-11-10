# VS Code 工作區設定

## 目錄導覽與維護

settings.json 指向 scripts/tooling 的 Go/gopls wrappers；tasks.json 提供 bootstrap、doctor、Go check 與設定檢查；extensions.json 列推薦擴充。

所有任務以儲存庫根目錄為工作目錄；沿用工具 wrapper，不加入個人絕對路徑、token 或自動版本更新。

解析三個 JSON，核對 task 的腳本與 cwd。Repomix 本次未收錄此目錄，須直接檢查。

[本目錄指引](AGENTS.md) · [根目錄說明](../README.md) · [根目錄指引](../AGENTS.md)
