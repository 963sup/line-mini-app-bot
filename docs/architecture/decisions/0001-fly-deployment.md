# ADR 0001：Fly.io 是唯一應用部署目標

- 狀態：已接受（使用者明確指定）；實際資源與 Worker 尚待配置驗收。
- 日期：2026-11-07。

範圍更新：Fly-only 決策維持；Worker 首版必要性的假設由 [ADR 0002](0002-baseline.md)
取代。以下 Worker 說明僅在後續需求確立時適用，目前不部署 Worker。

## 決策

Go HTTP server 與持續 Worker 部署 Fly.io。Firestore 仍為 Google Cloud 資料庫；不使用 Cloud Run 執行應用。
共用 module／映像可以使用不同 process command，但現有 Dockerfile 僅產生 server，fly.toml 僅配置 HTTP；新增 Worker 前必須更新建置與 process/Machine 配置。
HTTP auto-start 不負責喚醒輪詢 Worker；Worker 需要常駐 Machine、持續運算與 graceful shutdown。
區域依 Firestore 位置與量測確認，不把現有 nrt 視為已證明最佳延遲。單 Machine 不代表 HA。

## 認證與影響

LINE secrets 使用 Fly Secrets。Fly 執行環境的 Firestore ADC/IAM 需獨立配置，不能假設平台提供 Google service identity。
優先評估可驗證的短效 federation；若採 service account key，需明確安全注入、最小 IAM、輪替與撤銷，不提交 repo。
目前使用 Fly Secrets，不引入 Secret Manager 依賴。
本 ADR 不授權建立外部資源或部署。

## 驗收

映像包含 server/worker、SDK 相依與 internal；確認 HTTP routing 不導至 Worker、Worker 不自動停止、失租約恢復、SIGTERM、ADC/IAM、區域與成本。
正式資源規模、憑證方式與高可用仍需決策。[Fly 操作](../../operations/fly.md)。
