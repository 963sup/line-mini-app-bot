# .vscode 工作指引

適用於本目錄及其子目錄；先遵循 [根目錄 AGENTS.md](../AGENTS.md)，再套用較近的子目錄補充。使用者及主機指令優先；本檔不增加外部寫入授權。

## 職責與邊界

settings.json 指向 scripts/tooling 的 Go/gopls wrappers；tasks.json 提供 bootstrap、doctor、Go check 與設定檢查；extensions.json 列推薦擴充。

所有任務以儲存庫根目錄為工作目錄；沿用工具 wrapper，不加入個人絕對路徑、token 或自動版本更新。

## 驗證

解析三個 JSON，核對 task 的腳本與 cwd。Repomix 本次未收錄此目錄，須直接檢查。

命令除另有標示外均在儲存庫根目錄執行。閱讀 [本目錄 README](README.md) 與 [文件索引](../docs/README.md)；回報實際執行結果與限制。
