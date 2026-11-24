---
status: implemented
owner: LINE integration
---

# LINE 互動介面盤點

本文件保存盤點、遷移矩陣與驗證證據。互動分配、詞彙、Action、私聊／群組規則與禁止事項只以 [LINE 互動標準](../standards/line_interactions.md) 為權威；後端所有權見 [LINE 後端責任與驗收](../architecture/line.md)。

## 盤點範圍

已逐一檢查 `bootstrap/line*`、`transport/line`、`platform/line`、所有 `business/*/adapter/line`、Webhook 指令與 postback 登錄、Messaging 發送、Rich Menu 與 `frontend/miniapp`。Room 沒有核准的業務規則，維持不支援；MINI App 的 DDD、授權、OpenAPI 與資料模型未變更。

## 十種介面現況

| 介面                    | 重構後用途                     | 現行案例                                           |
| ----------------------- | ------------------------------ | -------------------------------------------------- |
| Text                    | 通知、結果、拒絕、錯誤、無資料 | membership、summary、weather、出勤結果與錯誤       |
| Quick Reply             | 當前對話的暫時選項             | 功能選單、打卡操作                                 |
| Buttons Template        | 單一資源或表單入口             | MINI App、請假、補卡、漏下班、組織／企業、物件範本 |
| Confirm Template        | 二元決策                       | 尚無核准案例；不得為了使用介面而新增流程           |
| Carousel Template       | 多筆同類資源                   | 本人出勤群組、月份、我的工程                       |
| Image Carousel Template | 圖片是辨識主體的多筆資源       | 尚無核准案例                                       |
| Flex                    | 結構化資訊或統計               | 本人月度出勤、群組出勤看板                         |
| Rich Menu               | 固定入口                       | 出勤三頁：上班、轉場支援、下班                     |
| LINE MINI App           | 複雜表單、查詢、編輯           | 現有 `frontend/miniapp` 路由與表單                 |
| Imagemap                | 圖片區域操作                   | 尚無核准案例                                       |

`TemplateMessage` 只是 Buttons、Confirm、Carousel 與 Image Carousel 的外殼，不另算互動目的。

## 設計速查：功能與呈現效果

下表供設計時比較平台能力。專案採用情況見上方「十種介面現況」，實際選用依 [LINE 互動標準](../standards/line_interactions.md)。

### Messaging API 功能矩陣

| 功能                | 1：1 入口 | 群組入口 | 呈現效果與條件                       |
| ------------------- | --------- | -------- | ------------------------------------ |
| Text Message        | ✓         | ✓        | 文字泡泡                             |
| Text Message v2     | ✓         | ✓        | 文字中插入提及等替換內容             |
| Image Message       | ✓         | ✓        | 可展開的圖片                         |
| Video Message       | ✓         | ✓        | 影片預覽與播放                       |
| Audio Message       | ✓         | ✓        | 語音播放列                           |
| Sticker Message     | ✓         | ✓        | LINE 貼圖                            |
| Location Message    | ✓         | ✓        | 地圖位置卡                           |
| Imagemap Message    | ✓         | ✓        | 圖片上的可點擊區域                   |
| Template Message    | ✓         | ✓        | 固定格式卡片，可呈現按鈕、確認或輪播 |
| Flex Message        | ✓         | ✓        | 依資料排版的卡片或看板               |
| Quick Reply         | ✓         | ✓        | 手機輸入區上方的暫時操作列           |
| Rich Menu           | ✓         | —        | 私聊底部的圖文選單                   |
| Rich Menu Switch    | ✓         | —        | 點選後切換私聊的圖文選單頁           |
| Reply Message       | ✓         | ✓        | 回覆觸發事件，外觀取決於訊息內容     |
| Push Message        | ✓         | ✓        | 主動送達指定對象，群組須已加入 Bot   |
| Postback Event      | ✓         | ✓        | 點擊後由 Webhook 收到操作資料        |
| LIFF／MINI App Link | ✓         | ✓        | 點選連結開啟應用頁面                 |
| Coupon Message      | ✓         | ✓        | 優惠券訊息                           |

### Messaging API 實際呈現效果

