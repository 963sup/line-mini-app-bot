# 正式部署與維運

此目錄面向正式服務的部署、驗收與復原。應用程式內的操作與 PDF 工具見[業務操作能力](../domains/operations/README.md)。

| 要完成的工作 | 入口 |
| --- | --- |
| 建置及驗收本機容器 | [容器建置與驗收](container-build.md) |
| 準備或更新建置、Trigger、IAM 與 runtime 設定 | [Cloud Run 部署](cloud-run.md) |
| 設定 runtime 變數、Secret Manager 與公開設定 | [正式環境設定](configuration.md) |
| 收集發布證據、進行真機驗收、切換流量或回復 | [發布驗收與復原](release-acceptance.md) |
| 同步 LINE Rich Menu、驗證 Alias 與處理重試 | [Rich Menu 發布](rich-menu.md) |
| 排查 Cloud Build、Cloud Run 或部署後同步失敗 | [正式環境故障排查](troubleshooting.md) |

根目錄的 [cloudbuild.yaml](../../cloudbuild.yaml)供手動或自訂 Trigger 使用，[Dockerfile](../../Dockerfile)定義容器，[firestore.indexes.json](../../firestore.indexes.json)定義索引。

正式設定與結果以目標環境的觀察為準。LINE 平台接合見 [LINE 整合](../integrations/line/README.md)，本機工具問題見[開發環境排查](../development/environment/troubleshooting.md)。

[文件入口](../README.md) · [維護指引](AGENTS.md)
