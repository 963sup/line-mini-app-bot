# 查核來源與限制

## LINE 官方文件入口

| 入口 | 何時查閱 |
| --- | --- |
| [LINE MINI App](https://developers.line.biz/en/docs/line-mini-app/) | 查 MINI App 開發、審查與發布，以及永久連結、Service Message 等能力。 |
| [LINE Developers Basics](https://developers.line.biz/en/docs/basics/) | 查 Provider、Channel、權杖等跨產品基本概念。 |
| [Messaging API](https://developers.line.biz/en/docs/messaging-api/) | 查 Bot 接入、Webhook、訊息格式與發送、Rich Menu、群組聊天。 |
| [LINE Pay Developers](https://developers-pay.line.me/) | 查 LINE Pay 付款整合與 API。導入時依服務地區核對適用規格。 |
| [LINE Developers News](https://developers.line.biz/en/news/1/) | 查功能上線、規格異動、停用日期與遷移公告。 |

常用子頁

- **MINI App**　[開發概覽](https://developers.line.biz/en/docs/line-mini-app/develop/develop-overview/) · [永久連結](https://developers.line.biz/en/docs/line-mini-app/develop/permanent-links/) · [服務通知](https://developers.line.biz/en/docs/line-mini-app/develop/service-messages/)
- **Basics**　[Channel Access Token](https://developers.line.biz/en/docs/basics/channel-access-token/)
- **Messaging API**　[接收訊息](https://developers.line.biz/en/docs/messaging-api/receiving-messages/) · [發送訊息](https://developers.line.biz/en/docs/messaging-api/sending-messages/) · [群組聊天](https://developers.line.biz/en/docs/messaging-api/group-chats/) · [Rich Menu](https://developers.line.biz/en/docs/messaging-api/rich-menus-overview/) · [API 參考](https://developers.line.biz/en/reference/messaging-api/)

各能力的 API 參考、限制與英文原文引文見[官方能力索引](capabilities.md#line-官方能力索引)。

## 查核紀錄

## 查核來源與限制

- -kkdai/LineBotTemplate（https://github.com/kkdai/LineBotTemplate）

- -kkdai/line-card-assistant（https://github.com/kkdai/line-card-assistant/tree/main/internal/linebot）

- -kkdai/liff-template-go（https://github.com/kkdai/liff-template-go）

- -kkdai/linebot-garbage-helper（https://github.com/kkdai/linebot-garbage-helper）

- - LINE<br>官方來源（https://github.com/line/line-developers-docs-source/tree/dea09dc6e63fc5c6a3754abfbb304fc6fce4cc54），

-   2026-11-11 重新確認 default branch main 為此revision，並讀取 LIFF 使用者資料頁核對後端 token 驗證。

-   Rich Menu 與永久連結沿用先前查證，本次未重新執行外部發布。

- - Context7 /line/line-bot-sdk-go/v8.23.0 與相同版本本機SDK 核對 client、Context、Rich Menu 方法與型別。

- 真實 LINE 收發、手機 LIFF、Rich Menu 套用與 Firestore連線仍須渠道與憑證驗收； 本機檢查不取代這些驗收。沒有新增版本庫測試方案或工具。
