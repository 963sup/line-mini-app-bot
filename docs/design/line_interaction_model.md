---
status: implemented
owner: LINE integration
---

# LINE 互動介面盤點

本文件保存盤點、遷移矩陣與驗證證據。互動分配、詞彙、Action、私聊／群組規則與禁止事項只以 [LINE 互動標準](../standards/line_interactions.md) 為權威；業務分析使用[用例設計概念](domain_rules.md#用例設計概念)，後端所有權見 [LINE 後端責任與驗收](../architecture/line.md)。

## 盤點範圍

已逐一檢查 `bootstrap/line*`、`transport/line`、`platform/line`、所有 `business/*/adapter/line`、Webhook 指令與 postback 登錄、Messaging 發送、Rich Menu 與 `frontend/miniapp`。Room 沒有核准的業務規則，維持不支援；MINI App 的 DDD、授權、OpenAPI 與資料模型未變更。

## 現行入口總覽

本表登錄 LINE 平台可直接觸發的 Chat、Group 與 App 入口，以及沒有畫面的 observer 與背景交付。MINI App 內部頁面跳轉、一般業務 API 與前端元件事件不重複算成 LINE 入口。

| Interaction | Trigger                                          | Context | Owner                                                                | 行為                                                | 回覆或輸出                        |
| ----------- | ------------------------------------------------ | ------- | -------------------------------------------------------------------- | --------------------------------------------------- | --------------------------------- |
| Transport   | `POST /webhooks/line/{channel}`                  | 全部    | `transport/line`                                                     | 限制 body、驗簽、設定批次期限並依序分派             | HTTP 狀態                         |
| Chat        | `FollowEvent`                                    | 私聊    | Repository                                                           | 觀測個人身分並建立或恢復個人 Repository             | 無                                |
| Chat        | 已登錄文字指令                                   | 私聊    | Navigation、Attendance、Organization、Membership、Repository、Object | 執行對應用例或產生入口                              | Text／Quick Reply／Template／Flex |
| Chat        | 未知文字                                         | 私聊    | Navigation                                                           | 顯示目前可用功能                                    | Quick Reply                       |
| Chat        | `help`、`miniapp`、`attendance-history` postback | 私聊    | Navigation、Attendance                                               | 顯示選單、MINI App 入口或我的出勤                   | Quick Reply／Template／Flex       |
| Group       | `JoinEvent`                                      | 群組    | Team、Repository、Navigation                                         | 建立群組觀測、解析資源並顯示功能                    | Quick Reply                       |
| Group       | `LeaveEvent`                                     | 群組    | Team、Repository                                                     | 將群組觀測標記為非啟用                              | 無                                |
| Group       | `MemberJoinedEvent`                              | 群組    | Team、Repository                                                     | 更新群組成員觀測                                    | 無                                |
| Group       | 一般文字訊息                                     | 群組    | Team、Repository、Summary                                            | 更新群組觀測並靜默收集可摘要訊息                    | 無                                |
| Group       | 已登錄文字指令                                   | 群組    | Navigation、Attendance、Summary、Object                              | 執行對應用例或產生入口                              | Text／Quick Reply／Template／Flex |
| Group       | 位置訊息                                         | 群組    | Weather                                                              | 查詢位置預報                                        | Text                              |
| Group       | 符合設定的貼圖                                   | 群組    | Navigation                                                           | 顯示目前可用功能                                    | Quick Reply                       |
| Group       | Postback                                         | 群組    | Team、Repository、Navigation                                         | 更新觀測；已登錄 action 另執行對應互動              | 無或對應 Message                  |
| Delivery    | 摘要背景工作完成                                 | 群組    | Summary、`platform/line.Messaging`                                   | 使用原 operation retry key 推送摘要                 | Push Text                         |
| Publication | Rich Menu 部署工作                               | 私聊    | Navigation、`platform/line/richmenu`                                 | 組合定義、圖片、alias，取得租約後同步並設定預設選單 | 三頁 Rich Menu                    |

文字指令的完整字面值與用途見 [LINE Bot 聊天指令](../references/line_commands.md)。Observer 只記錄已驗簽的平台事實，不宣稱 LINE 群組、成員資格或外部 ID 是業務授權。

## 十種介面現況

| 介面                    | 重構後用途                     | 現行案例                                                 |
| ----------------------- | ------------------------------ | -------------------------------------------------------- |
| Text                    | 通知、結果、拒絕、錯誤、無資料 | membership、summary、weather、出勤結果與錯誤             |
| Quick Reply             | 當前對話的暫時選項             | 功能選單、打卡操作                                       |
| Buttons Template        | 單一資源或表單入口             | MINI App、請假、補卡、更正、漏下班、組織／企業、物件範本 |
| Confirm Template        | 二元決策                       | 尚無核准案例；不得為了使用介面而新增流程                 |
| Carousel Template       | 多筆同類資源                   | 本人出勤群組、月份、我的工程                             |
| Image Carousel Template | 圖片是辨識主體的多筆資源       | 尚無核准案例                                             |
| Flex                    | 結構化資訊或統計               | 個人月度出勤、群組出勤統計                               |
| Rich Menu               | 固定入口                       | 出勤三頁：上班、支援、下班                               |
| LINE MINI App           | 複雜表單、查詢、編輯           | 現有 `frontend/miniapp` 路由與表單                       |
| Imagemap                | 圖片區域操作                   | 尚無核准案例                                             |

`TemplateMessage` 只是 Buttons、Confirm、Carousel 與 Image Carousel 的外殼，不另算互動目的。

## 設計速查：功能與呈現效果

下表供設計時比較平台能力。專案採用情況見上方「十種介面現況」，實際選用依 [LINE 互動標準](../standards/line_interactions.md)。

### Messaging API 功能矩陣

| 功能                | 1：1 入口         | 群組入口    | 呈現效果與條件                                                                                                                     |
| ------------------- | ----------------- | ----------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| Text Message        | ✓                 | ✓           | 文字泡泡                                                                                                                           |
| Text Message v2     | ✓（一般文字替換） | ✓（含提及） | mention 僅限群組／多人聊天的 Reply 或 Push；官方帳號與被提及者須在同一聊天室，每則最多 20 次提及。本專案支援的業務情境見盤點範圍。 |
| Image Message       | ✓                 | ✓           | 可展開的圖片                                                                                                                       |
| Video Message       | ✓                 | ✓           | 影片預覽與播放                                                                                                                     |
| Audio Message       | ✓                 | ✓           | 語音播放列                                                                                                                         |
| Sticker Message     | ✓                 | ✓           | LINE 貼圖                                                                                                                          |
| Location Message    | ✓                 | ✓           | 地圖位置卡                                                                                                                         |
| Imagemap Message    | ✓                 | ✓           | 圖片上的可點擊區域                                                                                                                 |
| Template Message    | ✓                 | ✓           | 固定格式卡片，可呈現按鈕、確認或輪播                                                                                               |
| Flex Message        | ✓                 | ✓           | 依資料排版的卡片或看板                                                                                                             |
| Quick Reply         | ✓                 | ✓           | 手機輸入區上方的暫時操作列                                                                                                         |
| Rich Menu           | ✓                 | —           | 私聊底部的圖文選單                                                                                                                 |
| Rich Menu Switch    | ✓                 | —           | 點選後切換私聊的圖文選單頁                                                                                                         |
| Reply Message       | ✓                 | ✓           | 回覆觸發事件，外觀取決於訊息內容                                                                                                   |
| Push Message        | ✓                 | ✓           | 主動送達指定對象，群組須已加入 Bot                                                                                                 |
| Postback Event      | ✓                 | ✓           | 點擊後由 Webhook 收到操作資料                                                                                                      |
| LIFF／MINI App Link | ✓                 | ✓           | 點選連結開啟應用頁面                                                                                                               |
| Coupon Message      | ✓                 | ✓           | 優惠券訊息                                                                                                                         |

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
| Postback         | Buttons 按鈕   | 回傳 postback.data，供後端處理                                         | `{"type":"postback","label":"我的出勤","data":"action=attendance-history"}`                                   |
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

#### 從畫面反查可用 Action

| 元件                              | 可配置的 Action                                                                   | 版面或執行限制                                                                               |
| --------------------------------- | --------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| Quick Reply                       | Postback、Message、URI、Datetime Picker、Camera、Camera Roll、Location、Clipboard | iOS／Android；每則最多 13 項。一次送出多則訊息時，只顯示最後一則的 Quick Reply。             |
| Buttons／Confirm Template         | Postback、Message、URI、Datetime Picker、Clipboard                                | Buttons 最多 4 個操作；Confirm 固定 2 個。                                                   |
| Carousel／Image Carousel Template | Postback、Message、URI、Datetime Picker、Clipboard                                | 最多 10 欄；Carousel 每欄最多 3 個操作，各欄操作數須一致；Image Carousel 每張圖片 1 個操作。 |
| Flex Message                      | Postback、Message、URI、Datetime Picker、Clipboard                                | 掛在支援 action 的元件，例如 button；可先用官方模擬器預覽版面。                              |
| Rich Menu                         | Postback、Message、URI、Datetime Picker、Clipboard、Rich Menu Switch              | 每個點擊區設定一個操作；以 alias 切頁。                                                      |
| Imagemap                          | 專用的 Message、URI、Clipboard 區域物件                                           | 每個操作要有 area（x、y、width、height）；開啟網址的欄位是 linkUri。                         |

Clipboard 需 iOS／Android 的 LINE 14.0.0 以上。設計有複製操作的卡片時，可同時顯示要複製的文字。[Action 規格](https://developers.line.biz/en/reference/messaging-api/#action-objects)、[Quick Reply](https://developers.line.biz/en/docs/messaging-api/using-quick-reply/)、[Template 規格](https://developers.line.biz/en/reference/messaging-api/#template-messages)、[Imagemap](https://developers.line.biz/en/reference/messaging-api/#imagemap-message)。

設計預覽可直接看 [Template 與 Imagemap 的官方畫面範例](https://developers.line.biz/en/docs/messaging-api/message-types/)、[Flex Message Simulator](https://developers.line.biz/flex-simulator/) 及 [Rich Menu 分頁示例](https://developers.line.biz/en/docs/messaging-api/switch-rich-menus/)。

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

#### Rich Menu 交付與發布

- 圖稿使用 JPEG 或 PNG，寬 800–2500 px、高至少 250 px、寬高比至少 1.45，檔案最大 1 MB。建立選單時，`size` 填圖稿尺寸，`areas[].bounds` 填各點擊區的像素座標與尺寸。每頁最多 20 個點擊區，聊天室底部的 `chatBarText` 最多 14 字元，`selected` 決定預設是否展開。
- 多頁切換需要先建立目標選單與 alias。使用 `POST /v2/bot/richmenu/alias`，傳入 `richMenuAliasId` 與 `richMenuId`，切換按鈕再引用該 alias。已上傳的圖片無法原地替換。換圖需建立新選單、上傳新圖，再更新預設設定或 alias。
- 顯示優先順序由高到低為 Messaging API 的個人選單、Messaging API 的預設選單、LINE Official Account Manager 的預設選單。驗收預設選單時，應確認測試帳號是否已有個人選單。

[選單物件](https://developers.line.biz/en/reference/messaging-api/#rich-menu-object)、[圖片規格](https://developers.line.biz/en/reference/messaging-api/#upload-rich-menu-image)、[建立 alias](https://developers.line.biz/en/reference/messaging-api/#create-rich-menu-alias)、[顯示優先順序](https://developers.line.biz/en/docs/messaging-api/rich-menus-overview/#rich-menu-display)。

#### 發送對象、時限與設計驗證

| 需求           | 平台條件與專案邊界                                                                                                                                                                                                                    |
| -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 回覆當前操作   | Webhook 的 replyToken 只能使用一次，應在收到事件後 1 分鐘內使用。重送事件可從收到重送時起算 1 分鐘，但原 token 已使用或事件已發生 20 分鐘時不可用。平台 Reply／Push 每次最多 5 個訊息物件；本專案的 Reply 每次接收 1 個物件。         |
| 主動通知個人   | 可 Push 給好友，或在非好友於私聊傳訊後的 7 天內推送給該人。使用者封鎖官方帳號或刪除帳號時，API 仍可能回傳 HTTP 200 而使用者未收到。避免以 HTTP 200 更新「已讀」或「已送達」狀態。                                                     |
| 主動通知群組   | 官方帳號須在目標群組，Push 的 to 填群組 ID。一次要通知多個個人帳號可用 multicast，按受眾篩選用 narrowcast，對所有好友發送用 broadcast。這三種是平台的個人受眾發送能力；本專案的 PushText 入口以單一群組為目標。                       |
| 送出前驗證卡片 | 後端使用 channel access token 作為 Bearer，將 `{"messages":[...]}` 傳至 `POST https://api.line.me/v2/bot/message/validate/reply` 或 `/v2/bot/message/validate/push`。驗證對象只有 messages；收件人、replyToken 與手機畫面需另外驗收。 |

官方參考：[Reply 與 token 時限](https://developers.line.biz/en/reference/messaging-api/#send-reply-message)、[Push 對象與送達條件](https://developers.line.biz/en/reference/messaging-api/#send-push-message)、[發送方式比較](https://developers.line.biz/en/docs/messaging-api/sending-messages/)、[Reply 訊息物件驗證](https://developers.line.biz/en/reference/messaging-api/#validate-message-objects-of-a-reply-message)。

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

#### 分享條件與專案入口

Service Message 以一次服務操作取得的 notificationToken 管理後續通知。每個 token 自核發起有效 1 年，最多可發 5 則。成功發送後若還能續發，後端須保存回應中更新的 notificationToken，並讀取 expiresIn 與 remainingCount；兩者為 0 代表本次已發送、無法再續發。[官方 token 與發送回應](https://developers.line.biz/en/reference/line-mini-app/)。

`liff.sendMessages()` 須由聊天室開啟 LIFF browser，並取得 `chat_message.write` 授權。從「最近使用的服務」重新載入後無法傳回原聊天。設計指定收件對象的流程時，可提供分享對象選擇器，由使用者選擇並確認。[sendMessages](https://developers.line.biz/en/reference/liff/#send-messages)、[shareTargetPicker](https://developers.line.biz/en/reference/liff/#share-target-picker)、[Service Message 端點](https://developers.line.biz/en/reference/line-mini-app/)。

| 已實作畫面       | 前端入口                                                                                                                                                                                                                                                             |
| ---------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 應用首頁         | `/` 與 [/miniapp](../../frontend/miniapp/src/pages/miniapp/index.tsx)                                                                                                                                                                                                |
| 動態與日誌       | [/miniapp/feed](../../frontend/miniapp/src/pages/miniapp/feed.tsx)                                                                                                                                                                                                   |
| 探索             | [/miniapp/explore](../../frontend/miniapp/src/pages/miniapp/explore.tsx)                                                                                                                                                                                             |
| 儀表板           | [/miniapp/dashboard](../../frontend/miniapp/src/pages/miniapp/dashboard.tsx)                                                                                                                                                                                         |
| 專案             | [/miniapp/projects](../../frontend/miniapp/src/pages/miniapp/projects.tsx)                                                                                                                                                                                           |
| Repository       | [/miniapp/repositories](../../frontend/miniapp/src/pages/miniapp/repositories.tsx)                                                                                                                                                                                   |
| 我的出勤         | [/miniapp/attendance](../../frontend/miniapp/src/pages/miniapp/attendance.tsx)                                                                                                                                                                                       |
| 定位打卡         | [/miniapp/punch](../../frontend/miniapp/src/pages/miniapp/punch.tsx)                                                                                                                                                                                                 |
| 定位診斷         | [/location](../../frontend/miniapp/src/pages/location.tsx)                                                                                                                                                                                                           |
| 請假、補卡、審核 | [/miniapp/leave](../../frontend/miniapp/src/pages/miniapp/leave.tsx)、[/miniapp/correction](../../frontend/miniapp/src/pages/miniapp/correction.tsx)、[/miniapp/review](../../frontend/miniapp/src/pages/miniapp/review.tsx)                                         |
| 物件辨識與範本   | [/miniapp/object-recognition](../../frontend/miniapp/src/pages/miniapp/object-recognition.tsx)、[/miniapp/object-templates](../../frontend/miniapp/src/pages/miniapp/object-templates.tsx)、[/miniapp/poppler](../../frontend/miniapp/src/pages/miniapp/poppler.tsx) |

Calendar、Kanban、Tabs、Dialog、Sheet 等是 Web 設計選項，畫面與資料操作由前端元件和業務 API 提供。設計資源見 [LINE MINI App 官方設計文件原始碼](https://github.com/line/line-developers-docs-source/tree/main/docs/en/docs/line-mini-app/design)。

### 業務流程設計示例

| 功能       | 私聊入口的設計目的                                     | 群組入口的設計目的                               |
| ---------- | ------------------------------------------------------ | ------------------------------------------------ |
| 打卡、日誌 | 個人打卡、填寫日誌                                     | 出勤回報、分享工作成果                           |
| 請假、紀錄 | 提交申請、查個人歷史                                   | 顯示必要的狀態通知、團隊統計                     |
| 任務、審核 | 查看個人待辦與申請進度                                 | 指派任務、開啟具權限的主管審核頁                 |
| 公告       | 個人通知                                               | 團隊廣播                                         |
| 選單       | Rich Menu 提供常駐入口，Quick Reply 提供當下可用的操作 | Quick Reply 提供選項，群組出勤 Flex 看板呈現統計 |

例如，群組「出勤」可呈現應到、已打卡與未打卡人數，再由按鈕開啟詳細資料。私聊、群組與 MINI App 共用業務 API，由後端驗證資源權限與狀態轉移，再透過既有持久化層存取 Firestore。以上示例供設計選擇，已上線項目請查本文件的現況與登錄表。

## Quick Reply 完整登錄

`/menu`、`選單`、`功能`、`/help`、群組加入事件與一對一未知文字均收斂到相同 Quick Reply。`appURL` 由 Bootstrap 注入；只有可解析且具有 host 的 HTTPS URL 可產生日誌與空間 URI。Quick Reply 的 MINI App postback 只在 `appURL` 非空時顯示，點擊後再由 Buttons Template 回覆正式入口。

### 一對一聊天室

| 順序 | 功能 ID               | 顯示名稱 | Action   | Payload／目的地         | 出現條件              |
| ---: | --------------------- | -------- | -------- | ----------------------- | --------------------- |
|    1 | `attendance-history`  | 我的出勤 | Message  | `出勤紀錄`              | 一對一選單固定顯示    |
|    2 | `create-organization` | 建立組織 | Message  | `建立組織`              | 一對一選單固定顯示    |
|    3 | `create-enterprise`   | 建立企業 | Message  | `建立企業`              | 一對一選單固定顯示    |
|    4 | `repository-space`    | 空間     | URI      | `appURL?repositories=1` | `appURL` 是有效 HTTPS |
|    5 | `open-mini-app`       | MINI App | Postback | `action=miniapp`        | `appURL` 非空         |

### 群組聊天室

| 順序 | 功能 ID                 | 顯示名稱 | Action   | Payload／目的地                                                | 出現條件                                                                              |
| ---: | ----------------------- | -------- | -------- | -------------------------------------------------------------- | ------------------------------------------------------------------------------------- |
|    1 | `attendance-punch`      | 打卡     | Message  | `打卡`                                                         | 群組選單固定顯示                                                                      |
|    2 | `journal-compose`       | 日誌     | URI      | `appURL` + `journal=1&journalCompose=1&journalRepository=<ID>` | `appURL` 是有效 HTTPS；只有恰好一個有效 repository ID 時附 ID，否則仍可開啟未綁定日誌 |
|    3 | `attendance-board`      | 出勤     | Message  | `出勤`                                                         | 群組選單固定顯示                                                                      |
|    4 | `leave-registration`    | 請假     | Message  | `請假`                                                         | 群組選單固定顯示                                                                      |
|    5 | `missing-punch-request` | 補卡     | Message  | `補卡`                                                         | 群組選單固定顯示                                                                      |
|    6 | `time-correction`       | 更正     | Message  | `更正`                                                         | 群組選單固定顯示                                                                      |
|    7 | `conversation-summary`  | 摘要     | Message  | `摘要`                                                         | 群組選單固定顯示                                                                      |
|    8 | `repository-space`      | 空間     | URI      | `appURL?repositories=1&repository=<ID>`                        | `appURL` 是有效 HTTPS，且恰好一個 ID 符合 32 位小寫十六進位格式                       |
|    9 | `open-mini-app`         | MINI App | Postback | `action=miniapp`                                               | `appURL` 非空                                                                         |

### 打卡流程

群組輸入「打卡」後回覆 Text + Quick Reply，依序為「上班」「下班」「休息開始」「休息結束」「確認休息」。五項皆為 URI Action，目的地是後端建立的短效 launch token 對應 `appURL?launch=<token>`；缺少群組、使用者、`appURL` 或 launch store 時不產生此選單，改以 Text 說明錯誤。

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
| 請假、補卡、更正與漏下班入口      | 混用單一補正入口        | Buttons Template  | 依使用者目的拆分入口與引導   |
| 通知、摘要、天氣、邀請、錯誤      | Text                    | Text              | 原分配正確                   |
| 固定出勤入口                      | Rich Menu               | Rich Menu         | 原分配正確                   |

## 指令、Postback 與適用規則

| 領域         | 私聊                  | 群組                                 | 指令／Action                                  | 結果介面                                   |
| ------------ | --------------------- | ------------------------------------ | --------------------------------------------- | ------------------------------------------ |
| navigation   | 是                    | 是                                   | `/help`、`/menu`、`/miniapp`、別名、`miniapp` | Quick Reply／Buttons／Text error           |
| attendance   | 本人紀錄              | 打卡、看板、請假、補卡、更正、漏下班 | `attendance-history` postback                 | Quick Reply／Buttons／Carousel／Flex／Text |
| organization | 建立組織／企業        | 否                                   | 文字指令                                      | Buttons 或 Text                            |
| membership   | 邀請管理              | 否                                   | 文字指令                                      | Text                                       |
| team         | 否                    | 群組生命週期與成員觀測               | Join／Leave／MemberJoined／Message／Postback  | 無                                         |
| repository   | Follow 觀測、我的工程 | 觀測與群組資源解析                   | Follow／群組事件／文字指令                    | 無／Carousel／Text                         |
| object       | 物件範本              | 依現有 handler                       | 文字指令                                      | Buttons                                    |
| summary      | 否                    | 靜默收集與摘要                       | 一般文字／`摘要`                              | 無／Reply Text／背景 Push Text             |
| weather      | 否                    | 位置訊息                             | location event                                | Text                                       |

業務 postback 的 `data` 使用本專案約定的 query 格式。`ParsePostback` 要求恰有一個非空的 `action` 值，各 Adapter／用例再驗證參數、資源 ID 與授權。Rich Menu 切頁則由 LINE 依 alias 執行，專案在回傳資料使用 `tab=<alias>`；導航處理器不將這個事件當作業務指令。介面可見性不構成授權。

## 已消除的不一致

- 移除 Help 專用 Flex 與 `HelpCard`，避免 `/help`、`/menu` 與未知文字產生兩套功能標準。
- 顯示詞固定為「我的出勤」「請假」「補卡」「更正」「摘要」；Message payload 使用各自登錄的短指令。
- 單一 URI 不再塞入 Text 或單項 Quick Reply；多筆資源不再用裸 URL Text。
- 出勤群組與月份選擇不再由同一個通用 Flex helper 假裝成不同互動目的。
- 領域直接使用 LINE SDK 原生型別，沒有新增 Message、Template、Action 或路由鏡像層，也沒有新增依賴。

## 必要例外

| 例外                  | 原因                                     | 保護方式                               |
| --------------------- | ---------------------------------------- | -------------------------------------- |
| 本人月度出勤使用 Flex | 同時呈現月統計、逐日明細、狀態與分頁     | 標準登記；架構 allowlist；具體回覆測試 |
| 群組出勤統計使用 Flex | 同時呈現三種狀態人數、資料時間與計算說明 | 標準登記；架構 allowlist；具體回覆測試 |

Confirm、Image Carousel 與 Imagemap 沒有目前需求，因此維持零實作；這不是缺漏，也不得建立示範性流程。

## Use Case × Interaction Context

本矩陣以使用者目的盤點現行責任，不以 LINE 元件或業務領域分組。欄位描述各介面在同一用例中的呈現責任；業務規則、授權與狀態轉移仍由所屬領域負責，不因入口不同而複製。

| 標記 | 定義                                       |
| ---- | ------------------------------------------ |
| 入口 | 使用者可從該介面發起用例                   |
| 導流 | 該介面只提供前往完成介面的連結或操作引導   |
| 完成 | 使用者可在該介面完成主要操作或取得主要結果 |
| 結果 | 該介面只接收完成結果、通知或摘要           |
| 禁止 | 因隱私、資源歸屬或互動語意不得在該介面提供 |
| 無   | 現行沒有核准入口；不代表缺漏或待辦         |

同一格可以有兩個標記，例如「入口、完成」。平台事件、observer、背景交付與 Rich Menu 發布不是使用者用例，已由「現行入口總覽」登錄，不重複列入。

| 使用者目的       | 業務能力     | LINE 私聊        | LINE 群組        | MINI App   | 現行呈現或入口                                                      |
| ---------------- | ------------ | ---------------- | ---------------- | ---------- | ------------------------------------------------------------------- |
| 查看可用功能     | Navigation   | 入口、完成       | 入口、完成       | 無         | 私聊／群組 Quick Reply；私聊另有 Rich Menu                          |
| 開啟 MINI App    | Navigation   | 入口、導流       | 入口、導流       | 完成       | Quick Reply postback 或 `/miniapp` → Buttons Template               |
| 建立組織         | Organization | 入口、完成、導流 | 禁止             | 結果       | 私聊對話收集名稱並建立；Buttons Template 開啟建立後資源             |
| 建立企業         | Enterprise   | 入口、完成、導流 | 禁止             | 結果       | 私聊對話收集名稱並檢查開通資格；Buttons Template 開啟建立後資源     |
| 產生組織邀請     | Membership   | 入口、完成       | 禁止             | 無         | 私聊「邀請加入 組織ID」→ Text 邀請碼                                |
| 接受組織邀請     | Membership   | 入口、完成       | 禁止             | 無         | 私聊「接受邀請 邀請碼」→ Text 結果                                  |
| 撤銷組織邀請     | Membership   | 入口、完成       | 禁止             | 無         | 私聊「撤銷邀請 邀請碼」→ Text 結果                                  |
| 查看我的出勤     | Attendance   | 入口、完成       | 禁止             | 入口、完成 | 私聊 Carousel／Flex；MINI App 出勤頁                                |
| 開始上班         | Attendance   | 入口、導流       | 入口、導流       | 完成       | 私聊 Rich Menu 提示；群組「打卡」→「上班」；MINI App 驗證定位後寫入 |
| 結束下班         | Attendance   | 入口、導流       | 入口、導流       | 完成       | 私聊 Rich Menu 提示；群組「打卡」→「下班」；MINI App 驗證定位後寫入 |
| 開始休息         | Attendance   | 導流             | 入口、導流       | 完成       | 群組「打卡」→「休息開始」；MINI App 完成狀態轉移                    |
| 結束休息         | Attendance   | 導流             | 入口、導流       | 完成       | 群組「打卡」→「休息結束」；MINI App 完成狀態轉移                    |
| 確認本班休息紀錄 | Attendance   | 導流             | 入口、導流       | 完成       | 群組「打卡」→「確認休息」；MINI App 完成確認                        |
| 取得支援作業指引 | Attendance   | 入口、完成       | 無               | 無         | 私聊 Rich Menu 回覆指引，不結束、不新增也不轉移班次                 |
| 查看群組出勤統計 | Attendance   | 無               | 入口、完成       | 入口、完成 | 群組 Flex；MINI App 出勤頁依授權查詢                                |
| 登記請假         | Attendance   | 導流             | 入口、導流       | 完成       | 私聊提示回原群組；群組 Buttons Template → 請假表單；不需要定位      |
| 補卡             | Attendance   | 導流             | 入口、導流       | 完成       | 私聊提示回原群組；群組 Buttons Template → 補卡表單；一律人工審核    |
| 申請更正         | Attendance   | 導流             | 入口、導流       | 完成       | 私聊提示回原群組；群組 Buttons Template → 出勤紀錄 → 既有班次更正   |
| 查詢漏下班       | Attendance   | 無               | 入口、完成、導流 | 完成       | 群組查詢結果；需要處理時開啟補正表單                                |
| 審核出勤申請     | Attendance   | 無               | 禁止             | 入口、完成 | MINI App 審核頁；不得在群組公開個人申請內容                         |
| 查看可存取工程   | Repository   | 入口、完成、導流 | 無               | 入口、完成 | 私聊 Carousel；MINI App Repository 頁                               |
| 開啟目前群組空間 | Repository   | 無               | 入口、導流       | 完成       | 群組 Quick Reply URI；只有唯一有效資源 ID 時預選                    |
| 建立工作日誌     | Journal      | 無               | 入口、導流       | 入口、完成 | 群組 Quick Reply URI；MINI App 日誌表單                             |
| 查看工作日誌     | Journal      | 無               | 無               | 入口、完成 | MINI App 動態與日誌列表                                             |
| 管理物件範本     | Object       | 入口、導流       | 入口、導流       | 入口、完成 | Buttons Template → 物件範本頁                                       |
| 辨識物件         | Object       | 無               | 無               | 入口、完成 | 物件辨識頁；聊天未核准複雜檔案流程                                  |
| 擷取文件內容     | Object       | 無               | 無               | 入口、完成 | Poppler 頁；聊天未核准複雜檔案流程                                  |
| 搜尋跨領域資源   | Explore      | 無               | 無               | 入口、完成 | MINI App 探索頁                                                     |
| 管理專案         | Project      | 無               | 無               | 入口、完成 | MINI App 專案頁；聊天沒有核准操作                                   |
| 執行專案驗收     | Project      | 無               | 無               | 入口、完成 | MINI App 專案驗收流程；聊天沒有核准操作                             |
| 取得摘要         | Summary      | 禁止             | 入口、完成、結果 | 無         | 群組 Reply Text 回應受理；背景 Push Text 交付摘要                   |
| 查詢位置天氣     | Weather      | 無               | 入口、完成       | 無         | 群組位置訊息 → Text                                                 |

### 矩陣判讀

- Chat、Group 與 App 是同一業務用例的互動介面，不是三套業務能力；新增介面入口不得複製規則或狀態轉移。
- 每個入口只解析互動情境並映射 Actor、Resource 與 Action 或 Query；Authorization、State、結果及 Event 仍由既有業務用例負責，不建立通用 Interaction Domain 或跨領域處理器。
- Messaging API 與 MINI App 的 Command、Query、結果訊息及純導航都必須依 [LINE 互動標準](../standards/line_interactions.md) 保留可追溯語意；沒有業務寫入時不虛構狀態轉移或事件。
- 上班、下班、休息開始、休息結束與休息確認是不同操作及狀態轉移；「定位」只是打卡流程的驗證條件，不是涵蓋請假、支援或補卡的業務用例。
- 「支援」只提供作業指引；不得暗示系統會自動轉移、結束或新增班次。既有選單的舊 alias 與 `action=transfer` postback 僅保留讀取相容，不得成為顯示詞或新發布 action。
- 請假、補卡與更正由聊天導流、MINI App 完成表單；請假不需要定位，補卡與更正雖共用補正審核核心，仍是兩個入口與使用者目的，且不得共用打卡寫入語意。
- 日誌與物件操作需要表單或檔案能力時，由聊天導流、MINI App 完成；不得在聊天重建複雜表單。
- 個人出勤、邀請、審核與摘要已有明確的私密或群組邊界；介面對稱不是目標。不同命令造成不同寫入或狀態轉移時，即使同屬一個領域也必須分列。
- Explore、Project 與物件辨識目前是 App-only；沒有已核准的聊天入口，因此不列為缺漏。
- 下一輪只需逐列核對「同一用例是否共用同一應用服務」及「入口是否符合本矩陣」；無須先建立新的 LINE 核心框架。

## 完整性與下一步

現行入口以「現行入口總覽」、Quick Reply 完整登錄、Rich Menu、MINI App 畫面與指令矩陣共同構成盤點基線。新增或移除 LINE 入口時，必須同步登錄 Trigger、Context、Owner、行為及輸出；只改程式或只補視覺介面都不算完成。

下一步只來自兩類差距：已核准用例尚未具有合適入口，或現行入口與 [LINE 互動標準](../standards/line_interactions.md) 的分配不一致。Confirm、Image Carousel、Imagemap、Room、Service Message、multicast、narrowcast 與 broadcast 目前都沒有核准用例，不因平台提供能力而列入實作待辦。

## 驗證與量測

`backend/line_interaction_architecture_test.go` 禁止兩個已登記檔案以外的正式碼建立 Flex。各 Adapter 測試核對具體 SDK 型別、標籤、Action、payload、URI 與私聊／群組限制。最終 LOC、依賴變化及全量測試結果以本次變更的交付報告為準。
