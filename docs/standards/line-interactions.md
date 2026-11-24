# LINE 互動標準

本標準是 LINE 一對一私聊、群組、Room 與 LINE MINI App 互動分配的唯一權威來源。業務領域擁有規則與資料；本標準擁有呈現方式、名稱、Action、路由與錯誤回覆的一致性。

## 互動分配

選擇介面時必須先判斷互動目的，再選擇 LINE SDK 原生型別。

| 互動目的                                           | 必須使用                | LINE SDK 型別                               | 不得用於                           |
| -------------------------------------------------- | ----------------------- | ------------------------------------------- | ---------------------------------- |
| 純通知、結果與錯誤                                 | Text                    | `TextMessage`                               | 選項清單、資源入口、統計看板       |
| 暫時性選項與對話引導                               | Quick Reply             | `QuickReply` + `TextMessage`                | 永久導航、長期資源清單             |
| 單一資源操作與表單入口                             | Buttons Template        | `TemplateMessage` + `ButtonsTemplate`       | 多筆同類資源、二元確認             |
| 二元決策、審核與確認                               | Confirm Template        | `TemplateMessage` + `ConfirmTemplate`       | 三個以上選項、一般導航             |
| 多筆同類資源選擇                                   | Carousel Template       | `TemplateMessage` + `CarouselTemplate`      | 純統計、異質功能集合               |
| 多筆圖片資源選擇                                   | Image Carousel Template | `TemplateMessage` + `ImageCarouselTemplate` | 無圖片或圖片不影響選擇的資源       |
| 結構化資訊、統計看板或原生 Template 無法呈現的內容 | Flex                    | `FlexMessage`                               | 原生 Template 已能完整滿足的互動   |
| 固定功能入口                                       | Rich Menu               | `RichMenuRequest`                           | 暫時流程、查詢結果、動態資源       |
| 複雜表單、查詢與編輯                               | LINE MINI App           | LIFF／Web UI                                | 可由一個原生訊息動作完成的簡單操作 |
| 圖片區域操作                                       | Imagemap                | `ImagemapMessage`                           | 一般按鈕、純裝飾圖片               |

`TemplateMessage` 是 Buttons、Confirm、Carousel 與 Image Carousel 的傳輸外殼，不是獨立的互動目的。

## 選擇順序

1. 沒有使用者選擇且內容是通知、結果或錯誤時，使用 Text。
2. 選項只服務當前對話步驟時，使用 Quick Reply。
3. 一個資源或表單入口使用 Buttons Template；只有兩個互斥決策時使用 Confirm Template。
4. 多筆同類資源使用 Carousel；圖片是辨識資源的主要資訊時使用 Image Carousel。
5. 原生 Template 無法表達必要的結構或統計時才使用 Flex，並在「Flex 例外」記錄原因。
6. 長期固定入口使用 Rich Menu；需要複雜輸入、查詢或編輯時開啟 MINI App。
7. 點擊區域必須由圖片位置決定時才使用 Imagemap。

## Action

| Action                 | 使用時機                                             | 必要欄位與限制                                           |
| ---------------------- | ---------------------------------------------------- | -------------------------------------------------------- |
| `MessageAction`        | 使用者需要在聊天室看到送出的文字，且文字是已登錄指令 | `Label` 使用顯示詞彙；`Text` 使用命令詞彙                |
| `PostbackAction`       | 結構化選擇、分頁、切換或不應出現在聊天室的操作       | `Data` 使用 URL query encoding；必須只有一個 `action`    |
| `UriAction`            | 開啟 HTTPS 或 LIFF 入口                              | Label 描述目的；URI 必須由可信設定與經驗證的資源 ID 建立 |
| `DatetimePickerAction` | 使用者選擇日期或時間且不需要完整表單                 | 必須指定 mode、範圍與 postback action                    |
| `RichMenuSwitchAction` | 切換已發布的 Rich Menu alias                         | alias、label 與回復策略必須固定並受測試保護              |

領域 Adapter 可以直接建立符合本標準的 LINE SDK 原生型別。不得建立第二套 Message、Template、Action 或路由抽象；跨領域共用只限無業務語意的傳輸驗證與發送能力。

## 一對一與群組

- 一對一私聊承載個人資料、個人紀錄、邀請、建立與管理入口。
- 群組承載群組範圍操作、群組統計與群組共同可見結果。
- 需要個人隱私的群組操作必須引導至一對一私聊或經授權的 MINI App。
- 需要群組資源的 URI 必須由已驗證的群組對應取得資源 ID；畫面是否顯示不是授權證據。
- Room 未有明確業務規則與測試前不得視同群組。
- 同一能力跨一對一與群組時必須使用相同功能名稱；情境差異以補充詞區分，例如「本人出勤紀錄」與「群組出勤看板」。

