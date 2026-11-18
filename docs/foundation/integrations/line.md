# LINE 整合：能力與邊界

> 目標設計，尚不代表功能已實作。新功能採當前官方支援方案；不建立舊功能相容層。
> 查核日期：2026-11-07。官方來源快照：`090500a12f9777995006d9fb5cafab8d83a792b2`。

環境變數統一使用 `LINE_MESSAGING_API_` 與 `LINE_MINI_APP_`，詳見 [環境變數規格](line-environment.md)。

## 1. 先選場景

| 場景 | 識別與輸入 | 後端責任 |
| --- | --- | --- |
| 一對一聊天 | `source.type=user`、`userId` | 個人操作、好友生命週期、Reply／Push |
| 群組聊天 | `source.type=group`、`groupId`；`userId` 視事件可用性 | 群組生命週期、成員事件、群組授權、Reply／Push |
| MINI App | LIFF SDK＋HTTPS API | 後端驗證 LINE token、資源授權、共用業務使用案例 |

群組須在 Console 開啟 Allow bot to join group chats。回群訊息會對群內所有成員顯示，
個人敏感資料不能直接回到群組。`groupId` 不能當成 `userId`；群組事件也不能假設總有發話者 ID。
一般成員事件不等於取得完整名單的權限，名單 API 仍須檢查帳號資格。

## 2. 訊息與互動能力

### 2.1 優先納入設計

| 能力 | 用途 | 邊界與驗證 |
| --- | --- | --- |
| Text／Text v2 | 一般文字、需要時使用 v2 表達能力 | 依需要的 mention／替換等功能選型；文字長度按官方規格驗證 |
| Flex Message | 卡片、bubble／carousel、自訂版面 | JSON、`altText`、API 驗證、Simulator 與裝置測試；不可使用 deprecated Filler |
| Quick Reply | 隨訊息附上快速操作 | 一對一與群組均可；不是持久導覽列 |
| Rich Menu | 一對一聊天底部導覽、開啟 MINI App | 預設／個人化、圖片、點擊區、alias／切換；不提供群組 Rich Menu，PC 不顯示 |
| Actions | Postback、Message、URI、日期時間、相機／相簿、位置、選單切換、Clipboard | 可用類型依承載元件與客戶端而異，不是每種 action 都能放在 Flex 或 Rich Menu |
| 媒體／位置／貼圖 | 圖片、影片、音訊、位置、貼圖的收發情境 | 收到的檔案依 content API 取得；格式、URL、保存與隱私分別處理 |

Rich Menu 管理與 Flex JSON 組裝放在 LINE adapter／呈現邊界，業務層只輸出業務結果。
Rich Menu 由 API 或 Official Account Manager 管理，先指定管理來源，避免混用。
Postback data 只是輸入，不能作為授權證據；執行操作前重新驗證身分、資源與目前狀態。

### 2.2 發送方式獨立選擇

| 方式 | 對象與用途 | 選用規則 |
| --- | --- | --- |
| Reply | 回應有 reply token 的事件 | token 短效且只能使用一次，不能延後當排程通知 |
| Push | 指定使用者或群組 | 檢查接收資格、訊息額度、限流與失敗語義 |
| Multicast | 指定多個 user ID | 按需求啟用，不用來發送多個群組 |
| Broadcast／Narrowcast | 好友廣播／受眾分眾 | 行銷需求確認後啟用，檢查方案、資格、額度與非同步進度 |
| MINI App Service Message | 回應 MINI App 的特定使用者操作 | 獨立 API、官方模板與 notification token，不是任意 Flex／Push |

HTTP 成功代表 API 接受請求，不代表使用者已收到或讀取。
僅在 endpoint 支援時使用 retry key；重試仍須配合業務冪等性。

## 3. Webhook 事件與可靠性

1. 對未改動的原始 body 驗證簽章，限制 body 大小；驗證後才解析與分流。
2. 處理一次請求中的多個事件，以及驗證請求可能出現的空事件陣列。
3. 依 `webhookEventId` 去重，不以 `isRedelivery` 作唯一判斷；事件可能重複或順序不同。
4. 依來源與事件類型分派：Message、Postback、Follow／Unfollow、Join／Leave、Member Joined／Left、Unsend。
5. Unsend 需依保存政策處理已儲存的內容，避免繼續顯示或使用已收回資料。
6. 需要非同步處理時先可靠保存工作再回成功；不在請求結束後依賴未受管理 goroutine。

