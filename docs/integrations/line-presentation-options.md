# LINE 呈現方式選型

核對日期：2026-11-09。依 [LINE Developers 文件來源](https://github.com/line/line-developers-docs-source) `main` 修訂 `403d8adfcd58e495292a1e840c1f707004051f67`。此頁供設計選型，不擴大目前[基礎出勤範圍](../features/attendance.md)。實際可用性仍須依 LINE 客戶端、帳號設定與 API 參考驗證。

## Messaging API 訊息類型

| 類型 | 適用場景 | 限制與維護 | 本專案 |
| --- | --- | --- | --- |
| 文字、文字 v2 | 狀態、錯誤、簡短指令；v2 可替換提及與 emoji | 文字無複雜版面；新文字功能可能只加入 v2 | 一般文字已用；v2 待需求 |
| 貼圖、圖片、影片、音訊 | 情緒回應或展示媒體 | 媒體需託管、維護與可及性文案 | 未實作 |
| 位置 | 使用者傳送座標；現有群組位置天氣 | 位置是使用者輸入，不能當可信到場證明 | 已接收群組位置；不作出勤認證 |
| 優惠券 | 發送優惠券 | 需符合 LINE 優惠券資源與帳號條件 | 延後 |
| Imagemap | 單張圖上多個點擊區域 | 圖片尺寸、點擊熱區及替代文字需同步維護 | 未實作 |
| Template：Buttons、Confirm、Carousel、Image Carousel | 固定結構的決策、確認、橫向列表或圖片列表 | 版面和 action 數量受 API 限制；內容較 Flex 不靈活 | 未實作 |
| Flex | 自訂卡片與多欄內容 | 須維護 JSON 版面、替代文字和窄螢幕呈現 | 功能說明、打卡選單等已用 |

以上是訊息物件類型，詳見[官方訊息類型](https://developers.line.biz/en/docs/messaging-api/message-types/)與[Flex 指南](https://developers.line.biz/en/docs/messaging-api/using-flex-messages/)。文字 v2 是獨立格式，不應假定現有文字訊息會自動升級。

## 訊息上的互動與周邊介面

| 能力 | 適用場景 | 限制與維護 | 本專案 |
| --- | --- | --- | --- |
| Quick Reply | 訊息下方一個功能一顆小按鈕 | 每訊息最多 13 個；按下後通常消失；僅能用支援的 action | 「選單」已用訊息及 URI action |
| Rich Menu、選單切換 | 聊天底部常駐入口、分頁 | 上傳圖片、熱區與 alias 綁定需同步；切換 action 僅適用 Rich Menu | Rich Menu 與雙分頁設定已備；部署狀態另驗 |
| Action | 訊息、回傳資料、URI、日期時間等點擊行為 | 各容器支援種類不同；不可把 Rich Menu 切換放進 Quick Reply | 依現有命令使用；新增前核對容器 |
| 載入提示、引用 | 等待回覆及針對特定訊息回覆 | 是輔助能力，不是訊息版型；依 API 與聊天情境使用 | 未實作 |
| LIFF／LINE MINI App、一般網頁 | 複雜表單、本人出勤與更正 | 需處理身分、授權、HTTPS、裝置測試；一般網頁無自動 LINE 身分 | MINI App 已實作；一般定位示範頁保留相容 |

參考：[Quick Reply](https://developers.line.biz/en/docs/messaging-api/using-quick-reply/)、[Rich Menu](https://developers.line.biz/en/docs/messaging-api/using-rich-menus/)、[Actions](https://developers.line.biz/en/docs/messaging-api/actions/)、[載入提示](https://developers.line.biz/en/docs/messaging-api/use-loading-indicator/)、[引用 token](https://developers.line.biz/en/docs/messaging-api/get-quote-tokens/)、[LIFF](https://developers.line.biz/en/docs/liff/overview/)、[LINE MINI App](https://developers.line.biz/en/docs/line-mini-app/quickstart/)。不要把 LINE Official Account Manager 手動發布能力當作 Messaging API 已實作功能。
