# 正式環境設定

此頁維護 Cloud Run runtime 設定、Secret Manager 綁定與公開設定遷移。[部署步驟](cloud-run.md)與[發布驗收](release-acceptance.md)另列。

正式環境將 LINE 憑證與 `OPENAI_API_KEY`、`GEMINI_API_KEY`、`CWA_API_KEY` 綁定為 Secret Manager references。先核對目標版本已啟用，再建立並驗證使用該 reference 的 Ready revision。Managed Trigger 若只更新映像，會沿用既有服務設定。服務目前的綁定與值以目標環境為準。

<a id="設定契約"></a>

<a id="secret-manager-與環境變數"></a>

## Secret Manager 與環境變數

[Secret Manager 與兩條部署路徑的關係如下。cloudbuild.yaml 的「部署檔參照」只在手動或 自訂 Trigger 真正執行該檔時生效；「歷史快照」記錄 2026-11-08 revision line-mini-app-bot-00027-695 的 metadata，不包含 Secret payload。此表不是現行服務宣告； 部署前使用發布 snapshot重新讀取 serving revision、 流量、執行帳戶及 Secret references，另核對必要非機密 env 值：](release-acceptance.md#發布證據與前置條件)

| Secret Manager 名稱 | 容器環境變數 | 部署檔參照 | 2026-11-08 歷史快照／用途 |
| --- | --- | --- | --- |
| cwa-api-key | CWA_API_KEY | 1 | 綁定 1；天氣功能使用。 |
| gemini-api-key | GEMINI_API_KEY | 提交時必填正整數版本 | 未綁定；Secret 尚無 version，不能解析 latest。OpenAI provider 不使用它。 |
| line-messaging-api-channel-access-token | LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN | 2 | 綁定 2；LINE Reply／Push 使用。舊版 1 保留。 |
| line_mini_app_channel_secret | LINE_MINI_APP_CHANNEL_SECRET | 1 | 未綁定；目前 server 未讀取，僅為未來 Service Message 預留。 |
| line-messaging-api-channel-secret | LINE_MESSAGING_API_CHANNEL_SECRET | 3 | 綁定 3；LINE Webhook 驗簽使用。版本 2 帶有多餘換行；版本 1 已停用。 |
| openai-api-key | OPENAI_API_KEY | 1 | 綁定 1；目前 OpenAI 摘要使用。 |

- Secret 名稱存在不代表已有 enabled version。新增或輪替 version 後仍須更新 Cloud Run reference 並建立 Ready revision；不得用文件或 Secret 名稱推定 payload 正確。

- _SUMMARY_PROVIDER=openai 與 _SUMMARY_MODEL=gpt-4.1-mini 只在執行 cloudbuild.yaml 時透過 --set-env-vars 成為 runtime 設定。摘要還需要 SUMMARY_TASK_QUEUE、SUMMARY_TASK_URL、SUMMARY_TASK_SERVICE_ACCOUNT。現行服務已由 Cloud Run service configuration 成套設定 OpenAI provider；queue 層級也已授予 runtime service account roles/cloudtasks.enqueuer，但仍須以真實任務驗收 OIDC、模型與 LINE Push。

- Cloud Build substitution：_MESSAGING_API_CHANNEL_ID<br>注入位置／用途：LINE_MESSAGING_API_CHANNEL_ID；必填<br>預設：2012305727

- Cloud Build substitution：_MINI_APP_ENV注入 LINE_MINI_APP_ENV，選擇 developing、review 或 published，並注入該階段公開三元組；預設 developing。

- Cloud Build substitution：_CHANNEL_SECRET_NAME<br>注入位置／用途：LINE_MESSAGING_API_CHANNEL_SECRET 的 secret 名稱<br>預設：line-messaging-api-channel-secret

- Cloud Build substitution：_CHANNEL_SECRET_VERSION<br>注入位置／用途：上述 secret 的版本<br>預設：3

- Cloud Build substitution：_CHANNEL_ACCESS_TOKEN_NAME<br>注入位置／用途：LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN 的 secret 名稱<br>預設：line-messaging-api-channel-access-token

- Cloud Build substitution：_CHANNEL_ACCESS_TOKEN_VERSION<br>注入位置／用途：上述 secret 的版本<br>預設：2

- 使用版本 1 前須確認已存在且 enabled；若實際版本不同，明確覆寫。 目前 preflight 對全部六個 Secret 綁定檢查正整數版本；Gemini 沒有預設版本，提交建置時必填 _GEMINI_API_KEY_VERSION。 手動發布前須核對所有綁定並使用明確數字版本；不能把 preflight 通過視為全部機密已就緒。 輪替時新增 secret version，再帶新的數字版本部署；保留舊版本供回滾使用。

- [GOOGLE_CLOUD_PROJECT 自動設為此次 Build 的 project。MINI App 兩個 ID 全空時停用身分功能； 只提供其中一個會在建置前失敗。Channel、LIFF ID 都不是機密，但必須符合實際環境。 Service Message token、MCP API key、使用者 ID token 與 ADC 憑證不屬於本部署的 substitutions。 應用設定命名仍依 LINE 契約。](../integrations/line/configuration.md)

- cloudbuild.yaml 使用 set-env-vars／set-secrets 宣告完整集合，會移除未列出的手動 env／secret 綁定； 新增功能時先更新檔案，不能只在 Console 臨時補值。這些值都在執行期注入，沒有 bake 進映像。

## MINI App 公開設定遷移

公開鍵名與驗證規則見 [LINE 環境設定](../integrations/line/configuration.md#公開環境設定)。根 `.env.example` 與手動 Cloud Build substitutions 不會由 managed image-only trigger 自動匯入服務。切換至不含硬編碼 fallback 的映像前，先以 Cloud Run `--update-env-vars` 補齊所選組三元組及 selector；不要使用會覆蓋其他環境變數的 `--set-env-vars`，也不要更動現有 Secret references。此遷移會建立 revision，必須在取得部署授權後另行執行及驗收；本機測試或 PR CI 不代表遠端已遷移。

手動 Cloud Build 範例以 `^|^` 切換 gcloud dictionary 分隔符，因此 Endpoint query 中的逗號不會被拆為其他變數；URL 若需要 pipe 字元，使用 URL 編碼 `%7C`。一般命令的逗號分隔規則與替代 delimiter 語法以 `gcloud topic escaping` 為準。
