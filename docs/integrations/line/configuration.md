# LINE 環境變數規格

<a id="line-環境變數規格"></a>

環境變數使用全大寫與底線，依產品區分前綴，不共用模糊的 `LINE_CHANNEL_*` 名稱。Messaging API 三項與 MINI App 環境選擇已供接入能力讀取與檢查；Service Message token 尚未使用。接入與部署見 [LINE 設定指南](setup.md)。

<a id="1-命名與來源"></a>

## 1. 命名與來源

| 前綴                | 所屬產品           | 規則                                                |
| ------------------- | ------------------ | --------------------------------------------------- |
| LINE_MESSAGING_API_ | LINE Messaging API | 官方帳號、Webhook、Reply／Push 與 Rich Menu         |
| LINE_MINI_APP_      | LINE MINI App      | LIFF、身分驗證、MINI App Channel 與 Service Message |

Go 欄位可對應為 MessagingAPI 與 MiniApp 兩組設定，禁止兩組憑證互相回退。 本機、測試與正式部署沿用相同變數名稱，由不同部署環境注入各自的值。

<a id="2-messaging-api"></a>

## 2. Messaging API

| 環境變數                                | 用途                               | 機密 | 必要條件                  |
| --------------------------------------- | ---------------------------------- | ---- | ------------------------- |
| LINE_MESSAGING_API_CHANNEL_ID           | 明確識別 Messaging API Channel     | 否   | 啟用 Messaging API 時     |
| LINE_MESSAGING_API_CHANNEL_SECRET       | 驗證 Webhook 簽章                  | 是   | 接收 Webhook 時           |
| LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN | 呼叫 Reply／Push、Rich Menu 等 API | 是   | 使用預先簽發 token 發送時 |

Channel access token 須依官方選定的 token 類型管理有效期與輪替；若改用即時簽發， 另行設計金鑰與簽發設定，不把 client assertion 或私鑰塞入 access token 欄位。

<a id="3-mini-app"></a>

## 3. MINI App

以 `LINE_MINI_APP_ENV` 選擇部署階段。各階段的 Channel ID、LIFF ID、Endpoint URL、啟動驗證與舊鍵相容性見[公開環境設定](#公開環境設定)。

啟用 MINI App Service Message 且採預先簽發 token 時，使用機密變數 `LINE_MINI_APP_CHANNEL_ACCESS_TOKEN` 呼叫 Service Message API。

- MINI App 的 Developing／Review／Published 有不同內部 Channel；每個部署只選擇一個環境的 Channel／LIFF 設定，不接受其他環境的 token，不得混用。Messaging API 與 MINI App 即使位於相同 Provider，也使用各自的 Channel 憑證。

- 目前沒有使用 MINI App Channel secret 的流程，因此不預設加入該變數。 未來若選定的授權流程需要，再明確增加 LINE_MINI_APP_CHANNEL_SECRET。

- Quick Reply的「MINI App」先送既有Webhook postback action=miniapp，回覆一張Flex開啟卡；手機再點卡片後使用原MINI App URI，比原先多一次點擊。桌面LINE使用altUri.desktop開啟已設定的HTTPS網頁入口；Quick Reply本身不支援此屬性，Rich Menu與其他quick chips不變。功能說明卡的MINI App按鈕共用同一桌面分流規則。

- Developing、Review、Published 共用 https://line-mini-app-bot-jtghegqd7q-de.a.run.app/miniapp。這是公開網頁 Endpoint，不取代各環境 miniapp.line.me 永久連結，也不授予登入或群組權限。本次僅固定程式設定，不代表已部署或完成桌面 SSO 驗收。

- 依據：官方URI action與外部瀏覽器，官方來源庫revision c15d6d3fa2eff7fcae65554c20be5b1e60f9a10a。桌面需要手動LINE登入，不能把網頁入口當授權或繞過目前成員核驗。

  - [官方URI action](https://developers.line.biz/en/reference/messaging-api/#uri-action)

  - [外部瀏覽器](https://developers.line.biz/en/docs/line-mini-app/develop/external-browser/)

<a id="4-不放進固定環境變數的資料"></a>

## 4. 不放進固定環境變數的資料

| 資料                                    | 正確位置                                             |
| --------------------------------------- | ---------------------------------------------------- |
| 使用者 LINE ID token／LIFF access token | 請求或安全的使用者 session；按流程驗證，不記錄於日誌 |
| Service notification token              | 按使用者操作／session 安全保存，發送後更新最新值     |
| Webhook reply token                     | 短效事件處理資料，不作全域設定或長期排程憑證         |
| user ID、group ID                       | 具存取控制的業務資料，不作服務共用認證               |

<a id="5-注入與驗證"></a>

## 5. 注入與驗證

- mise.toml 管理工具與非機密共用預設；真實 Channel／LIFF ID 由部署環境提供，避免提交某個環境的固定值。

- [本機由執行環境安全注入憑證。Cloud Run 注入、機密輪替及 revision 驗收見正式環境設定。](../../operations/configuration.md)

- 前端只公開必要的 LIFF ID 等非機密資訊；所有 Channel token／secret 限後端使用。

- 啟動時依啟用功能檢查必填值，錯誤只列變數名稱，不輸出值；不讀取舊 LINE_CHANNEL_* 別名。

- Channel ID 與 LIFF ID 使用字串，不能互換；群組來源 ID 也不能拿來作使用者授權。

- 單靠變數名稱不能證明 token 屬於正確 Channel；部署驗收需檢查環境對應並測試實際授權。

- [整合行為與資格限制見 LINE 能力矩陣。](capabilities.md)

## 公開環境設定

根目錄 [.env.example](../../../.env.example) 列出十個公開設定變數；[Go MINI App 設定](../../../backend/internal/bootstrap/miniapp.go) 是唯一讀取與驗證入口。範例值可替換為其他產品的 Channel、LIFF 與 Endpoint，程式不內建個人 ID 或主機。

| LINE_MINI_APP_ENV  | 所選變數前綴              |
| ------------------ | ------------------------- |
| developing（預設） | LINE_MINI_APP_DEVELOPING_ |
| review             | LINE_MINI_APP_REVIEW_     |
| published          | LINE_MINI_APP_PUBLISHED_  |

每個前綴包含 `CHANNEL_ID`、`LIFF_ID`、`ENDPOINT_URL`。僅所選組必填：Channel ID 為數字字串；LIFF ID 非空且與 Channel ID 前綴配對；Endpoint 必須是無 credentials／fragment 的絕對 HTTPS URL，可保留路徑與 query。未知階段或所選組缺失／非法值會在啟動時回報變數名稱，不輸出設定內容。未選組不影響啟動。

公開 config API、ID token verifier 與 Session Channel 使用同一選擇，不接受其他環境 audience。手機永久連結仍由所選 LIFF ID 產生 `https://miniapp.line.me/<LIFF ID>`；桌面 Flex 使用該組 Endpoint，不能從請求 Host 或 postback 改寫。Messaging API 設定獨立。

`.env` 不自動載入；本機由 shell／執行環境安全注入。舊 `LINE_MINI_APP_CHANNEL_ID`、`LINE_MINI_APP_LIFF_ID`、`LINE_MINI_APP_WEB_URL` 不再生效，也不作 fallback。公開設定不表示已完成 LINE Console、openid、HTTPS、ADC 或 Firestore 權限驗收。

正式環境的注入方式、公開設定遷移與 revision 驗收見[正式環境設定](../../operations/configuration.md#mini-app-公開設定遷移)。
