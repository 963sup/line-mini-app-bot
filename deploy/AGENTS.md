# deploy 工作指引

## 用途與範圍

保存 Cloud Build 設定與 Firestore 索引。

## 架構與責任邊界

此處定義部署設定與索引，容器組裝由根目錄 Dockerfile 管理。

## 閱讀與修改入口

先查閱[cloudbuild.yaml](cloudbuild.yaml)、[firestore/indexes.json](firestore/indexes.json)、[本目錄導覽](README.md)、[上層工作指引](../AGENTS.md)。

## 相關規格

[文件索引](../docs/README.md) · [模組閱讀入口](../docs/README.md#按模組閱讀)


適用於本目錄及其子目錄；先遵循 [根目錄 AGENTS.md](../AGENTS.md)，再套用較近的子目錄補充。使用者及主機指令優先；本檔不增加外部寫入授權。

## 職責與邊界

cloudbuild.yaml 為手動提交或自訂 trigger 的建置部署流程；firestore/indexes.json 保存索引設定。現有 GitHub managed trigger 使用 Dockerfile，不執行此 YAML。

部署目標為 Cloud Run；區分 build 與 runtime service account，機密使用 Secret Manager reference。修改檔案不代表已套用雲端設定或索引。

## 驗證

部署工作先讀 ../docs/guides/release/cloud-run.md 與部署 ADR；本次只核對本機檔案，不執行 deploy、IAM 或索引寫入。

命令除另有標示外均在儲存庫根目錄執行。閱讀 [本目錄 README](README.md) 與 [文件索引](../docs/README.md)；回報實際執行結果與限制。

## 文件閱讀路由

先由 [本模組相關文件](README.md#相關文件) 與 [模組閱讀入口](../docs/README.md#按模組閱讀) 選擇文件。局部用途與命令維護在本模組 README；跨模組架構、功能與營運文件依 docs 主題分類保存。索引更新依[文件維護指引](../docs/AGENTS.md#主題子目錄)。

[開發與維運指引圖索引](../docs/diagrams.md#開發與維運指引) · [未來設計圖](../docs/diagrams.md#未來設計)。

## Cloud Run 部署

應用程式僅部署至 Cloud Run。部署前讀取[部署決策](../docs/decisions/0003-cloud-run-primary.md)與[操作指南](../docs/guides/release/cloud-run.md)，使用專屬 runtime service account 與最小 IAM 權限。雲端工具查證與登入依[MCP 指引](../docs/development/tooling/mcp/AGENTS.md)。