去重狀態需區分處理中、完成與失敗，避免先標完成卻未執行；外部發送不在 Firestore transaction callback 內。
未知或非本產品範圍事件採安全忽略並觀測，不能觸發未授權業務。

## 4. MINI App：與 Bot 分開管理

| 項目 | 規則 |
| --- | --- |
| 前端 | 使用 LIFF SDK 初始化、登入／授權、需要時分享與永久連結；處理外部瀏覽器與功能可用性 |
| 後端身分 | 驗證 LINE ID token 的預期 channel／audience 與有效性；不能信任前端 profile／decoded token |
| Provider | Messaging API 與 MINI App 的 provider 配置先確認；跨 provider 不能直接用 user ID 當相同身分 |
| Channel | 區分 Messaging API 與 MINI App 憑證；MINI App 的 Developing／Review／Published 環境分開配置 |
| 好友關係 | 開啟 MINI App 不代表已加入官方帳號好友，也不代表具有任意 Push 資格 |
| Service Message | 正式使用限符合資格的 verified MINI App 與通過審查的模板；僅回應使用者操作，不作廣告 |
| 通知狀態 | 保存 API 回傳的最新 notification token、剩餘次數及有效性，安全管理並協調並行發送 |
| 發布 | 檢查當地支援、審查、同意畫面、隱私政策與設計規範；需要時再納入付款等功能 |

Service Message 顯示於區域對應的 MINI App 通知聊天室，不等同官方帳號一對一聊天。
Google `idtoken` 套件不能驗證 LINE ID token。LINE SDK 的版本由 `go.mod` 管理。

## 5. 範圍選擇：現行不等於全部實作

- 現行但按需求選用：Coupon、受眾／統計、Account Link、Beacon、付款、Quick-fill、分享功能等。
- 本產品預設用 Flex 處理豐富卡片；Template／Imagemap 仍見於官方現行文件，並非已淘汰，只是不列為本產品預設方案。
- 排除舊 `source.type=room` 多人聊天室的業務支援；官方說明新多人聊天早已整合為群組。收到範圍外來源安全忽略。
- 不使用官方標示 deprecated 的欄位、元件或 API，不建立舊版相容層。實作前再次核對使用 endpoint 的現行 reference 與 SDK。

## 6. 驗收清單

- 一對一／群組來源、缺少 user ID、群組退出與重入、個人資料不洩漏至群組。
- 簽章錯誤、多事件、空事件、重複／亂序、Unsend、下游失敗及重試。
- Flex JSON／altText、按鈕授權、Rich Menu 作用範圍、URI／Postback 不同路徑。
- LINE token 無效、錯誤 channel、跨使用者操作、跨 provider 身分關聯。
- Reply token 不可用、Push 額度／資格、Service Message 未獲資格或 token 更新失敗。

## 7. 官方來源

本次閱讀 Messaging API overview、group chats、receiving／sending messages、Actions、Quick Reply、
Flex、Rich Menu、MINI App getting started／Service Messages，以及 LIFF user data 文件。
來源庫明示不含官方網站所有內容；本頁是產品能力基線，不宣稱已逐頁讀完整個 LINE 平台。

- [來源快照](https://github.com/line/line-developers-docs-source/tree/090500a12f9777995006d9fb5cafab8d83a792b2)
- [Messaging API overview](https://developers.line.biz/en/docs/messaging-api/overview/)
- [群組](https://developers.line.biz/en/docs/messaging-api/group-chats/)、[Webhook](https://developers.line.biz/en/docs/messaging-api/receiving-messages/)
- [訊息發送](https://developers.line.biz/en/docs/messaging-api/sending-messages/)、[API reference](https://developers.line.biz/en/reference/messaging-api/)
- [Flex](https://developers.line.biz/en/docs/messaging-api/using-flex-messages/)、[Rich Menu](https://developers.line.biz/en/docs/messaging-api/rich-menus-overview/)
- [Actions](https://developers.line.biz/en/docs/messaging-api/actions/)、[Quick Reply](https://developers.line.biz/en/docs/messaging-api/using-quick-reply/)
- [MINI App 開發](https://developers.line.biz/en/docs/line-mini-app/develop/develop-overview/)、[Service Messages](https://developers.line.biz/en/docs/line-mini-app/develop/service-messages/)
- [LIFF 後端使用者資料](https://developers.line.biz/en/docs/liff/using-user-profile/)
