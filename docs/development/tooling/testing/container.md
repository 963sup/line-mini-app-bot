# 容器驗收工具

映像入口、版本與 BuildKit 契約由[容器建置規格](../../../operations/container-build.md)維護；正式部署依[發布驗收](../../../operations/release-acceptance.md)。

## 本機建置與 smoke 驗收

本機執行 `bash scripts/tooling/tool-exec.sh mise run container-build`，建置後自動執行 [映像驗收](../../../../scripts/test/container-smoke.sh)。環境有代理 CA 時，設定 BUILD_CA_FILE 為可讀 CA 檔案，透過 BuildKit secret 提供，不寫入映像。GitHub 的 container job 執行同一驗收，不提供雲端憑證、不發布 LINE 選單；此驗收確認執行檔、CLI dispatch、非 root 身分、CA、Poppler 與靜態資源，不取代真實 HTTP 與雲端整合測試。
