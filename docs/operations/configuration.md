# 設定管理

設定分為公開 runtime 值、秘密、建置參數與本機測試值。每個值必須有單一來源、明確型別、允許範圍與安全預設。

## Runtime 設定

主要群組如下：

| 類型           | 例子                                                 |
| -------------- | ---------------------------------------------------- |
| 服務           | `PORT`、`FRONTEND_DIR`                               |
| Google Cloud   | `GOOGLE_CLOUD_PROJECT`、`FIRESTORE_DATABASE_ID`      |
| LINE Messaging | channel ID、channel secret、access token             |
| MINI App       | environment 及各階段的 channel ID、LIFF ID、endpoint |
| 摘要／Agent    | provider、model、queue、target、service account      |
| 外部供應端     | OpenAI、Gemini、CWA keys                             |
| 功能開關       | Pub/Sub topic、Rich Menu image directory 等          |

完整名稱與實際讀取條件以 `backend/internal/bootstrap/`、`cloudbuild.yaml` 與 `.env.example` 為準。

## 秘密

Token、channel secret 與 API key 只透過 Secret Manager 或受控本機環境注入。部署使用明確數字版本，不使用不受控的 `latest`。輪替流程先新增版本、建立 Ready revision、驗證後切流量，再依回退窗口停用舊版本。

## 驗證與優先序

設定在 bootstrap 邊界解析；缺少成組值、型別錯誤或不支援的 environment 必須啟動失敗並指出設定名稱。映像內預設只可用於非機密且安全的值。Cloud Run service configuration 可能與儲存庫部署檔不同，正式操作前必須讀取目標 revision 的實際 metadata。

## 安全

不得把 secret payload、OAuth URL、驗證碼或 ADC 檔寫入文件、提交或一般 log。公開 channel／LIFF ID 雖非秘密，仍需按環境一致配置，不能當授權證據。

## 群聊摘要啟用條件

沿用 GOOGLE_CLOUD_PROJECT、LINE_MESSAGING_API_CHANNEL_ID、LINE_MESSAGING_API_CHANNEL_SECRET、LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN、FIRESTORE_DATABASE_ID 與 runtime ADC。 以下 provider、選用的 API key、模型與三項 task 設定全部未設定時，摘要模組不啟用；設定任何一項後則檢查完整性。這是模組部署條件，群組使用時無需再逐一開啟。

| 環境變數                     | 值                                                                                                                                    |
| ---------------------------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| SUMMARY_PROVIDER             | openai 或 gemini；未設定時沿用 Gemini。                                                                                               |
| OPENAI_API_KEY               | 選用 OpenAI 時，由 Secret Manager 的 openai-api-key 注入。                                                                            |
| GEMINI_API_KEY               | 選用 Gemini 時，由 Secret Manager 注入。                                                                                              |
| SUMMARY_MODEL                | 該 key／專案可使用、支援對應 provider API 的文字模型名稱（OpenAI 使用 Chat Completions、Gemini 使用 generateContent）；不硬編碼版本。 |
| SUMMARY_TASK_QUEUE           | projects/PROJECT_ID/locations/REGION/queues/QUEUE_ID。                                                                                |
| SUMMARY_TASK_URL             | 本服務 https://HOST/internal/summary，不能有 query 或 fragment。                                                                      |
| SUMMARY_TASK_SERVICE_ACCOUNT | Cloud Tasks 發送 OIDC token 使用的專用服務帳戶 email。                                                                                |

[摘要服務設定](../../backend/internal/bootstrap/summary.go)

## 群組天氣查詢

- 部署環境注入 CWA_API_KEY（氣象資料開放平臺會員授權碼）；未設定時回覆尚未設定。 API Key 不放進前端或日誌。此功能不需要 Firestore，不影響出勤的設定。

- 每次分享查詢一次 API，3 秒逾時、4 MiB 回應上限，不自動跟隨重新導向。第一版沒有快取或背景工作。 重送位置事件不再次查詢；回覆失敗不保證送達，使用者可重新分享。

[氣象署查詢](../../backend/internal/business/weather/adapter/cwa/forecast/cwa.go) · [LINE 接線](../../backend/internal/bootstrap/line.go)
