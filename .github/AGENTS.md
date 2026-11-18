# .github 工作指引

適用於本目錄及其子目錄；先遵循 [根目錄 AGENTS.md](../AGENTS.md)，再套用較近的子目錄補充。使用者及主機指令優先；本檔不增加外部寫入授權。

## 職責與邊界

workflows/check.yml 管理設定、分析範圍、Shell、Go、前端與 API 檢查；ISSUE_TEMPLATE/、pull_request_template.md 與 SECURITY.md 管理協作入口。

保留既有 CI、最小 permissions 與 action 固定 revision。此處 CI 不等於 Cloud Run managed trigger，文件修改不授權部署。

## 驗證

核對 workflow 所引用腳本與任務存在；從根目錄執行設定檢查及 git diff --check。

命令除另有標示外均在儲存庫根目錄執行。閱讀 [本目錄 README](README.md) 與 [文件索引](../docs/README.md)；回報實際執行結果與限制。
