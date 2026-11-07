# LINE MINI App Bot

以 Go 建置 LINE Messaging API Webhook、MINI App API 與 Firestore 整合。
目前提供 HTTP 服務與可選的 [LINE 接入試作](docs/development/line-prototype.md)，
沒有健康檢查路由；Firestore 與出勤功能尚未實作。

首版功能方向為「基礎出勤」，應用部署至 Google Cloud Run，機密透過
Secret Manager 綁定。
目前先穩定基礎，不新增測試方案、案例或工具；保留既有 CI。
架構與階段範圍依 [已接受的基線](docs/architecture/decisions/0002-baseline.md)，
首版只處理上班、下班與本人紀錄，見 [基礎出勤](docs/features/attendance.md)。
開發順序為 [執行／部署基礎 → 資料與 LINE 接入 → 基礎出勤](docs/development/foundation.md)，
不在基礎設施尚未確認可用時直接堆疊業務功能；具體出勤規則集中在功能規格確認。

## 開始開發

```bash
bash scripts/bootstrap.sh core
bash scripts/tool-exec.sh mise run doctor
```

完整 Cloud／MCP 安裝與限制請見 [Codex Cloud 設定](docs/development/codex-cloud.md)。

- [文件索引](docs/README.md)
- [架構與流程](docs/architecture/overview.md)
- [Go 開發](docs/development/go.md)
- [貢獻方式](CONTRIBUTING.md)
- [安全問題回報](.github/SECURITY.md)

工具版本與非機密環境變數由 `mise.toml` 管理。請勿提交 LINE token、channel secret 或雲端憑證。

## 執行服務

```bash
bash scripts/tool-exec.sh mise run dev
# LINE 憑證未注入時不註冊 Webhook；連線方式見接入試作文件。
```

服務讀取 `PORT`（預設 8080）與 Messaging API 三項設定，不會自動載入 `.env`。
設定名稱見 `.env.example`。LINE SDK 已用於接入試作；Firestore 依賴仍待實際串接。

授權尚未選定，目前依 `LICENSE` 保留所有權利。
