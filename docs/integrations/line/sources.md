# 查核來源與限制

## LINE 官方文件入口

| 入口                                                                  | 何時查閱                                                             |
| --------------------------------------------------------------------- | -------------------------------------------------------------------- |
| [LINE MINI App](https://developers.line.biz/en/docs/line-mini-app/)   | 查 MINI App 開發、審查與發布，以及永久連結、Service Message 等能力。 |
| [LINE Developers Basics](https://developers.line.biz/en/docs/basics/) | 查 Provider、Channel、權杖等跨產品基本概念。                         |
| [Messaging API](https://developers.line.biz/en/docs/messaging-api/)   | 查 Bot 接入、Webhook、訊息格式與發送、Rich Menu、群組聊天。          |
| [LINE Pay Developers](https://developers-pay.line.me/)                | 查 LINE Pay 付款整合與 API。導入時依服務地區核對適用規格。           |
| [LINE Developers News](https://developers.line.biz/en/news/1/)        | 查功能上線、規格異動、停用日期與遷移公告。                           |

常用子頁

- **MINI
  App**　[開發概覽](https://developers.line.biz/en/docs/line-mini-app/develop/develop-overview/) ·
  [永久連結](https://developers.line.biz/en/docs/line-mini-app/develop/permanent-links/) ·
  [服務通知](https://developers.line.biz/en/docs/line-mini-app/develop/service-messages/)
- **Basics**　[Channel Access Token](https://developers.line.biz/en/docs/basics/channel-access-token/)
- **Messaging
  API**　[接收訊息](https://developers.line.biz/en/docs/messaging-api/receiving-messages/) ·
  [發送訊息](https://developers.line.biz/en/docs/messaging-api/sending-messages/) ·
  [群組聊天](https://developers.line.biz/en/docs/messaging-api/group-chats/) ·
  [Rich Menu](https://developers.line.biz/en/docs/messaging-api/rich-menus-overview/) ·
  [API 參考](https://developers.line.biz/en/reference/messaging-api/)

各能力的 API 參考、限制與英文原文引文見[官方能力索引](capabilities.md#line-官方能力索引)。

## 查核紀錄

| 來源                                                                                                                     | 查核範圍                                     | 版本或時間                                                           |
| ------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------- | -------------------------------------------------------------------- |
| [LINE 官方文件原始碼](https://github.com/line/line-developers-docs-source/tree/dea09dc6e63fc5c6a3754abfbb304fc6fce4cc54) | 讀取 LIFF 使用者資料頁，核對後端 token 驗證  | 2026-11-11 確認 main 指向 `dea09dc6e63fc5c6a3754abfbb304fc6fce4cc54` |
| Context7 `/line/line-bot-sdk-go/v8.23.0` 與本機 SDK                                                                      | 核對 client、Context 與 Rich Menu 方法及型別 | v8.23.0                                                              |

社群實作參考（原查核未記錄固定版本）：

- [kkdai/LineBotTemplate](https://github.com/kkdai/LineBotTemplate)
- [kkdai/line-card-assistant](https://github.com/kkdai/line-card-assistant/tree/main/internal/linebot)
- [kkdai/liff-template-go](https://github.com/kkdai/liff-template-go)
- [kkdai/linebot-garbage-helper](https://github.com/kkdai/linebot-garbage-helper)

<a id="查核來源與限制-1"></a>

### 驗收限制

上述紀錄僅證明所列版本的文件與 SDK 查核，不證明外部發布成功。真實 LINE 收發、手機 LIFF、Rich Menu 套用與 Firestore 連線仍須使用目標渠道與憑證驗收；本機檢查只涵蓋本地行為。
