# LINE MINI App Bot

LINE MINI App Bot 是 Go 與 Next.js monorepo，整合 LINE Messaging API、MINI App、Firestore 與 Google Cloud。正式形態為單一 Go service，在 Cloud Run 同時提供 HTTP API、Webhook 與靜態輸出的前端。

## 專案結構

| 路徑                | 責任                                                    |
| ------------------- | ------------------------------------------------------- |
| `backend/`          | Go 業務模組、身分、平台 adapter、transport 與 bootstrap |
| `api/`              | OpenAPI 契約、產生設定與 TypeScript client              |
| `frontend/miniapp/` | Next.js Pages Router MINI App                           |
| `assets/`           | Rich Menu 等應用資產                                    |
| `scripts/db/`       | Firestore 設定工具                                      |
| `docs/`             | 現行工程文件                                            |
| `docs-old/`         | 尚未蒸餾或只供歷史查閱的舊文件                          |

後端入口是 `backend/main.go`，依賴在 `backend/internal/bootstrap/` 組裝。HTTP 契約以 `api/contract/` 為權威來源。

## 架構

- [文件索引](docs/README.md)
- [架構總覽](docs/architecture/overview.md)
- [架構邊界](docs/architecture/boundaries.md)
- [依賴方向](docs/architecture/dependencies.md)
- [整合架構](docs/architecture/integration.md)
- [LINE 後端責任、修改入口與驗收](docs/architecture/line.md)
- [部署架構](docs/architecture/deployment.md)
- [架構決策](docs/decisions/README.md)

專案採模組化單體與輕量 Hexagonal Architecture。業務規則留在 Domain／Application；HTTP、LINE、Firestore 與雲端 SDK 由 Adapter 隔離；Bootstrap 只負責組裝。

## 技術

| 範圍       | 技術                                                              |
| ---------- | ----------------------------------------------------------------- |
| Backend    | Go、`net/http`、Firestore、Cloud Storage、OpenTelemetry           |
| Frontend   | Next.js、React、TypeScript、Tailwind CSS                          |
| Contract   | OpenAPI、oapi-codegen、openapi-typescript                         |
| Platform   | LINE Messaging API、LINE MINI App／LIFF                           |
| Deployment | Docker、Cloud Build、Artifact Registry、Cloud Run、Secret Manager |

精確版本以 [Go module](backend/go.mod)、[frontend package](frontend/miniapp/package.json)、[API client package](api/client/package.json) 與 `pnpm-lock.yaml` 為準。

## 設定

服務不會自動載入 `.env`。環境變數名稱與範例見 [.env.example](.env.example)，解析與必填條件以 `backend/internal/bootstrap/` 為準。正式秘密必須透過受控環境或 Secret Manager 注入，不得提交 LINE token、channel secret、API key 或雲端憑證。

- [設定管理](docs/operations/configuration.md)
- [安全規範](docs/standards/security.md)
- [外部系統](docs/references/external_systems.md)

## 開發與驗證

從儲存庫根目錄安裝 workspace 依賴：

```bash
pnpm install --frozen-lockfile
```

依變更範圍執行：

```bash
go -C backend test ./...
go -C backend vet ./...
go -C backend build ./...
pnpm --filter @repo/miniapp check
pnpm --filter @repo/miniapp test
git diff --check
```

實際任務與建置入口以目前 `package.json`、Go module 及 [CI workflow](.github/workflows/check.yml) 為準。工具成功不代表 LINE、Firestore、Cloud Run 或真實裝置已完成整合驗收。

- [開發流程](docs/workflows/development.md)
- [重構流程](docs/workflows/refactoring.md)
- [驗證流程](docs/workflows/verification.md)
- [發布流程](docs/workflows/release.md)
- [貢獻指南](CONTRIBUTING.md)
- [安全問題回報](SECURITY.md)

## 工程規範

- [文件](docs/standards/documentation.md)
- [架構](docs/standards/architecture.md)
- [後端](docs/standards/backend.md)
- [前端](docs/standards/frontend.md)
- [API](docs/standards/api.md)
- [安全](docs/standards/security.md)
- [測試](docs/standards/testing.md)

根 [AGENTS.md](AGENTS.md) 定義共通工程規則；`api/`、`backend/`、`frontend/` 與 `assets/` 的 AGENTS.md 只補充各模組限制。

## 儲存庫分析

可按需使用 scoped Repomix bundle 或程式圖協助定位，但結果必須回到目前原始碼、契約與測試核對。聚合輸出放在儲存庫外或忽略路徑，不得把它當成正確性或整合證據。

本專案尚未選擇開源授權；權利依 [LICENSE](LICENSE) 保留。
