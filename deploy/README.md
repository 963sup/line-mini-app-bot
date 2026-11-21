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

Repository 流程已接入部署前的 LINE 導航同步 Job；managed inline Trigger 另行維護，MUST 使用相容的 /server sync-line-navigation 入口與必要環境，不能沿用 /richmenu。參見 [自動發布與雲端前置條件](../docs/guides/line-setup/rich-menu.md)。

## 容器建置契約

Dockerfile 將 backend/main.go 編譯成映像內的 /server；HTTP 模式不帶參數，導航同步使用 /server sync-line-navigation。部署設定 MUST 與這個入口一致，修改 repository YAML 不會更新 managed inline Trigger。

本機執行 `bash scripts/tooling/tool-exec.sh mise run container-build`，建置後自動執行 [映像驗收](../scripts/tooling/container-smoke.sh)。環境有代理 CA 時，設定 BUILD_CA_FILE 為可讀 CA 檔案，透過 BuildKit secret 提供，不寫入映像。GitHub 的 container job 執行同一驗收，不提供雲端憑證、不發布 LINE 選單；此驗收確認執行檔、CLI dispatch、非 root 身分、CA、Poppler 與靜態資源，不取代真實 HTTP 與雲端整合測試。

Node、Go 基底版本及 pnpm MUST 與 mise.toml 一致，由 check-config.py 檢查；基底映像使用 digest 固定內容。升級時 MUST 一併更新版本與 digest，再建置驗收。Debian apt 套件仍隨套件來源更新，這不是逐位元可重現建置。版本依據見 [Docker 官方建置指引](https://docs.docker.com/build/building/best-practices/)。

依賴安裝與原始碼分層，BuildKit cache mount 重用 pnpm、Next.js 與 Go 快取。暫存 worker 不保證跨建置保留快取；尚未設定外部 cache 匯入／匯出，不宣稱 Cloud Build 已獲得跨次加速。AI_DOCUMENT_MODEL 可由部署環境覆寫，變更模型不需重建映像。目前 managed 服務未明確設定此值，因此保留映像既有預設，避免部署後停用文件 AI；等所有部署入口明確提供模型後再移除預設。

## 減少不必要部署

[Trigger 路徑政策](build-trigger-policy.json) 列出會影響映像或部署的檔案；純 docs、README、AGENTS 與 GitHub CI 文件變更不啟動 managed Trigger。混合程式與文件的提交仍會建置。新增建置輸入時 MUST 同步政策與雲端 includedFiles／ignoredFiles；只修改本檔不會套用雲端。

[過期建置檢查](../scripts/tooling/cloud-build-latest.py) 比對同一 Trigger 的最新建置建立時間，在建置前與第一次部署／發布之前執行。過期流程僅取消自己，API 查詢或取消失敗會中止流程，不放行部署。手動無 Trigger 的建置維持可執行。較新建置即使失敗，也不讓舊建置自動回補部署；回復版本應明確另開建置。

已在編譯的舊版本要等到下一個檢查點才取消；通過最後檢查點後不再取消，以免中止已開始的服務更新或 LINE 發布。檢查不是全域部署鎖；檢查後才抵達的新提交不會使已進入部署階段的流程停止。沒有新增排程器、worker、外部快取或 IAM 授權。

2026-11-21 已將同一腳本內容嵌入 managed inline Trigger 的 latest-before-build／latest-before-deploy 步驟，並套用路徑政策；未執行正式部署。嵌入形式避免尚未推送的 repository 腳本使線上流程找不到檔案。更新腳本時 MUST 同步雲端嵌入內容；repository YAML 使用腳本路徑，不會自動更新 managed Trigger。

LINE 同步既有 SHA-256 包含選單定義（含 MINI App 目標）與圖片，會核對遠端圖片、alias 與 default；一致時不重複寫入 LINE。仍保留讀取與 lease，才能偵測外部變更；不是看到 Git 沒改就無條件略過。此行為由 richmenu/publication_test.go 的重複同步測試驗證。

依據：[Cloud Build 查詢與過濾](https://docs.cloud.google.com/build/docs/view-build-results) 及固定版 gcloud 的 builds list／cancel／triggers update github leaf help。

## Cloud CLI 步驟的 Python 入口

Google Cloud CLI 592.0.0 slim 映像的 Python 位於 /usr/lib/google-cloud-sdk/platform/bundledpythonunix/bin/python3，PATH 不提供 python3。managed inline Trigger 與 repository YAML 的 latest-before-* 步驟 MUST 使用這個已驗證的入口；Dockerfile 只負責應用映像，無法修正建置步驟映像的入口。

修改步驟映像或 Python 路徑時，MUST 執行 `bash scripts/tooling/cloud-build-smoke.sh`，使用實際映像檢查腳本啟動與過期建置測試；container CI job 執行同一檢查。本測試不帶憑證、不連網、不部署，不能取代 build service account 的正式雲端權限驗證。
