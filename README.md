# LINE MINI App Bot

以 Go 建置 LINE Messaging API Webhook、MINI App API 與 Firestore 整合。
目前提供最小 HTTP 服務與 `GET /healthz`；LINE 與 Firestore 業務整合尚未實作。

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
目前僅使用標準庫，因此沒有 `go.sum`；加入第三方依賴後由 Go 產生。

授權尚未選定，目前依 `LICENSE` 保留所有權利。
