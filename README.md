# LINE MINI App Bot

以 Go 建置 LINE Messaging API Webhook、MINI App API 與 Firestore 整合。
目前提供 HTTP 服務與可選的 [LINE 接入能力](docs/guides/line-setup/README.md)：
一對一／群組指令、Flex／Postback、MINI App 身分驗證與 Rich Menu CLI。
另有 Firestore 唯讀連線檢查工具；基礎出勤已實作 Firestore 持久化，目標環境的端到端驗收仍需另行確認。服務沒有健康檢查路由。

出勤分工：群組發起上下班，從連結開啟 MINI App、取得位置後送出；1:1 私訊「出勤紀錄」選擇群組，再輸入「出勤紀錄 編號」查本人最近五筆。各群組狀態獨立，定位只記錄位置，不判斷工作半徑。

另有 [Go + Next.js／React 定位示範](docs/architecture/integrations/web-geolocation.md)：啟動後開啟
`http://localhost:8080/location`，點擊按鈕取得位置並傳送至 Go 驗證，不保存位置。

首版功能方向為「基礎出勤」，應用部署至 Google Cloud Run，機密透過
Secret Manager 綁定。
目前 GitHub `main` 的 managed trigger 使用 `Dockerfile` 建置映像並更新 Cloud Run；
它不執行 `deploy/cloudbuild.yaml`。該檔保留給手動提交或改成自訂 Trigger 時使用。
runtime 環境變數與 Secret reference 儲存在 Cloud Run service configuration；
兩種部署方式、首次設定與 IAM 見 [Cloud Run 部署](docs/guides/release/cloud-run.md)。
三個 MINI App 環境共用的 HTTPS Endpoint 已內建；桌面 Flex 入口與手機永久連結的分流見 [LINE 環境變數](docs/architecture/integrations/line-environment.md#3-mini-app)。
目前先穩定基礎，不新增測試方案、案例或工具；保留既有 CI。
架構與階段範圍依 [已接受的基線](docs/decisions/0002-baseline.md)，
首版只處理上班、下班與本人紀錄，見 [基礎出勤](docs/domains/attendance/README.md)。
開發順序為 [執行／部署基礎 → 資料與 LINE 接入 → 基礎出勤](docs/guides/local-development/environment-verification.md)，
不在基礎設施尚未確認可用時直接堆疊業務功能；具體出勤規則集中在功能規格確認。

原始碼分為 `backend/`（Go）、`frontend/miniapp/`（Next.js）與 `api/`（OpenAPI 與 TS client）。後端入口為 `backend/cmd/server/main.go`，功能整合依出勤、摘要與天氣歸位。
首次建置後可在開發時開啟 `http://localhost:3000/miniapp`；正式靜態輸出由 Go 提供。遷移範圍與尚未完成項目見 [Monorepo 決策](docs/decisions/0004-monorepo.md)。
API 契約、產生碼與相容性流程見 [API 工作流程](api/README.md)。
完整檔案樹見 [目前目錄](docs/architecture/runtime/code-layout.md)，功能入口見 [文件索引](docs/README.md)。

## 技術棧與 SDK

專案採用 Go、Next.js／React、TypeScript、OpenAPI 與 Google Cloud 的 Monorepo 架構。
Next.js 正式建置採靜態匯出，由 Go 在 Cloud Run 提供網頁與 HTTP API。

| 分類 | 技術與用途 | 設定與來源 |
| --- | --- | --- |
| 後端 | Go、標準函式庫 `net/http` | [後端說明](backend/README.md)、[go.mod](backend/go.mod) |
| 前端 | Next.js、React、TypeScript | [前端依賴](frontend/miniapp/package.json)、[靜態匯出設定](frontend/miniapp/next.config.ts) |
| UI | shadcn/ui、Radix UI、Tailwind CSS、Lucide 圖示 | [前端說明](frontend/README.md) |
| API 契約 | OpenAPI；以 oapi-codegen 與 openapi-typescript 產生 Go／TypeScript 程式碼 | [API 工作流程](api/README.md) |
| 資料儲存 | Cloud Firestore | [持久化文件](docs/architecture/README.md#資料與儲存) |
| 工作區與建置 | Node.js、pnpm workspaces、Turborepo；mise 管理工具鏈 | [工作區設定](package.json)、[工具設定](mise.toml) |
| 部署 | Docker 多階段建置、Cloud Build、Cloud Run、Secret Manager | [Dockerfile](Dockerfile)、[部署說明](docs/guides/release/cloud-run.md) |
| 品質檢查 | GitHub Actions、Go 測試與檢查工具、Node Test Runner／tsx／jsdom | [CI 說明](.github/README.md)、[工具設定](mise.toml)、[前端測試指令](frontend/miniapp/package.json) |

主要 SDK 與套件如下；完整依賴與固定版本以 [backend/go.mod](backend/go.mod)、
[前端 package.json](frontend/miniapp/package.json) 及 [API client package.json](api/client/package.json) 為準。

| SDK／套件 | 用途 |
| --- | --- |
| `github.com/line/line-bot-sdk-go/v8` | LINE Messaging API 與 Webhook 整合 |
| `cloud.google.com/go/firestore` | Firestore 存取 |
| `github.com/sashabaranov/go-openai` | OpenAI 整合，使用第三方 Go SDK |
| `google.golang.org/api`、`golang.org/x/oauth2` | Google API 與身分驗證支援 |
| `github.com/getkin/kin-openapi`、`github.com/oapi-codegen/nethttp-middleware`、`github.com/oapi-codegen/runtime` | OpenAPI 契約處理、請求驗證與 Go 產生碼支援 |
| `openapi-fetch` | TypeScript API client |
| `pdfjs-dist` | 前端 PDF 解析 |

技術基準與契約歸屬見[依賴與契約維護位置](docs/architecture/runtime/dependencies.md)。
本節描述技術組成，不代表外部服務或正式環境已完成端到端驗收。首版功能範圍仍以[已接受的基線](docs/decisions/0002-baseline.md)為準。

## 開始開發

```bash
bash scripts/tooling/bootstrap.sh core
bash scripts/tooling/tool-exec.sh pnpm build
bash scripts/tooling/tool-exec.sh pnpm check
bash scripts/tooling/tool-exec.sh pnpm test
bash scripts/tooling/tool-exec.sh pnpm dev
```

完整 Cloud／MCP 安裝與限制請見 [Codex Cloud 設定](docs/guides/local-development/setup.md)。
預設初始化會安裝 Go、Web 依賴及 Mermaid CLI。需要 Graphify／Repomix 與 MCP 時使用
`bash scripts/tooling/bootstrap.sh all`。Firestore Toolbox 需要先提供 Google ADC，GKE MCP 預設停用。
工具安裝、
MCP 握手與聊天主機實際可用性須分別驗收，不因快取存在而視為全新環境已驗證。
工具與非機密環境以 `mise.toml` 為準；日常命令使用 `scripts/tooling/tool-exec.sh`。


- [文件索引與 Backend／API／Frontend／Scripts／Deploy 閱讀入口](docs/README.md#按模組閱讀)
- [需求到實作導航](docs/architecture/runtime/code-layout.md#從需求定位修改範圍)
- [變更驗收矩陣](docs/development/acceptance.md#變更驗收矩陣)
- [架構與流程](docs/architecture/runtime/overview.md)
- [Go 開發](docs/standards/go.md)
- [貢獻方式](CONTRIBUTING.md)
- [安全問題回報](.github/SECURITY.md)

工具版本與非機密環境變數由 `mise.toml` 管理。請勿提交 LINE token、channel secret 或雲端憑證。

## 儲存庫分析

在儲存庫根目錄執行既有 CLI 任務：

```bash
bash scripts/tooling/tool-exec.sh mise run graphify-update
bash scripts/tooling/tool-exec.sh graphify query "registerLINEWebhook"
bash scripts/tooling/tool-exec.sh mise run repomix
```

Graphify 輸出 `graphify-out/graph.json`，目前任務只做 AST 抽取；文件、TOML、
HTML/CSS 與鎖檔不保證被解析。使用 Repomix 的 `repomix-output.xml` 閱讀文件及設定原文，
再對照實際來源；打包與查詢成功都不代表程式正確或整合已驗收。
兩工具均排除 `.agents`、`.claude`、`agent` 與 `skills-lock.json`；`.codex` 保留為分析來源，
仍受工具原生格式支援與內建排除限制。各 ignore 規則不等於完整收錄保證。
範圍契約及獨立 audit 的用途與限制見 [分析證據](docs/development/tooling/analysis-scope.md)。

## 執行服務

```bash
bash scripts/tooling/tool-exec.sh mise run dev
# LINE 憑證未注入時不註冊 Webhook；連線方式見接入試作文件。
```

服務讀取 `PORT`（預設 8080）、Messaging API 三項與 MINI App 環境選擇設定，不會自動載入 `.env`。
設定名稱見 `.env.example`；MINI App 預設使用 developing，亦可選 review 或 published。
MINI App 頁面為 `/miniapp`；公開 Channel／LIFF ID 與 Endpoint 由所選環境三元組提供，缺值會啟動失敗；實際登入另需 Session 資料庫設定。
操作與驗證限制見 [LINE 接入說明](docs/guides/line-setup/README.md)。

授權尚未選定，目前依 `LICENSE` 保留所有權利。

## 巢狀目錄導覽

各目錄 README 說明用途，AGENTS.md 規範該子樹的修改與驗證；根目錄指引持續適用。

| 目錄 | 說明 | 工作指引 |
| --- | --- | --- |
| `.agents/` | [共用 Agent Skills](.agents/README.md) | [AGENTS.md](.agents/AGENTS.md) |
| `.codex/` | [Codex 專案配置](.codex/README.md) | [AGENTS.md](.codex/AGENTS.md) |
| `.github/` | [GitHub 協作與 CI](.github/README.md) | [AGENTS.md](.github/AGENTS.md) |
| `.vscode/` | [VS Code 工作區設定](.vscode/README.md) | [AGENTS.md](.vscode/AGENTS.md) |
| `api/` | [HTTP API 契約](api/README.md) | [AGENTS.md](api/AGENTS.md) |
| `assets/` | [應用素材](assets/README.md) | [AGENTS.md](assets/AGENTS.md) |
| `backend/` | [Go 後端](backend/README.md) | [AGENTS.md](backend/AGENTS.md) |
| `deploy/` | [部署設定與 Firestore 索引](deploy/README.md) | [AGENTS.md](deploy/AGENTS.md) |
| `docs/` | [專案文件](docs/README.md) | [AGENTS.md](docs/AGENTS.md) |
| `scripts/` | [腳本工具](scripts/README.md) | [AGENTS.md](scripts/AGENTS.md) |
| `frontend/` | [MINI App 前端](frontend/README.md) | [AGENTS.md](frontend/AGENTS.md) |

進入 `frontend/miniapp/` 另讀其 [Next.js 指引](frontend/miniapp/AGENTS.md)。分析結果與涵蓋限制見 [分析範圍與證據限制](docs/development/tooling/analysis-scope.md)。

## 業務操作開發原則

採用 [資源導向的通用業務操作模型](docs/architecture/contracts/business-operations.md)：Actor、Resource、Action、Context、Authority、Policy、Decision 與 Result。
開發前定義操作契約，交付時對照實作與證據；讀取不強制建立事件，採納提案時才使用 Merge。
發現不符合時依 [Graphify／Repomix／Context7 符合性流程](docs/development/workflow.md#business-operation-compliance) 在任務範圍內修正。
統一操作模型，隔離業務規則；共用必要機制，避免不必要的抽象。

[現行實作圖索引](docs/diagrams.md#現行實作) · [未來設計圖](docs/diagrams.md#未來設計)。
