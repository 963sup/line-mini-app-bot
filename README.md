# LINE MINI App Bot

以 Go 建置 LINE Messaging API Webhook、MINI App API 與 Firestore 整合。
目前提供 HTTP 服務與可選的 [LINE 接入能力](docs/development/line-prototype.md)：
一對一／群組指令、Flex／Postback、MINI App 身分驗證與 Rich Menu CLI。
另有 Firestore 唯讀連線檢查工具；基礎出勤已實作 Firestore 持久化，目標環境的端到端驗收仍需另行確認。服務沒有健康檢查路由。

出勤分工：群組發起上下班，從連結開啟 MINI App、取得位置後送出；1:1 私訊「出勤紀錄」選擇群組，再輸入「出勤紀錄 編號」查本人最近五筆。各群組狀態獨立，定位只記錄位置，不判斷工作半徑。

另有 [Go + Next.js／React 定位示範](docs/integrations/web-geolocation.md)：啟動後開啟
`http://localhost:8080/location`，點擊按鈕取得位置並傳送至 Go 驗證，不保存位置。

首版功能方向為「基礎出勤」，應用部署至 Google Cloud Run，機密透過
Secret Manager 綁定。
目前 GitHub `main` 的 managed trigger 使用 `Dockerfile` 建置映像並更新 Cloud Run；
它不執行 `deploy/cloudbuild.yaml`。該檔保留給手動提交或改成自訂 Trigger 時使用。
runtime 環境變數與 Secret reference 儲存在 Cloud Run service configuration；
兩種部署方式、首次設定與 IAM 見 [Cloud Run 部署](docs/operations/cloud-run.md)。
目前先穩定基礎，不新增測試方案、案例或工具；保留既有 CI。
架構與階段範圍依 [已接受的基線](docs/architecture/decisions/0002-baseline.md)，
首版只處理上班、下班與本人紀錄，見 [基礎出勤](docs/features/attendance.md)。
開發順序為 [執行／部署基礎 → 資料與 LINE 接入 → 基礎出勤](docs/development/foundation.md)，
不在基礎設施尚未確認可用時直接堆疊業務功能；具體出勤規則集中在功能規格確認。

原始碼分為 `backend/`（Go）、`frontend/miniapp/`（Next.js）與 `api/`（OpenAPI 與 TS client）。後端入口為 `backend/cmd/server/main.go`，功能整合依出勤、摘要與天氣歸位。
首次建置後可在開發時開啟 `http://localhost:3000/miniapp`；正式靜態輸出由 Go 提供。遷移範圍與尚未完成項目見 [Monorepo 決策](docs/architecture/decisions/0004-monorepo.md)。
API 契約、產生碼與相容性流程見 [API 工作流程](api/README.md)。
完整檔案樹見 [目前目錄](docs/architecture/layout.md)，功能入口見 [文件索引](docs/README.md)。

## 開始開發

```bash
bash scripts/bootstrap.sh core
bash scripts/tool-exec.sh mise run doctor
bash scripts/tool-exec.sh mise run install-web
bash scripts/tool-exec.sh pnpm install --frozen-lockfile
bash scripts/tool-exec.sh pnpm build
bash scripts/tool-exec.sh pnpm check
bash scripts/tool-exec.sh pnpm test
bash scripts/tool-exec.sh pnpm dev
```

完整 Cloud／MCP 安裝與限制請見 [Codex Cloud 設定](docs/development/codex-cloud.md)。
需要 Graphify／Repomix 與 MCP 時使用 `bash scripts/bootstrap.sh all`；工具安裝、
MCP 握手與聊天主機實際可用性須分別驗收，不因快取存在而視為全新環境已驗證。
工具與非機密環境以 `mise.toml` 為準；日常命令使用 `scripts/tool-exec.sh`。


- [文件索引](docs/README.md)
- [架構與流程](docs/architecture/overview.md)
- [Go 開發](docs/development/go.md)
- [貢獻方式](CONTRIBUTING.md)
- [安全問題回報](.github/SECURITY.md)

工具版本與非機密環境變數由 `mise.toml` 管理。請勿提交 LINE token、channel secret 或雲端憑證。

## 儲存庫分析

在儲存庫根目錄執行既有 CLI 任務：

```bash
bash scripts/tool-exec.sh mise run graphify-update
bash scripts/tool-exec.sh graphify query "registerLINEWebhook"
bash scripts/tool-exec.sh mise run repomix
```

Graphify 輸出 `graphify-out/graph.json`，目前任務只做 AST 抽取；文件、TOML、
HTML/CSS 與鎖檔不保證被解析。使用 Repomix 的 `repomix-output.xml` 閱讀文件及設定原文，
再對照實際來源；打包與查詢成功都不代表程式正確或整合已驗收。
兩工具均排除 `.agents`、`.claude`、`agent` 與 `skills-lock.json`；`.codex` 保留為分析來源，
仍受工具原生格式支援與內建排除限制。各 ignore 規則不等於完整收錄保證。
範圍契約及獨立 audit 的用途與限制見 [分析證據](docs/tooling/analysis-scope.md)。

## 執行服務

```bash
bash scripts/tool-exec.sh mise run dev
# LINE 憑證未注入時不註冊 Webhook；連線方式見接入試作文件。
```

服務讀取 `PORT`（預設 8080）、Messaging API 三項與 MINI App 兩項設定，不會自動載入 `.env`。
設定名稱見 `backend/.env.example`；Messaging API 與 MINI App 可各自啟用。
MINI App 頁面為 `/miniapp`；Channel 未設定時頁面顯示設定提示。
操作與驗證限制見 [LINE 接入說明](docs/development/line-prototype.md)。

授權尚未選定，目前依 `LICENSE` 保留所有權利。
