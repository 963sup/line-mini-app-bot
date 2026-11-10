# deploy 工作指引

適用於本目錄及其子目錄；先遵循 [根目錄 AGENTS.md](../AGENTS.md)，再套用較近的子目錄補充。使用者及主機指令優先；本檔不增加外部寫入授權。

## 職責與邊界

cloudbuild.yaml 為手動提交或自訂 trigger 的建置部署流程；firestore/indexes.json 保存索引設定。現有 GitHub managed trigger 使用 Dockerfile，不執行此 YAML。

部署目標為 Cloud Run；區分 build 與 runtime service account，機密使用 Secret Manager reference。修改檔案不代表已套用雲端設定或索引。

## 驗證

部署工作先讀 ../docs/operations/cloud-run.md 與部署 ADR；本次只核對本機檔案，不執行 deploy、IAM 或索引寫入。

命令除另有標示外均在儲存庫根目錄執行。閱讀 [本目錄 README](README.md) 與 [文件索引](../docs/README.md)；回報實際執行結果與限制。
