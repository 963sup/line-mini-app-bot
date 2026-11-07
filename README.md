# LINE MINI App Bot

以 Go 建置 LINE Messaging API Webhook、MINI App API 與 Firestore 整合。
目前提供最小 HTTP 服務與 `GET /healthz`；LINE 與 Firestore 業務整合尚未實作。

首版功能方向為「基礎出勤」，應用僅部署至 Fly.io，機密透過 Fly Secrets 注入。
目前先穩定基礎，不新增測試方案、案例或工具；保留既有 CI。
架構與階段範圍依 [已接受的基線](docs/architecture/decisions/0001-baseline.md)，
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
# 另一個終端：curl http://localhost:8080/healthz
```

服務讀取 `PORT`（預設 8080），不會自動載入 `.env`。其他設定名稱見 `.env.example`。
目前 HTTP 服務僅使用標準庫；`go.mod`／`go.sum` 已納入整合所需的第三方依賴，
不表示 LINE 與 Firestore 功能已完成串接。

授權尚未選定，目前依 `LICENSE` 保留所有權利。
