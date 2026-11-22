# LINE Webhook 接線與驗證
各任務的入口與事件對照見 [LINE 任務接線表](wiring.md)。

1. 依[環境變數規格](configuration.md)注入 Messaging API Channel ID、secret 與 access token。Cloud Run 的機密綁定見[正式環境設定](../../operations/configuration.md)。
2. 在 LINE Developers 設定 Webhook URL 為 `<service-url>/webhooks/line/<channelID>` 並啟用 Webhook。需要群組功能時，允許官方帳號加入群組。
3. 分別驗證私聊和群組事件。HTTP 200、Cloud Run revision Ready 與 LINE 訊息送達各自確認。簽章、請求限制、回應碼和重送依 [HTTP 契約](../../contracts/http.md#基礎入口)。

後端以已驗證的事件來源核對身分；Postback 資料不能授予角色或資源權限。Postback 的格式與行為由各 handler 定義。私聊與群組的文字入口集中在 [LINE Bot 聊天指令](commands.md)。

群聊一般文字沒有 Bot 回覆時，仍可能按[摘要收集規則](../../domains/summary/summary.md)保存。出勤與摘要的保存、失敗和重送行為，依各自的領域契約。