| 元件              | 使用者看到什麼、怎麼操作                                                     |
| ----------------- | ---------------------------------------------------------------------------- |
| Buttons Template  | 固定格式功能選單：一張卡片顯示說明與操作按鈕，點選其中一項。                 |
| Confirm Template  | 雙按鈕確認操作：問題下方並列兩個選項，例如「確認」與「取消」。               |
| Carousel Template | 橫向滑動多張卡片，每張卡片可呈現一筆資料與對應按鈕。                         |
| Image Carousel    | 橫向滑動圖片選單，依圖片辨識項目，點選該圖執行操作。                         |
| Flex Message      | 自訂版面，可把標籤、數值、圖片與按鈕排在同一張卡片上。                       |
| Imagemap Message  | 圖片區域互動：一張圖內可設定多個點擊區，各自執行對應操作。                   |
| Quick Reply       | 暫時性快捷操作列：手機聊天室輸入區上方出現選項，供使用者回應當前訊息。       |
| Rich Menu         | 私聊底部可展開、收合的常駐圖文選單；設定頁籤與切換動作後可在多頁選單間切換。 |

### LINE 原生提供的 9 種 Action

| Action           | 元件範例或限制 | 點擊效果                                                               | Action 物件範例                                                                                               |
| ---------------- | -------------- | ---------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| Postback         | Buttons 按鈕   | 回傳 postback.data，供後端處理                                         | `{"type":"postback","label":"本人出勤紀錄","data":"action=attendance-history"}`                               |
| Message          | Quick Reply    | 以使用者身分送出文字，Bot 收到文字訊息                                 | `{"type":"message","label":"打卡","text":"打卡"}`                                                             |
| URI              | Buttons 按鈕   | 開啟 HTTPS 網頁或 MINI App，部署時代入正式網址                         | `{"type":"uri","label":"開啟表單","uri":"https://example.com/form"}`                                          |
| Datetime Picker  | Buttons 按鈕   | 選日期、時間或日期時間；結果在 postback.params，識別值在 postback.data | `{"type":"datetimepicker","label":"選日期","data":"field=date","mode":"date"}`                                |
| Camera           | 限 Quick Reply | 開啟相機，使用者拍攝並送出後收到圖片訊息                               | `{"type":"camera","label":"拍照"}`                                                                            |
| Camera Roll      | 限 Quick Reply | 開啟相簿，使用者選取並送出後收到媒體訊息                               | `{"type":"cameraRoll","label":"從相簿選擇"}`                                                                  |
| Location         | 限 Quick Reply | 開啟位置選擇器，使用者送出後收到位置訊息                               | `{"type":"location","label":"傳送位置"}`                                                                      |
| Rich Menu Switch | 限 Rich Menu   | 切換至已建立的 alias，並回傳 postback 事件                             | `{"type":"richmenuswitch","label":"下一頁","richMenuAliasId":"attendance-next","data":"tab=attendance-next"}` |
| Clipboard        | Buttons 按鈕   | 將指定文字複製到裝置剪貼簿                                             | `{"type":"clipboard","label":"複製代碼","clipboardText":"A123"}`                                              |

Go 實作將 Action 物件掛到元件，例如 `&messaging_api.MessageAction{Label: "打卡", Text: "打卡"}`。送出整則訊息使用 `(*Messaging).Reply(ctx, token, message)`；LINE 用戶端負責執行點擊動作。日期選擇器的 `mode` 可用 `date`、`time` 或 `datetime`。

Reply 與 Push 決定訊息如何送達；畫面取決於送出的訊息類型。從群組連結開啟 MINI App 時，後端仍須驗證使用者身分、成員資格與資源權限。MINI App 的 Service Message 用於符合平台條件的個人服務通知。

