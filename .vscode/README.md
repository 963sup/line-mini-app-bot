# VS Code 工作區設定

## 用途與範圍

保存共用 VS Code 設定與開發任務。

## 架構與責任邊界

編輯器任務呼叫儲存庫工具入口，工具版本由專案設定管理。

## 閱讀與修改入口

先查閱[settings.json](settings.json)、[tasks.json](tasks.json)、[extensions.json](extensions.json)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[本層規格與指引](../docs/development/tooling/mise.md)


## 目錄導覽與維護

settings.json 指向 scripts/tooling 的 Go/gopls wrappers；tasks.json 提供 bootstrap、doctor、Go check 與設定檢查；extensions.json 列推薦擴充。

所有任務以儲存庫根目錄為工作目錄；沿用工具 wrapper，不加入個人絕對路徑、token 或自動版本更新。

解析三個 JSON，核對 task 的腳本與 cwd。Repomix 本次未收錄此目錄，須直接檢查。

[本目錄指引](AGENTS.md) · [根目錄說明](../README.md) · [根目錄指引](../AGENTS.md)

[開發與維運指引圖索引](../docs/diagrams.md#開發與維運指引) · [未來設計圖](../docs/diagrams.md#未來設計)。
