# Go LINE 接入能力

以 kkdai 的 Go 範例為參考，沿用官方 Go SDK v8.23.0。此階段提供連線能力，
尚無出勤寫入、持久去重、群組角色管理、Service Message 或背景工作。

## 啟動

```bash
bash scripts/tooling/tool-exec.sh mise run dev
```

PORT 預設 8080，不自動讀 .env。保留 /location 定位示範，沒有健康檢查路由。
Messaging API 與 MINI App 分別啟用；各組全缺時停用、部分缺失時啟動失敗。
設定名稱與 Channel 分離規則見 [LINE 環境變數](../integrations/line-environment.md)。

## Messaging API、一對一與群組

注入 LINE_MESSAGING_API_CHANNEL_ID、LINE_MESSAGING_API_CHANNEL_SECRET、
LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN。Cloud Run 透過 Secret Manager 注入機密。
Webhook URL 為 `<service-url>/webhooks/line/<channelID>`。

| 操作 | 一對一 | 群組 |
| --- | --- | --- |
| /help、/menu、選單、功能 | Flex 功能選單 | 公開 Flex 功能選單 |
| /status | 一對一連線狀態 | 公開群組連線狀態 |
| 其他文字 | 顯示選單 | 忽略，避免干擾群組對話 |
| Postback action=help／action=status | 依實際事件來源回覆 | 依實際事件來源回覆 |

在 LINE Developers 開啟 Webhook；群組使用還需允許官方帳號加入群組。
一般文字不會被記錄或回顯。群組不查個人資料，不依據 Postback 的 userId／groupId
授權，也不把成員身分視為群組管理員。未知 action／額外 Postback 欄位直接忽略。
MINI App 完成設定後，Flex 選單會加入開啟 MINI App 的按鈕。

body 上限 1 MiB；過大回 413、錯誤簽章回 401、無效輸入回 400。
空事件回 200，standby、redelivery、非文字、room 及不支援事件忽略。
整批發送共用 5 秒期限；失敗僅記錄固定警告，不重試，也不改用 Push，Webhook 回 200。
200 不代表訊息送達；跳過 redelivery 不等於可靠接收或完整冪等性。

## MINI App：Go + Next.js／React

注入配對的 LINE_MINI_APP_CHANNEL_ID 與 LINE_MINI_APP_LIFF_ID，Channel 必須啟用
openid scope；要顯示名稱時再啟用 profile。設定 MINI App Endpoint URL 為
`https://<service-host>/miniapp`，與部署環境的 Developing／Review／Published 對應。
已設定的 MINI App 永久連結為 `https://miniapp.line.me/<LIFF ID>`。

1. Go 提供 Next.js 靜態輸出的 /miniapp 頁面與 React 程式，載入 LINE 官方 LIFF v2 SDK。
2. 前端以 /api/v1/miniapp/config 取得公開 LIFF ID，執行 liff.init。
3. 外部瀏覽器由使用者點擊登入，redirectUri 保留本站目前頁面與查詢參數。
4. 點擊身分驗證後，前端取得 liff.getIDToken()，以 Bearer header 呼叫 GET /api/v1/me。
5. Go 將 ID token 與預期 Channel ID POST 至 LINE /oauth2/v2.1/verify，
   並檢查回應 issuer、audience、非空 subject 與到期時間。

不接受前端 profile、decoded token 或任意 userId 作身分。成功只回傳本人
userId／displayName，設定 no-store，不寫入資料庫或一般日誌。
每次請求都驗證 token，不建立 session cookie、localStorage token 或群組權限。
此身分驗證不自動授予出勤資格；跨 Channel 的 user ID 對應也不能自行假設。

/api/v1/me 未設定回 503；未登入／過期／錯 Channel 回 401；
上游逾時／無效回應回 503。錯誤為 `{"error":{"code":"...","message":"..."}}`。
登出清除本頁狀態，LIFF 瀏覽器可能再次自動登入；這不等於撤銷 LINE 帳號授權。

程式接入不代表 MINI App 已完成 Channel 建立、資格確認或發布審查。
本機可看頁面與未設定提示；真實登入需要正確的 HTTPS Endpoint 與 Channel 設定。

## Rich Menu：Go CLI

Rich Menu 用於官方帳號的一對一聊天；群組使用 Flex 指令選單。
範本為 2500 x 843，左至右依序為「功能說明」、「連線狀態」、「開啟 MINI App」。
先準備相同尺寸、對應三個區塊的 PNG／JPEG，檔案不超過 1 MiB。

只預覽 JSON，不發送外部請求：

```bash
bash scripts/tooling/tool-exec.sh go run ./cmd/richmenu -app-url https://miniapp.line.me/LIFF_ID
```

已確認選單圖片、目標官方帳號與替換影響後，由操作者執行：

```bash
bash scripts/tooling/tool-exec.sh go run ./cmd/richmenu -app-url https://miniapp.line.me/LIFF_ID -image /absolute/path/menu.png -apply
```

需要 LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN。此命令建立選單並上傳圖片，
不預設綁定；加 -set-default 才替換該官方帳號的預設選單。
每次 apply 若未提供 -menu-id 就會建立新選單，並非冪等部署。
建立後會印出 richMenuId；上傳或綁定失敗時保留該 ID，可用 -menu-id 續做。
沒有自動刪除或重試；結果不明時先從 LINE Console／API 確認，避免重複建立。
替換前記錄原本預設 ID，回復時使用 LINE 官方設定預設選單 API；
不要假設 API 選單可在 Official Account Manager 編輯。

## Firestore 唯讀連線

設定 GOOGLE_CLOUD_PROJECT、選用 FIRESTORE_DATABASE_ID（預設 (default)）與 Google ADC。
Cloud Run 使用執行服務帳戶，本機使用自己的 ADC。指定一份允許讀取的既有文件：

```bash
bash scripts/tooling/tool-exec.sh go run ./cmd/firestorecheck -document COLLECTION/DOCUMENT
```

成功只表示該文件的讀取權限與連線可用，不顯示內容、不建立資料。
NotFound 不當作通過，需確認 database 與文件存在；此檢查不證明寫入／transaction 權限。
HTTP 服務仍未接入出勤持久化，不能把工具通過當作業務功能完成。

## 查核來源與限制

- [kkdai/LineBotTemplate](https://github.com/kkdai/LineBotTemplate)
- [kkdai/line-card-assistant](https://github.com/kkdai/line-card-assistant/tree/main/internal/linebot)
- [kkdai/liff-template-go](https://github.com/kkdai/liff-template-go)
- [kkdai/linebot-garbage-helper](https://github.com/kkdai/linebot-garbage-helper)
- [LINE 官方來源](https://github.com/line/line-developers-docs-source/tree/090500a12f9777995006d9fb5cafab8d83a792b2)，
  核對 LIFF 後端 token 驗證、Rich Menu 建立／圖片／預設綁定與 MINI App 永久連結。
- Context7 /line/line-bot-sdk-go/v8.23.0 與相同版本本機 SDK 核對 client、Context、Rich Menu 方法與型別。

真實 LINE 收發、手機 LIFF、Rich Menu 套用與 Firestore 連線仍須渠道與憑證驗收；
本機檢查不取代這些驗收。沒有新增版本庫測試方案或工具。
