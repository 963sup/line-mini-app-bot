# 正式環境設定

此頁維護 Cloud Run runtime 設定、Secret Manager 綁定與公開設定遷移。[部署步驟](cloud-run.md)與[發布驗收](release-acceptance.md)另列。

<a id="設定契約"></a>

<a id="secret-manager-與環境變數"></a>

```mermaid
flowchart LR
  h["Secret Manager 與環境變數"]
  h --- n0["Secret Manager 與兩條部署路徑的關係如下。<br/>deploy/cloudbuild.yaml 的「部署檔參照」只在<br/>手動或 自訂 Trigger 真正執行該檔時生效；「歷<br/>史快照」記錄 2026-11-08 revision <br/>line-mini-app-bot-00027-695 的 metadata，不<br/>包含 Secret payload。此表不是現行服務宣告； <br/>部署前使用發布 snapshot重新讀取 serving <br/>revision、 流量、執行帳戶及 Secret <br/>references，另核對必要非機密 env 值："]
  click n0 "release-acceptance.md#發布證據與前置條件"
```

```mermaid
flowchart LR
  h["Secret Manager 與環境變數"]
  h --- n0["Secret Manager 名稱：cwa-api-key<br/>容器環境變數：CWA_API_KEY<br/>部署檔參照：1<br/>2026-11-08 歷史快照／用途：綁定 1；天氣功能<br/>使用。"]
  h --- n1["Secret Manager 名稱：gemini-api-key<br/>容器環境變數：GEMINI_API_KEY<br/>部署檔參照：latest<br/>2026-11-08 歷史快照／用途：未綁定；Secret 尚<br/>無 version，不能解析 latest。OpenAI provider<br/> 不使用它。"]
  h --- n2["Secret Manager 名稱：<br/>line-messaging-api-channel-access-token<br/>容器環境變數：<br/>LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN<br/>部署檔參照：2<br/>2026-11-08 歷史快照／用途：綁定 2；LINE <br/>Reply／Push 使用。舊版 1 保留。"]
  h --- n3["Secret Manager 名稱：<br/>line_mini_app_channel_secret<br/>容器環境變數：LINE_MINI_APP_CHANNEL_SECRET<br/>部署檔參照：1<br/>2026-11-08 歷史快照／用途：未綁定；目前 <br/>server 未讀取，僅為未來 Service Message 預留<br/>。"]
  h --- n4["Secret Manager 名稱：<br/>line-messaging-api-channel-secret<br/>容器環境變數：<br/>LINE_MESSAGING_API_CHANNEL_SECRET<br/>部署檔參照：3<br/>2026-11-08 歷史快照／用途：綁定 3；LINE <br/>Webhook 驗簽使用。版本 2 帶有多餘換行；版本 <br/>1 已停用。"]
  h --- n5["Secret Manager 名稱：openai-api-key<br/>容器環境變數：OPENAI_API_KEY<br/>部署檔參照：1<br/>2026-11-08 歷史快照／用途：綁定 1；目前 <br/>OpenAI 摘要使用。"]
```

```mermaid
flowchart LR
  h["Secret Manager 與環境變數"]
  h --- n0["Secret 名稱存在不代表已有 enabled version。<br/>新增或輪替 version 後仍須更新 Cloud Run <br/>reference 並建立 Ready revision；不得用文件<br/>或 Secret 名稱推定 payload 正確。"]
  h --- n1["_SUMMARY_PROVIDER=openai 與 <br/>_SUMMARY_MODEL=gpt-4.1-mini 只在執行 <br/>deploy/cloudbuild.yaml 時透過 --set-env-vars<br/> 成為 runtime 設定。摘要還需要 <br/>SUMMARY_TASK_QUEUE、SUMMARY_TASK_URL、<br/>SUMMARY_TASK_SERVICE_ACCOUNT。現行服務已由 <br/>Cloud Run service configuration 成套設定 <br/>OpenAI provider；queue 層級也已授予 runtime <br/>service account roles/cloudtasks.enqueuer，<br/>但仍須以真實任務驗收 OIDC、模型與 LINE Push<br/>。"]
```

