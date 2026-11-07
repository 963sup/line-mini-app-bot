# 最小 LINE 接入試作

參考 kkdai/LineBotTemplate，使用官方 Go SDK v8.23.0；不含出勤寫入、MINI App、
Firestore、持久去重或背景工作，不作正式業務入口。

## 啟動與設定

透過執行環境注入 LINE_MESSAGING_API_CHANNEL_ID、LINE_MESSAGING_API_CHANNEL_SECRET、
LINE_MESSAGING_API_CHANNEL_ACCESS_TOKEN，實際值不得提交至 Git。
全部未提供時只啟動 HTTP 服務，不註冊 Webhook；部分設定缺失時啟動失敗，錯誤只列設定名稱。
執行 `bash scripts/tool-exec.sh mise run dev`，預設 PORT=8080；不自動讀 .env，沒有健康檢查路由。

## 接入流程

Cloud Run 資源與憑證就緒、取得部署授權後，透過 Secret Manager 綁定設定並部署。
LINE Console 的 Webhook URL 為 `<cloud-run-service-url>/webhooks/line/<channelID>`。
啟用 Webhook，確認 Official Account Manager 自動回覆不干擾，再使用一對一聊天發文字。
預期回覆「LINE 連線已接通，目前尚未提供出勤功能。」。
簽章有效的空事件回 200。群組、standby、非文字與 isRedelivery 重送事件忽略，不做 Push。

## 限制

body 上限 1 MiB；過大回 413、簽章錯誤回 401、其他無效輸入回 400。
錯誤路徑或未啟用路由回 404，錯誤 method 回 405。
整批回覆共用 5 秒期限，發送失敗只記固定警告、不自動重試，Webhook 仍回 200。
此試作的 200 不表示回覆送達或工作已持久保存；跳過 isRedelivery 不等於完整冪等性。
不得直接把此 handler 用作出勤寫入；後續平台的可靠接收契約尚未實作。
目前環境未提供 LINE 憑證，尚未驗證真實收訊、回覆或 Cloud Run 部署；未新增測試工程。

## 查核來源

- [範例](https://github.com/kkdai/LineBotTemplate/blob/master/main.go)
- [官方文件](https://github.com/line/line-developers-docs-source/blob/main/docs/en/docs/messaging-api/receiving-messages/index.html.md)，
  本次讀取 blob SHA：18b253b625e077fb43bb45dfc980d7bcbcf83ecf（不是整庫提交 ID）。
- Context7 `/line/line-bot-sdk-go/v8.23.0` 核對 ParseRequest、WithHTTPClient、ReplyMessage；
  WithContext 與事件欄位另核對本機相同版本 SDK 原始碼。
