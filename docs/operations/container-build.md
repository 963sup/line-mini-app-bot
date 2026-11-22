# 容器建置與驗收

## 容器建置契約

Dockerfile 將 backend/main.go 編譯成映像內的 /server；HTTP 模式不帶參數，導航同步使用 /server sync-line-navigation。部署設定 MUST 與這個入口一致，修改 repository YAML 不會更新 managed inline Trigger。

本機執行 `bash scripts/tooling/tool-exec.sh mise run container-build`，建置後自動執行 [映像驗收](../../scripts/test/container-smoke.sh)。環境有代理 CA 時，設定 BUILD_CA_FILE 為可讀 CA 檔案，透過 BuildKit secret 提供，不寫入映像。GitHub 的 container job 執行同一驗收，不提供雲端憑證、不發布 LINE 選單；此驗收確認執行檔、CLI dispatch、非 root 身分、CA、Poppler 與靜態資源，不取代真實 HTTP 與雲端整合測試。

Node、Go 基底版本及 pnpm MUST 與 mise.toml 一致，由 check-config.py 檢查；基底映像使用 digest 固定內容。升級時 MUST 一併更新版本與 digest，再建置驗收。Debian apt 套件仍隨套件來源更新，這不是逐位元可重現建置。版本依據見 [Docker 官方建置指引](https://docs.docker.com/build/building/best-practices/)。

依賴安裝與原始碼分層，BuildKit cache mount 重用 pnpm、Next.js 與 Go 快取。暫存 worker 不保證跨建置保留快取；尚未設定外部 cache 匯入／匯出，不宣稱 Cloud Build 已獲得跨次加速。AI_DOCUMENT_MODEL 可由部署環境覆寫，變更模型不需重建映像。目前 managed 服務未明確設定此值，因此保留映像既有預設，避免部署後停用文件 AI；等所有部署入口明確提供模型後再移除預設。

## BuildKit 與建置機密

Dockerfile 使用 `RUN --mount=type=secret,id=build-ca`，建置器必須支援並啟用 BuildKit。Dockerfile 的 syntax directive 只選擇語法，本機 Docker 的預設也無法證明 Cloud Build 的設定。更換 builder 時，以實際 builder 驗證 BuildKit／Buildx 和 secret mount。建置過程須保留 TLS 驗證，CA 與憑證不得寫入映像。

[Docker BuildKit](https://docs.docker.com/build/buildkit/) · [Docker build secrets](https://docs.docker.com/build/building/secrets/)

[Dockerfile](../../Dockerfile) · [Cloud Run 部署](cloud-run.md) · [正式維運索引](README.md)