## 命名與標籤

- 顯示名稱描述使用者目的；命令文字可以較短，但同一入口不得出現多個顯示名稱。
- 操作標籤使用動詞加受詞，例如「填寫請假」、「開啟組織」、「查看工程」。
- 建立、查看、編輯、審核、確認、取消與返回必須使用一致動詞。
- `AltText` 必須描述訊息目的，不得只寫「選單」或「訊息」。
- 群組與一對一的同名操作必須具有相同業務語意；不同語意必須使用不同名稱。

目前核准名稱如下：

| 功能         | 顯示名稱      | 命令或 action                    |
| ------------ | ------------- | -------------------------------- |
| 個人出勤歷史 | 本人出勤紀錄  | `出勤紀錄`／`attendance-history` |
| 群組出勤統計 | 群組出勤看板  | `出勤`                           |
| 打卡流程     | 打卡          | `打卡`                           |
| 請假表單     | 請假登記      | `請假`                           |
| 補卡表單     | 補卡申請      | `補卡`                           |
| 群聊摘要     | 群聊摘要      | `摘要`                           |
| 組織建立     | 建立組織      | `建立組織`                       |
| 企業建立     | 建立企業      | `建立企業`                       |
| MINI App     | 開啟 MINI App | `miniapp`                        |

## 路由與資源識別

- 文字命令與 postback action 必須在 `transport.Handler` 宣告，啟動時不得重複。
- Postback 必須以 `url.Values` 編碼並由 `ParsePostback` 解析；每個領域只驗證自己宣告的參數。
- URI 必須由可信的 `appURL` 與已驗證的資源 ID 建立，不得拼接未驗證的使用者輸入。
- 同一能力在 Message、Template、Rich Menu 與 MINI App 中必須使用同一個 action 名稱與資源 ID 格式。
- 分頁、月份、群組與其他選擇狀態放在 postback data；授權仍由業務 Adapter 或用例重新驗證。

## 錯誤與結果

- 可直接向使用者說明的通知、成功、拒絕、無資料、設定缺失與可修正錯誤使用 Text。
- 重試中的 Webhook 或外部服務失敗必須保留既有 HTTP 重試語意；不得先回覆成功訊息。
- 錯誤文字必須指出結果、是否有寫入及下一步；不得暴露 token、內部錯誤或資源存在性。
- 同一錯誤語意在一對一、群組與不同入口使用相同核心文字，僅加入必要的情境指引。

## Flex 例外

Flex 只允許下列已證明無法由原生 Template 完整表達的案例。新增例外必須先更新本表，記錄具體限制與替代方案。

| 實作                                                              | 原因                                                 | 原生 Template 的限制                                       |
| ----------------------------------------------------------------- | ---------------------------------------------------- | ---------------------------------------------------------- |
| `business/attendance/adapter/line/attendance_history.go` 月度結果 | 同時呈現多項統計、逐日明細、狀態說明與月份／分頁操作 | Buttons 與 Carousel 無法在單一訊息內保持統計層級與可讀明細 |
| `business/attendance/adapter/line/board.go` 群組出勤看板          | 同時呈現上班、休息、已下班人數與資料時間、計算範圍   | Text 缺少結構層級；Template 無法清楚呈現統計看板           |

## 禁止事項

- 各業務領域不得自行建立不同的互動分配標準。
- 不得重複封裝 LINE SDK Message、Template 或 Action 型別。
- 原生 Template 可以完整滿足需求時不得使用 Flex。
- 不得以 Text 貼上 URL 取代 Buttons 或 Carousel 的資源入口。
- 不得以 Flex 模擬 Quick Reply、Buttons、Confirm、Carousel 或 Image Carousel。
- 不得建立只有轉送、重新命名或複製欄位的互動抽象層。
- 不得因介面方便而移動業務規則、授權或資料一致性責任。

## 驗證

- 架構測試必須拒絕未記錄的 Flex 使用及 LINE SDK 越界依賴。
- 每個 Message／Template 建構測試必須確認具體 SDK 型別、顯示名稱、Action、Payload、適用情境與 URI 資源範圍。
- Webhook 測試必須確認命令與 postback 唯一、無效參數不被處理、私聊與群組規則不互相越界。
- MINI App 實機驗收與 Webhook mock 證據必須分開記錄。
