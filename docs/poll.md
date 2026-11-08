# 群組投票

在 Bot 所在群組輸入 `投票 午餐｜便當｜麵食` 建立 2–4 選項的單選投票。
題目最多 100 字，選項最多 40 字且不可重複。點選「選項 1」等按鈕投票；
每位使用者一票，再次點選其他選項會改票。票數與投票結果顯示在群組。

使用訊息中顯示的 ID：

- `投票結果 <ID>`：查詢最新票數。
- `結束投票 <ID>`：僅建立者可執行；結束後不可改票。

沿用 LINE Messaging API 與 Firestore 設定，資料存於 channels/{channel}/polls。
投票者識別取自已驗簽 webhook。選票與事件結果由後端保存，並非匿名資料；
目前沒有自動刪除期限。無法識別使用者時拒絕投票，跨群組操作被拒絕。
Firestore 交易原子更新選票、票數與事件結果；LINE 重送不會重複計票。
LINE 回覆採盡力傳送，若回覆失敗，可用查詢指令查看已保存結果。

設計使用 LINE Go SDK v8.23.0 的 PostbackAction 與 FlexMessage，
依 Context7 的 SDK 文件及 LINE OpenAPI 規格核對。
本地檢查不代表 LINE 真實裝置與 Firestore 整合驗收已完成。