```mermaid
flowchart LR
  h["Secret Manager 與環境變數"]
  h --- n0["Cloud Build substitution：<br/>_MESSAGING_API_CHANNEL_ID<br/>注入位置／用途：<br/>LINE_MESSAGING_API_CHANNEL_ID；必填<br/>預設：2012305727"]
  h --- n1["Cloud Build substitution：_MINI_APP_ENV<br/>注入 LINE_MINI_APP_ENV，選擇 developing、review 或 published，並注入該階段公開三元組；預設 developing。"]
  h --- n3["Cloud Build substitution：<br/>_CHANNEL_SECRET_NAME<br/>注入位置／用途：<br/>LINE_MESSAGING_API_CHANNEL_SECRET 的 secret <br/>名稱<br/>預設：line-messaging-api-channel-secret"]
  h --- n4["Cloud Build substitution：<br/>_CHANNEL_SECRET_VERSION<br/>注入位置／用途：上述 secret 的版本<br/>預設：3"]
  h --- n5["Cloud Build substitution：<br/>_CHANNEL_ACCESS_TOKEN_NAME<br/>注入位置／用途：<br/>LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN 的 <br/>secret 名稱<br/>預設：<br/>line-messaging-api-channel-access-token"]
  h --- n6["Cloud Build substitution：<br/>_CHANNEL_ACCESS_TOKEN_VERSION<br/>注入位置／用途：上述 secret 的版本<br/>預設：2"]
```

```mermaid
flowchart LR
  h["Secret Manager 與環境變數"]
  h --- n0["使用版本 1 前須確認已存在且 enabled；若實際<br/>版本不同，明確覆寫。 目前 preflight 僅對 <br/>Messaging Channel secret／access token 檢查<br/>正整數版本； 它沒有檢查其他四個 Secret <br/>references，且 YAML 的 Gemini 預設仍是 <br/>latest。 手動發布前須核對所有綁定並使用明確<br/>數字版本；不能把 preflight 通過視為全部機密<br/>已就緒。 輪替時新增 secret version，再帶新的<br/>數字版本部署；保留舊版本供回滾使用。"]
  h --- n1["GOOGLE_CLOUD_PROJECT 自動設為此次 Build 的 <br/>project。MINI App 兩個 ID 全空時停用身分功能<br/>； 只提供其中一個會在建置前失敗。Channel、<br/>LIFF ID 都不是機密，但必須符合實際環境。 <br/>Service Message token、MCP API key、使用者 <br/>ID token 與 ADC 憑證不屬於本部署的 <br/>substitutions。 應用設定命名仍依 LINE 契約。"]
  click n1 "../integrations/line/configuration.md"
  h --- n2["deploy/cloudbuild.yaml 使用 set-env-vars／<br/>set-secrets 宣告完整集合，會移除未列出的手動<br/> env／secret 綁定； 新增功能時先更新檔案，不<br/>能只在 Console 臨時補值。這些值都在執行期注<br/>入，沒有 bake 進映像。"]
```

## MINI App 公開設定遷移

公開鍵名與驗證規則見 [LINE 環境設定](../integrations/line/configuration.md#公開環境設定)。根 `.env.example` 與手動 Cloud Build substitutions 不會由 managed image-only trigger 自動匯入服務。切換至不含硬編碼 fallback 的映像前，先以 Cloud Run `--update-env-vars` 補齊所選組三元組及 selector；不要使用會覆蓋其他環境變數的 `--set-env-vars`，也不要更動現有 Secret references。此遷移會建立 revision，必須在取得部署授權後另行執行及驗收；本機測試或 PR CI 不代表遠端已遷移。

手動 Cloud Build 範例以 `^|^` 切換 gcloud dictionary 分隔符，因此 Endpoint query 中的逗號不會被拆為其他變數；URL 若需要 pipe 字元，使用 URL 編碼 `%7C`。一般命令的逗號分隔規則與替代 delimiter 語法以 `gcloud topic escaping` 為準。