官方呈現範例與元件限制見 [訊息類型](https://developers.line.biz/en/docs/messaging-api/message-types/)、[Action 規格](https://developers.line.biz/en/reference/messaging-api/#action-objects) 與 [LINE MINI App](https://developers.line.biz/en/docs/line-mini-app/)。

### Messaging API 實作速查

在 Go 後端，沿用 `platform/line` 的 `(*Messaging).Reply(ctx, token, message)` 回覆單則訊息，或用 `(*Messaging).PushText(ctx, group, text, retryKey)` 推送群組文字。下表列出平台 JSON 欄位；選好內容後再決定如何發送。

| 訊息或元件             | 建構方式與主要輸入                                                                                                                                                                                           |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 文字、替換文字、優惠券 | `type: text` 搭配 `text`；`textV2` 用 `text` 與 `substitution` 插入提及等內容；`coupon` 指定 `couponId`                                                                                                      |
| 圖片、影片、音訊       | `image`、`video` 帶 HTTPS 的 `originalContentUrl`、`previewImageUrl`；`audio` 帶 `originalContentUrl`、毫秒數 `duration`                                                                                     |
| 貼圖、地點             | `sticker` 指定 `packageId`、`stickerId`；`location` 指定 `title`、`address`、`latitude`、`longitude`                                                                                                         |
| 固定格式卡片           | 外層 `type: template`、`altText`；內層 `template.type` 選 `buttons`、`confirm`、`carousel` 或 `image_carousel`。前兩種用 `actions`，後兩種用 `columns`                                                       |
| Flex、Imagemap         | `flex` 用 `altText` 與 `contents`（bubble 或 carousel）；`imagemap` 用 `baseUrl`、`baseSize`、`altText` 與含點擊區域的 `actions`                                                                             |
| Quick Reply            | 加在訊息的 `quickReply.items[]`，每項為 `{"type":"action","action":{"type":"message","label":"打卡","text":"打卡"}}`                                                                                         |
| Rich Menu              | 建立 `POST /v2/bot/richmenu`，上傳圖片至 `api-data.line.me/v2/bot/richmenu/{richMenuId}/content`（POST），設為預設用 `POST /v2/bot/user/all/richmenu/{richMenuId}`；專案發布由 `platform/line/richmenu` 負責 |
| 回覆、推送、接收事件   | `POST /v2/bot/message/reply` 傳 `replyToken` 與 `messages`；`POST /v2/bot/message/push` 傳 `to` 與 `messages`。Webhook 收到 `type: postback` 時由既有解析器讀取 `postback.data`                              |

### LINE MINI App 功能列表

MINI App 是 Web 應用。✓ 表示可從該聊天室入口開啟應用並設計此功能；各功能仍須具備對應的登入狀態、裝置權限或後端支援。

| 功能               | 1：1 入口 | 群組入口   | 呈現效果與條件                                   |
| ------------------ | --------- | ---------- | ------------------------------------------------ |
| LINE 身分識別      | ✓         | ✓          | 初始化 LIFF，取得並驗證登入身分                  |
| 表單填寫           | ✓         | ✓          | 輸入欄位、驗證後送出                             |
| 資料列表           | ✓         | ✓          | 逐筆呈現紀錄                                     |
| 詳細資料頁         | ✓         | ✓          | 點選一筆紀錄後查看完整內容                       |
| 搜尋與篩選         | ✓         | ✓          | 輸入關鍵字、條件後更新列表                       |
| Dashboard          | ✓         | ✓          | 彙整指標與狀態                                   |
| Calendar           | ✓         | ✓          | 按日期查看行程或出勤                             |
| Kanban 看板        | ✓         | ✓          | 按進度或狀態分欄                                 |
| Tabs 分頁          | ✓         | ✓          | 在同一頁切換內容區                               |
| Dialog 對話框      | ✓         | ✓          | 覆蓋目前畫面進行確認                             |
| Sheet 抽屜介面     | ✓         | ✓          | 展開面板顯示操作或篩選條件                       |
| 相機與檔案上傳     | ✓         | ✓          | 使用瀏覽器選檔或拍照，再送出檔案                 |
| Web Geolocation    | ✓         | ✓          | 使用者授權後取得裝置位置                         |
| 資料即時更新       | ✓         | ✓          | 依業務 API 重新查詢或訂閱更新                    |
| 分享至 LINE 聊天室 | ✓         | ✓          | 依 LIFF 啟動情境回傳訊息，或請使用者選擇分享對象 |
| Service Message    | 個人通知  | 非群組通知 | 已驗證 MINI App 依核准範本通知個別使用者         |

#### MINI App 函數與畫面

私聊與群組入口可以開啟同一頁面。以下是設計選項，各頁目前開放的功能與角色仍以現行路由及授權規則為準。

| 功能                         | 函數、元件或資料來源                                                                                                                                                                                                                                                    |
| ---------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 初始化、登入、識別           | `await liff.init({ liffId })` 後檢查 `liff.isLoggedIn()`；外部瀏覽器可呼叫 `liff.login()`。取得顯示資料用 `await liff.getProfile()`（profile scope），身分驗證用 `liff.getIDToken()`（openid scope）交給後端驗證                                                        |
| 表單、列表、搜尋、篩選、詳情 | 使用 Web 表單與頁面元件，透過專案 API client 讀寫資料；可用 Tabs 分類，Dialog 確認，Sheet／Drawer 放置篩選或編輯選項                                                                                                                                                    |
| Dashboard、Calendar、Kanban  | 前端依業務 API 回傳的統計、日期、狀態渲染；需要即時更新時，可重新查詢或訂閱後端提供的更新                                                                                                                                                                               |
| 相機、檔案、定位             | Web 使用 `<input type="file" accept="image/*" capture="environment">` 提示拍照，移除 capture 可選檔；`navigator.geolocation.getCurrentPosition(success, error)` 取得定位，須處理拒絕授權                                                                                |
| 分享到聊天室                 | `await liff.sendMessages(messages)` 傳回啟動應用的聊天，須符合啟動情境並取得 chat_message.write scope；讓使用者選擇分享對象可先檢查 `liff.isApiAvailable('shareTargetPicker')` 再呼叫 `await liff.shareTargetPicker(messages)`                                          |
| 個人服務通知                 | 已驗證 MINI App 使用核准範本。前端以 `liff.getAccessToken()` 取得使用者 token，後端以 channel access token 呼叫 `POST /message/v3/notifier/token`（liffAccessToken），再呼叫 `POST /message/v3/notifier/send?target=service`（notificationToken、templateName、params） |

### 業務流程設計示例

| 功能       | 私聊入口的設計目的     | 群組入口的設計目的               |
| ---------- | ---------------------- | -------------------------------- |
| 打卡、日誌 | 個人打卡、填寫日誌     | 出勤回報、分享工作成果           |
| 請假、紀錄 | 提交申請、查個人歷史   | 顯示必要的狀態通知、團隊統計     |
| 任務、審核 | 查看個人待辦與申請進度 | 指派任務、開啟具權限的主管審核頁 |
| 公告       | 個人通知               | 團隊廣播                         |

例如，群組出勤看板可呈現應到、已打卡與未打卡人數，再由按鈕開啟詳細資料。私聊、群組與 MINI App 共用業務 API，由後端驗證資源權限與狀態轉移，再透過既有持久化層存取 Firestore。以上示例供設計選擇，已上線項目請查本文件的現況與登錄表。

## Quick Reply 完整登錄

`/menu`、`選單`、`功能`、`/help`、群組加入事件與一對一未知文字均收斂到相同 Quick Reply。`appURL` 由 Bootstrap 注入；只有可解析且具有 host 的 HTTPS URL 可產生日誌與空間 URI。Quick Reply 的 MINI App postback 只在 `appURL` 非空時顯示，點擊後再由 Buttons Template 回覆正式入口。

### 一對一聊天室

| 順序 | 功能 ID               | 顯示名稱     | Action   | Payload／目的地         | 出現條件              |
| ---: | --------------------- | ------------ | -------- | ----------------------- | --------------------- |
|    1 | `attendance-history`  | 本人出勤紀錄 | Message  | `出勤紀錄`              | 一對一選單固定顯示    |
|    2 | `create-organization` | 建立組織     | Message  | `建立組織`              | 一對一選單固定顯示    |
|    3 | `create-enterprise`   | 建立企業     | Message  | `建立企業`              | 一對一選單固定顯示    |
|    4 | `repository-space`    | 空間         | URI      | `appURL?repositories=1` | `appURL` 是有效 HTTPS |
|    5 | `open-mini-app`       | MINI App     | Postback | `action=miniapp`        | `appURL` 非空         |

### 群組聊天室

| 順序 | 功能 ID                | 顯示名稱 | Action   | Payload／目的地                                                | 出現條件                                                                              |
| ---: | ---------------------- | -------- | -------- | -------------------------------------------------------------- | ------------------------------------------------------------------------------------- |
|    1 | `attendance-punch`     | 打卡     | Message  | `打卡`                                                         | 群組選單固定顯示                                                                      |
|    2 | `journal-compose`      | 日誌     | URI      | `appURL` + `journal=1&journalCompose=1&journalRepository=<ID>` | `appURL` 是有效 HTTPS；只有恰好一個有效 repository ID 時附 ID，否則仍可開啟未綁定日誌 |
|    3 | `attendance-board`     | 出勤     | Message  | `出勤`                                                         | 群組選單固定顯示                                                                      |
|    4 | `leave-registration`   | 請假登記 | Message  | `請假`                                                         | 群組選單固定顯示                                                                      |
|    5 | `correction-request`   | 補卡申請 | Message  | `補卡`                                                         | 群組選單固定顯示                                                                      |
|    6 | `conversation-summary` | 群聊摘要 | Message  | `摘要`                                                         | 群組選單固定顯示                                                                      |
|    7 | `repository-space`     | 空間     | URI      | `appURL?repositories=1&repository=<ID>`                        | `appURL` 是有效 HTTPS，且恰好一個 ID 符合 32 位小寫十六進位格式                       |
|    8 | `open-mini-app`        | MINI App | Postback | `action=miniapp`                                               | `appURL` 非空                                                                         |

### 打卡流程

群組輸入「打卡」後回覆 Text + Quick Reply，依序為「上班」「下班」「休息開始」「休息結束」「確認本班休息紀錄」。五項皆為 URI Action，目的地是後端建立的短效 launch token 對應 `appURL?launch=<token>`；缺少群組、使用者、`appURL` 或 launch store 時不產生此選單，改以 Text 說明錯誤。

## 重構前後分配矩陣

| 互動語意                          | 重構前                  | 重構後            | 統一理由                     |
| --------------------------------- | ----------------------- | ----------------- | ---------------------------- |
| `/help`、加入與未知文字的功能導航 | Flex                    | Quick Reply       | 暫時性功能引導               |
| MINI App 單一入口                 | Flex                    | Buttons Template  | 單一資源入口                 |
| 打卡五項選擇                      | Flex                    | Quick Reply       | 當前流程選項                 |
| 本人出勤群組／月份                | Flex                    | Carousel Template | 多筆同類資源                 |
| 本人月度出勤結果                  | Flex                    | Flex              | 統計、明細與分頁的結構化內容 |
| 群組出勤統計                      | Text                    | Flex              | 統計看板需要清楚資訊層級     |
| 物件範本入口                      | 含裸 URL 的 Text        | Buttons Template  | 單一資源入口                 |
| 組織／企業建立結果入口            | Text + 單項 Quick Reply | Buttons Template  | 單一新建資源入口             |
| 我的工程                          | 含多個裸 URL 的 Text    | Carousel Template | 多筆同類資源                 |
| 請假、補卡、漏下班表單            | Buttons Template        | Buttons Template  | 原分配正確                   |
| 通知、摘要、天氣、邀請、錯誤      | Text                    | Text              | 原分配正確                   |
| 固定出勤入口                      | Rich Menu               | Rich Menu         | 原分配正確                   |

## 指令、Postback 與適用規則

| 領域         | 私聊           | 群組                           | 指令／Action                                  | 結果介面                                   |
| ------------ | -------------- | ------------------------------ | --------------------------------------------- | ------------------------------------------ |
| navigation   | 是             | 是                             | `/help`、`/menu`、`/miniapp`、別名、`miniapp` | Quick Reply／Buttons／Text error           |
| attendance   | 本人紀錄       | 打卡、看板、請假、補卡、漏下班 | `attendance-history` postback                 | Quick Reply／Buttons／Carousel／Flex／Text |
| organization | 建立組織／企業 | 否                             | 文字指令                                      | Buttons 或 Text                            |
| membership   | 邀請管理       | 依既有授權                     | 文字指令                                      | Text                                       |
| repository   | 我的工程       | 觀測與群組資源解析             | 文字指令                                      | Carousel 或 Text                           |
| object       | 物件範本       | 依現有 handler                 | 文字指令                                      | Buttons                                    |
| summary      | 否             | 群聊摘要                       | `摘要`                                        | Text                                       |
| weather      | 否             | 位置訊息                       | location event                                | Text                                       |

所有 postback 仍由 `ParsePostback` 驗證為單一 `action`；所有資源 ID 與授權在 Adapter／用例重新驗證。介面可見性不構成授權。

## 已消除的不一致

- 移除 Help 專用 Flex 與 `HelpCard`，避免 `/help`、`/menu` 與未知文字產生兩套功能標準。
- 顯示詞固定為「本人出勤紀錄」「請假登記」「補卡申請」「群聊摘要」；Message payload 維持既有短指令，未破壞路由。
- 單一 URI 不再塞入 Text 或單項 Quick Reply；多筆資源不再用裸 URL Text。
- 出勤群組與月份選擇不再由同一個通用 Flex helper 假裝成不同互動目的。
- 領域直接使用 LINE SDK 原生型別，沒有新增 Message、Template、Action 或路由鏡像層，也沒有新增依賴。

## 必要例外

| 例外                  | 原因                                     | 保護方式                               |
| --------------------- | ---------------------------------------- | -------------------------------------- |
| 本人月度出勤使用 Flex | 同時呈現月統計、逐日明細、狀態與分頁     | 標準登記；架構 allowlist；具體回覆測試 |
| 群組出勤看板使用 Flex | 同時呈現三種狀態人數、資料時間與計算說明 | 標準登記；架構 allowlist；具體回覆測試 |

Confirm、Image Carousel 與 Imagemap 沒有目前需求，因此維持零實作；這不是缺漏，也不得建立示範性流程。

## 驗證與量測

`backend/line_interaction_architecture_test.go` 禁止兩個已登記檔案以外的正式碼建立 Flex。各 Adapter 測試核對具體 SDK 型別、標籤、Action、payload、URI 與私聊／群組限制。最終 LOC、依賴變化及全量測試結果以本次變更的交付報告為準。
