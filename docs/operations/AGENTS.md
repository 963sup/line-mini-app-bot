# 維運文件指引

沿用[文件指引](../AGENTS.md)，依[正式部署與維運索引](README.md)查找操作。

維運手冊說明目標環境、操作者、前置條件、成功證據與失敗後恢復方式。區分儲存庫宣告、線上設定及驗收結果。平台契約連回對應整合頁；環境名稱、工具版本和機密綁定以各自的權威來源為準。

更新手冊時核對命令來源、順序、連結與錨點。文件搬移不會套用線上設定。

調整根目錄 `cloudbuild.yaml`、`Dockerfile` 或 `firestore.indexes.json` 時，核對[部署決策](../decisions/0003-cloud-run-primary.md)、目標環境的建置入口與上傳範圍。部署目標為 Cloud Run，分開管理建置與執行帳戶，並透過 Secret Manager 綁定機密。
