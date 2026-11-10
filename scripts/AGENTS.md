# scripts 工作指引

適用於本目錄及其子目錄；遵循 [根目錄指引](../AGENTS.md)，使用者與主機指令優先。

## 職責與邊界

依 [腳本分類](README.md) 將工具、分析、API 與發布腳本分別放入 tooling、analysis、api、release。
工具版本、非機密環境與任務維持在 mise.toml；腳本實作任務，不複製機密或授予部署權限。

## 文件與驗證

從 [相關文件](README.md#相關文件) 與 [模組閱讀入口](../docs/README.md#按模組閱讀) 閱讀；跨模組規格依 docs 主題分類，新增或搬移文件同步更新索引。
命令從儲存庫根目錄執行，使用 scripts/tooling/tool-exec.sh。腳本變更執行設定檢查及相關任務 smoke check；純文件變更檢查連結與 git diff --check。不宣稱未執行的部署或驗收已成功。
