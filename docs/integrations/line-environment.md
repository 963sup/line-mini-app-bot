# LINE 環境變數規格

> 所有名稱使用全大寫與底線。依產品區分前綴，不共用模糊的 `LINE_CHANNEL_*` 名稱。
> Messaging API 三項設定已供 [接入試作](../development/line-prototype.md) 讀取與檢查；MINI App 設定尚未實作。

## 1. 命名與來源

| 前綴 | 所屬產品 | 規則 |
| --- | --- | --- |
| `LINE_MESSAGING_API_` | LINE Messaging API | 官方帳號、Webhook、Reply／Push 與 Rich Menu |
| `LINE_MINI_APP_` | LINE MINI App | LIFF、身分驗證、MINI App Channel 與 Service Message |

Go 欄位可對應為 `MessagingAPI` 與 `MiniApp` 兩組設定，禁止兩組憑證互相回退。
本機、測試與正式部署沿用相同變數名稱，由不同部署環境注入各自的值。

## 2. Messaging API

| 環境變數 | 用途 | 機密 | 必要條件 |
| --- | --- | --- | --- |
| `LINE_MESSAGING_API_CHANNEL_ID` | 明確識別 Messaging API Channel | 否 | 啟用 Messaging API 時 |
| `LINE_MESSAGING_API_CHANNEL_SECRET` | 驗證 Webhook 簽章 | 是 | 接收 Webhook 時 |
| `LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN` | 呼叫 Reply／Push、Rich Menu 等 API | 是 | 使用預先簽發 token 發送時 |

Channel access token 須依官方選定的 token 類型管理有效期與輪替；若改用即時簽發，
另行設計金鑰與簽發設定，不把 client assertion 或私鑰塞入 access token 欄位。

## 3. MINI App

| 環境變數 | 用途 | 機密 | 必要條件 |
| --- | --- | --- | --- |
| `LINE_MINI_APP_CHANNEL_ID` | 後端驗證 LINE ID token 的預期 Channel／audience | 否 | 使用 MINI App 身分驗證時 |
| `LINE_MINI_APP_LIFF_ID` | 前端 `liff.init` 的 LIFF ID | 否 | 啟動 MINI App 前端時 |
| `LINE_MINI_APP_CHANNEL_ACCESS_TOKEN` | 呼叫 MINI App Service Message API | 是 | 啟用 Service Message 且採預先簽發 token 時 |

MINI App 的 Developing／Review／Published 有不同內部 Channel；每個部署只能注入
該環境匹配的 Channel ID、LIFF ID 與 token，不得混用。Messaging API 與 MINI App
即使位於相同 Provider，也使用各自的 Channel 憑證。

目前沒有使用 MINI App Channel secret 的流程，因此不預設加入該變數。
未來若選定的授權流程需要，再明確增加 `LINE_MINI_APP_CHANNEL_SECRET`。

## 4. 不放進固定環境變數的資料

| 資料 | 正確位置 |
| --- | --- |
| 使用者 LINE ID token／LIFF access token | 請求或安全的使用者 session；按流程驗證，不記錄於日誌 |
| Service notification token | 按使用者操作／session 安全保存，發送後更新最新值 |
| Webhook reply token | 短效事件處理資料，不作全域設定或長期排程憑證 |
| user ID、group ID | 具存取控制的業務資料，不作服務共用認證 |

## 5. 注入與驗證

- `mise.toml` 管理工具與非機密共用預設；真實 Channel／LIFF ID 由部署環境提供，避免提交某個環境的固定值。
- Cloud Run 正式環境機密由 Secret Manager 綁定。
  本機由執行環境安全注入，不提交至 mise、範例檔或版本庫。
- 前端只公開必要的 LIFF ID 等非機密資訊；所有 Channel token／secret 限後端使用。
- 啟動時依啟用功能檢查必填值，錯誤只列變數名稱，不輸出值；不讀取舊 `LINE_CHANNEL_*` 別名。
- Channel ID 與 LIFF ID 使用字串，不能互換；群組來源 ID 也不能拿來作使用者授權。
- 單靠變數名稱不能證明 token 屬於正確 Channel；部署驗收需檢查環境對應並測試實際授權。

整合行為與資格限制見 [LINE 能力矩陣](line.md)。
