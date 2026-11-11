# Go LINE 接入能力

以 kkdai 的 Go 範例為參考，使用官方 Go SDK；依賴版本以 backend/go.mod 為準。
目前已接入 LINE Webhook、MINI App Session、Firestore 出勤與請假登記，以及 Cloud Tasks 摘要工作。
功能規則分別見[出勤](../features/attendance.md)、[請假](../features/leave.md)、
[看板與漏下班](../features/attendance-board.md)及[摘要](../features/summary.md)。
群組角色管理、Service Message 與常駐 Worker 不屬於目前實作。

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
| /help | Flex 功能選單 | 公開 Flex 功能選單 |
| /menu、選單、功能 | 快捷選單 | 打卡、出勤、請假、漏下班、摘要快捷選單 |
| 其他文字 | 顯示選單 | 忽略，避免干擾群組對話 |
| Postback action=help | 依實際事件來源回覆 | 依實際事件來源回覆 |

在 LINE Developers 開啟 Webhook；群組使用還需允許官方帳號加入群組。
摘要啟用時會依[摘要收集規則](../features/summary.md)保存群組一般文字，不能將忽略回覆理解為未保存。
不依據前端或 Postback 自行提供的 userId／groupId 授權，也不把成員身分視為群組管理員。
未知 action／額外 Postback 欄位直接忽略。
MINI App 完成設定後，Flex 選單會加入開啟 MINI App 的按鈕。

body 上限 1 MiB；過大回 413、錯誤簽章回 401、無效輸入回 400。
不同模組對事件保存、重送及失敗的處理不同；以[出勤](../features/attendance.md)與
[摘要](../features/summary.md)契約為準，不再套用早期連線試作的「全部忽略 redelivery」說明。
HTTP 200、revision Ready 與訊息送達須分別驗證。

## MINI App：Go + Next.js／React

注入配對的 LINE_MINI_APP_CHANNEL_ID 與 LINE_MINI_APP_LIFF_ID，Channel 必須啟用
openid scope；要顯示名稱時再啟用 profile。設定 MINI App Endpoint URL 為
`https://<service-host>/miniapp`，與部署環境的 Developing／Review／Published 對應。
已設定的 MINI App 永久連結為 `https://miniapp.line.me/<LIFF ID>`。

1. Go 提供 Next.js 靜態輸出的 /miniapp 頁面與 React 程式，載入 LINE 官方 LIFF v2 SDK。
2. 前端以 /api/v1/miniapp/config 取得公開 LIFF ID，執行 liff.init。
3. 外部瀏覽器由使用者點擊登入，redirectUri 保留本站目前頁面與查詢參數。
4. 前端取得 liff.getIDToken()，以 Bearer header 呼叫 POST /api/v1/session，body 為空物件。
5. Go 向 LINE 驗證 token 與預期 Channel，成功後建立 Firestore Session，回傳安全 Cookie、本人資料、CSRF token 與有效期限。
6. 後續本人與出勤 API 透過 SessionProvider.request 使用 Cookie／CSRF；需要近期活動且接近到期才續期。

不接受前端 profile、decoded token 或任意 userId 作身分。Session ID 透過 HttpOnly Cookie
傳遞，CSRF token 僅保留於記憶體；原始 Session ID、CSRF 與 LINE token 不存入資料庫或日誌。
Session 並不授予任意群組或他人資料的存取權。

登出呼叫 DELETE /api/v1/session 撤銷伺服器狀態並清除 Cookie；失敗時保留重試入口，
不能只清畫面就宣稱已登出。撤銷應用 Session 不等於撤銷 LINE 帳號授權。
期限、續期、Cookie 優先順序、既有 Bearer 相容性及錯誤處理以
[Session 契約](../architecture/browser-session.md)為單一來源；路徑與請求格式見 [API 契約](../../api/README.md)。

程式接入不代表 MINI App 已完成 Channel 建立、資格確認或發布審查。
本機可看頁面與未設定提示；真實登入需要正確的 HTTPS Endpoint 與 Channel 設定。

## Rich Menu：Go CLI

Rich Menu 用於官方帳號的一對一聊天；群組使用 Flex 指令選單。
範本為 2500 x 843，左至右依序為「功能說明」、「出勤紀錄」、「開啟 MINI App」。
先準備相同尺寸、對應三個區塊的 PNG／JPEG，檔案不超過 1 MiB。

只預覽 JSON，不發送外部請求：

```bash
bash scripts/tooling/tool-exec.sh go -C backend run ./cmd/richmenu -app-url https://miniapp.line.me/LIFF_ID
```

已確認選單圖片、目標官方帳號與替換影響後，由操作者執行：

```bash
bash scripts/tooling/tool-exec.sh go -C backend run ./cmd/richmenu -app-url https://miniapp.line.me/LIFF_ID -image /absolute/path/menu.png -apply
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
bash scripts/tooling/tool-exec.sh go -C backend run ./cmd/firestorecheck -document COLLECTION/DOCUMENT
```

成功只表示該文件的讀取權限與連線可用，不顯示內容、不建立資料。
NotFound 不當作通過，需確認 database 與文件存在；此檢查不證明寫入／transaction 權限。
HTTP 服務已有出勤持久化；此工具只測指定文件讀取，不能證明出勤交易或 Session 寫入可用。

## 查核來源與限制

- [kkdai/LineBotTemplate](https://github.com/kkdai/LineBotTemplate)
- [kkdai/line-card-assistant](https://github.com/kkdai/line-card-assistant/tree/main/internal/linebot)
- [kkdai/liff-template-go](https://github.com/kkdai/liff-template-go)
- [kkdai/linebot-garbage-helper](https://github.com/kkdai/linebot-garbage-helper)
- [LINE 官方來源](https://github.com/line/line-developers-docs-source/tree/dea09dc6e63fc5c6a3754abfbb304fc6fce4cc54)，
  2026-11-11 重新確認 default branch main 為此 revision，並讀取 LIFF 使用者資料頁核對後端 token 驗證。
  Rich Menu 與永久連結沿用先前查證，本次未重新執行外部發布。
- Context7 /line/line-bot-sdk-go/v8.23.0 與相同版本本機 SDK 核對 client、Context、Rich Menu 方法與型別。

真實 LINE 收發、手機 LIFF、Rich Menu 套用與 Firestore 連線仍須渠道與憑證驗收；
本機檢查不取代這些驗收。沒有新增版本庫測試方案或工具。
