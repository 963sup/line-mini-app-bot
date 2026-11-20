# .vscode 工作指引

## 用途與範圍

保存共用 VS Code 設定與開發任務。

## 架構與責任邊界

編輯器任務呼叫儲存庫工具入口，工具版本由專案設定管理。

## 閱讀與修改入口

先查閱[settings.json](settings.json)、[tasks.json](tasks.json)、[extensions.json](extensions.json)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[文件索引](../docs/README.md)


適用於本目錄及其子目錄；先遵循 [根目錄 AGENTS.md](../AGENTS.md)，再套用較近的子目錄補充。使用者及主機指令優先；本檔不增加外部寫入授權。

## 職責與邊界

settings.json 指向 scripts/tooling 的 Go/gopls wrappers；tasks.json 提供 bootstrap、doctor、Go check 與設定檢查；extensions.json 列推薦擴充。

所有任務以儲存庫根目錄為工作目錄；沿用工具 wrapper，不加入個人絕對路徑、token 或自動版本更新。

## 驗證

解析三個 JSON，核對 task 的腳本與 cwd。Repomix 本次未收錄此目錄，須直接檢查。

命令除另有標示外均在儲存庫根目錄執行。閱讀 [本目錄 README](README.md) 與 [文件索引](../docs/README.md)；回報實際執行結果與限制。

[開發與維運指引圖索引](../docs/diagrams.md#開發與維運指引) · [未來設計圖](../docs/diagrams.md#未來設計)。
