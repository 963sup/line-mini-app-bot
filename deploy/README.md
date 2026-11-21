# 部署設定與 Firestore 索引

## 用途與範圍

保存 Cloud Build 設定與 Firestore 索引。

## 架構與責任邊界

此處定義部署設定與索引，容器組裝由根目錄 Dockerfile 管理。

## 閱讀與修改入口

先查閱[cloudbuild.yaml](cloudbuild.yaml)、[firestore/indexes.json](firestore/indexes.json)、[本目錄工作指引](AGENTS.md)。

## 相關規格

[文件索引：Deploy](../docs/README.md#按模組閱讀) · [Cloud Run 部署流程](../docs/guides/release/cloud-run.md)


## 目錄導覽與維護

cloudbuild.yaml 為手動提交或自訂 trigger 的建置部署流程；firestore/indexes.json 保存索引設定。現有 GitHub managed trigger 使用 Dockerfile，不執行此 YAML。

部署目標為 Cloud Run；區分 build 與 runtime service account，機密使用 Secret Manager reference。修改檔案不代表已套用雲端設定或索引。

部署工作先讀 ../docs/guides/release/cloud-run.md 與部署 ADR；本次只核對本機檔案，不執行 deploy、IAM 或索引寫入。

[本目錄指引](AGENTS.md) · [根目錄說明](../README.md) · [根目錄指引](../AGENTS.md)

## 相關文件

- [文件索引：Deploy](../docs/README.md#按模組閱讀)
- [Cloud Run 部署流程](../docs/guides/release/cloud-run.md)
- [發布驗收](../docs/guides/release/release-acceptance.md)
- [故障排查](../docs/guides/release/troubleshooting.md)
- [Cloud Run 決策](../docs/decisions/0003-cloud-run-primary.md)

[開發與維運指引圖索引](../docs/diagrams.md#開發與維運指引) · [未來設計圖](../docs/diagrams.md#未來設計)。

Repository 流程已接入部署前的 LINE 導航同步 Job；managed inline Trigger 必須先遷移才可部署移除 /richmenu 的映像。參見 [自動發布與雲端前置條件](../docs/guides/line-setup/rich-menu.md)。
