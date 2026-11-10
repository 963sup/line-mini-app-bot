# 部署設定與 Firestore 索引

## 目錄導覽與維護

cloudbuild.yaml 為手動提交或自訂 trigger 的建置部署流程；firestore/indexes.json 保存索引設定。現有 GitHub managed trigger 使用 Dockerfile，不執行此 YAML。

部署目標為 Cloud Run；區分 build 與 runtime service account，機密使用 Secret Manager reference。修改檔案不代表已套用雲端設定或索引。

部署工作先讀 ../docs/operations/cloud-run.md 與部署 ADR；本次只核對本機檔案，不執行 deploy、IAM 或索引寫入。

[本目錄指引](AGENTS.md) · [根目錄說明](../README.md) · [根目錄指引](../AGENTS.md)
