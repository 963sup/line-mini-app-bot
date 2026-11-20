# GitHub 協作與 CI

## 用途與範圍

保存 GitHub 協作模板與 CI 工作流程。

## 架構與責任邊界

CI 執行儲存庫檢查，Cloud Run 部署流程另由部署設定管理。

## 閱讀與修改入口

先查閱[workflows/check.yml](workflows/check.yml)、[pull_request_template.md](pull_request_template.md)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[本層規格與指引](../CONTRIBUTING.md)


## 目錄導覽與維護

workflows/check.yml 管理設定、分析範圍、Shell、Go、前端與 API 檢查；ISSUE_TEMPLATE/、pull_request_template.md 與 SECURITY.md 管理協作入口。

保留既有 CI、最小 permissions 與 action 固定 revision。此處 CI 不等於 Cloud Run managed trigger，文件修改不授權部署。

核對 workflow 所引用腳本與任務存在；從根目錄執行設定檢查及 git diff --check。

configuration job 同時執行本機 tooling unittest；只用標準庫與隔離 fake server，不呼叫外部 MCP 服務或要求雲端憑證。

[本目錄指引](AGENTS.md) · [根目錄說明](../README.md) · [根目錄指引](../AGENTS.md)

[開發與維運指引圖索引](../docs/diagrams.md#開發與維運指引) · [未來設計圖](../docs/diagrams.md#未來設計)。
