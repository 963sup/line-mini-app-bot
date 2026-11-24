# LINE 後端責任與驗收

互動呈現、名稱、Action、私聊／群組適用規則與 Flex 例外，以 [LINE 互動標準](../standards/line_interactions.md) 為唯一權威來源；本文件只定義後端所有權與驗收邊界。

## 修改入口

| 需求                                 | 權威實作                                                               | 行為與驗收                                                                                                                                                                                                   |
| ------------------------------------ | ---------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 驗證 MINI App ID token               | [LINE Verifier](../../backend/internal/platform/line/verify.go)        | 核對 issuer、audience、期限與 subject。瀏覽器 Session、Cookie、CSRF 仍由 Identity 處理。                                                                                                                     |
| LINE ID 格式、群組名稱與成員資格     | [platform/line](../../backend/internal/platform/line/)                 | 純字串判定放在 `lineid`。群名為可失敗的顯示資料；資格查詢比對 LINE 回應的 user ID，404 表示非成員，供應端失敗保留不可用狀態。                                                                                |
| 回覆、推送、重送結果                 | [Messaging](../../backend/internal/platform/line/messaging.go)         | 每次呼叫複製 SDK client 再設定 context。推送沿用 retry key；409 視為已接受。無效 reply token 轉為具名錯誤，由 transport 判斷重送策略。                                                                       |
| Webhook 路由、驗簽、群組事件、分派   | [transport/line](../../backend/internal/transport/line/)               | 先限制原始 body 為 1 MiB 並驗簽，整批共用五秒 deadline。依序執行優先指令、觀測者、一般指令。只有對應處理器可要求重送。                                                                                       |
| 共用選單、導覽、貼圖、Rich Menu 定義 | [navigation](../../backend/internal/transport/line/navigation/)        | 保留順序、精確貼圖配對與行動／桌面雙入口。MINI App 永久網址由平台產生。                                                                                                                                      |
| Rich Menu 發布                       | [richmenu](../../backend/internal/platform/line/richmenu/)             | 狀態比對、圖片上傳、切換 alias、舊 alias 遷移與租約；獨立部署工作擁有發布期限。                                                                                                                              |
| 環境設定、HTTP client、依賴注入      | [bootstrap](../../backend/internal/bootstrap/)                         | LINE 登入、Messaging API、Rich Menu 發布使用各自的憑證與啟用條件。查核不得把部分設定當成完整啟用。                                                                                                           |
| 領域指令與用例轉換                   | [business](../../backend/internal/business/) 下各領域的 `adapter/line` | Attendance 處理打卡、看板與紀錄；Membership 處理邀請；Organization 處理建立對話；Object 處理物件入口；Repository 處理工程入口與觀測；Team 處理團隊觀測；Summary 處理收集、指令與交付；Weather 處理地點預報。 |

## 指令註冊與重送

有固定指令或 Postback 的業務 Adapter 在 `transport.Handler` 填入 `Owner`、`Commands` 與 `Postbacks`。別名也列入 `Commands`。建立 Webhook 時會跨優先與一般指令檢查宣告，指令以去除首尾空白並轉成小寫的鍵比對，Postback 的 action 採精確比對。帶參數指令與多步驟對話仍由領域 Adapter 判斷參數、來源及狀態。

優先指令認領事件後，該事件略過 Team、Repository 觀測及後續的一般指令。未被認領的事件先交給觀測者，再交給一般指令。任一事件的可重試錯誤都可能讓整批重送，因此每個寫入用例必須能重放。例如 Membership 接受邀請時，在同一筆 Firestore 交易更新成員與稽核紀錄，重送相同接受事件時沿用既有結果。

## 前後目錄與任務

| 優先級 | 原位置                                                                           | 現行位置                                                                     | 驗收重點                                                        |
| ------ | -------------------------------------------------------------------------------- | ---------------------------------------------------------------------------- | --------------------------------------------------------------- |
| P0     | Webhook、Summary delivery 各自執行 SDK 發送；bootstrap 各建訊息 client           | `platform/line/messaging.go` 與單一服務端 Messaging 實例                     | 失效 token 重送、供應端錯誤、取消、retry key 與簽章案例均通過。 |
| P1     | `business/summary/adapter/line/{messagecollection,summarization}` 及入口轉發     | 同一 `adapter/line` 的 `collection.go`、`commands.go`、`summary_delivery.go` | 一般訊息維持靜默；摘要入列、排程交付與停用狀態維持原語意。      |
| P1     | Group、Repository、Attendance 各自組裝 membership client；導航處理事件時讀取環境 | `bootstrap/line_config.go`、注入的成員查詢器與貼圖設定                       | 前兩者共用一秒期限；出勤維持五秒；跨主機重新導向被拒絕。        |
| P2     | 啟動程式中的路由字串、租約與多處設定讀取                                         | transport 擁有路由；Rich Menu 擁有租約；bootstrap 擁有設定快照               | 維持原端點、租約文件與歷史 alias；隔離平台與業務核心的依賴。    |

## 保留項目

- **SDK 的事件與訊息型別**留在 webhook、導航與業務 LINE adapter 的邊界。Flex 卡片、位置及 postback 已有具型別的 SDK 表示，重新定義整套鏡像 DTO 會讓每次 SDK 升級多一份映射。共同群組觀測才解碼成共用 `GroupEvent`；業務核心只接收自己的值與命令。
- **Team 與 Repository 的觀測者**各自讀寫其領域文件。Repository 另處理個人 follow，兩者使用不同的查找條件與 not-found 錯誤。共用的是已驗簽事件與群名查詢；合併寫入會改變重送時各領域的狀態邊界。
- **各消費者的小型 Membership port**表達各自的授權需求，Go 的隱式介面讓同一平台實作可直接注入。將 port 提升為全域授權服務會把「是 LINE 群組成員」誤作「能操作特定業務資源」。
- **能力專屬期限**來自原有行為。群名最多 400 ms，Webhook 整批五秒，摘要推送的 HTTP 上限十秒，Rich Menu 每次 HTTP 呼叫上限二十秒。發布工作額外持有八分鐘的總期限。這些用途共用標準 context 取消機制。
- **Rich Menu 舊 alias**可能已連結至個別使用者，保留遷移與擁有權檢查。群組、Session、摘要與 Rich Menu 的 Firestore 路徑、欄位及序列化維持原格式。
- **間接引用**包含 Enterprise、Membership、Organization、Project、Repository、Team 的 ID 驗證、查詢與 Firestore 程式。格式由 `lineid` 提供，判斷誰能操作資源由領域維護。共用 OpenAPI 產物及 HTTP 契約只描述欄位；文字生成提示中出現 LINE 字樣也不構成 LINE API 整合。

## 結構與量測

### 持續架構驗收

`backend/line_interaction_architecture_test.go` 將下列邊界固定為自動驗收：

- LINE 回覆、推送與 Rich Menu 發布操作只能位於 `platform/line`。
- Postback query 只能由 `transport/line/postback.go` 解析。
- `platform/line` 不得反向依賴 transport 或 business；`transport/line` 不得依賴 business。
- 各業務領域的 `adapter/line` 不得直接依賴另一個業務領域。

這些檢查只限制責任越界，不建立 Message、Action、Handler 或業務模型的第二套抽象。

### Channel Adapter 擴充驗收

從組裝端提供一個 `transport.Handler`，宣告其文字指令、別名與 Postback，並在 Adapter 判斷事件是否屬於該能力。現行宣告用於啟動衝突檢查，實際路由仍取決於 Handler 的匹配程式。[簽章 Webhook 擴充測試](../../backend/internal/bootstrap/line_extension_test.go) 用一個新增的測試檔完成註冊、送出簽章 HTTP 請求並核對三個回覆，且確認未知文字沒有回覆。這個案例修改了零個既有核心檔；實際部署新領域仍要將其依賴和 Handler 接入 Bootstrap。

以 `252da76` 對照本輪工作區，計算 LINE 平台、Transport、各領域 LINE Adapter 與 Bootstrap 中檔名以 line 開頭的正式 Go 檔，排除測試。依賴邊以來源套件目錄與直接 import 路徑的唯一組合計數，包含標準庫。

| 指標                      | 基準 | 重構後 |
| ------------------------- | ---: | -----: |
| 正式 Go 檔                |   34 |     38 |
| 實體行數（含空行與註解）  | 2415 |   2575 |
| 套件目錄                  |   14 |     14 |
| 直接 import 邊            |  152 |    159 |
| `url.ParseQuery` 呼叫位置 |    2 |      1 |

以下列出有移動或影響依賴方向的路徑，皆相對於 `backend/internal/`。

| 重構前                                                                                                         | 重構後                                                                                                                     |
| -------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| `bootstrap/{line,summary}.go` 建訊息 client，`transport/line/webhook.go` 發送回覆，摘要 delivery 呼叫 Push API | `bootstrap/server.go` 建立共用實例；`platform/line/{messaging,group_name}.go` 執行；Webhook 控制重送；Summary 轉譯業務錯誤 |
| `bootstrap/{group,repository,server}.go` 各建立 Membership client                                              | `bootstrap/line_config.go` 建立；`platform/line/membership.go` 執行；Group 與 Repository 共用實例                          |
| `bootstrap/miniapp.go` 組永久網址；`transport/line/navigation/events.go` 讀環境                                | `platform/line/miniapp.go` 組永久網址；`bootstrap/line.go` 注入貼圖設定；導航保留選單解讀                                  |
| `bootstrap/line.go` 綁定 Webhook 路徑；`bootstrap/line_navigation.go` 實作交易租約                             | `transport/line/webhook.go` 綁定；`platform/line/richmenu/lease.go` 取得與釋放租約；bootstrap 組裝工作                     |
| `business/summary/adapter/line/{messagecollection,summarization}/`                                             | `business/summary/adapter/line/{collection,commands,summary_delivery}.go`                                                  |
| `platform/line/membership.go` 轉發群組 ID 驗證                                                                 | `platform/line/lineid` 是格式唯一來源；群組事件解碼只依賴純格式套件                                                        |

基準為 `4550d50`。下表使用 `internal` 的 production Go 檔案，包含產生碼。LOC 含註解與空行，深度以 `internal` 下的套件目錄層數計算，import 邊數為各套件 `go list -json` 的直接 Imports 數量總和，含標準庫。Repomix 的檔案數另含測試檔。

| 指標                              | 重構前 | 重構後 |
| --------------------------------- | -----: | -----: |
| Go packages                       |    168 |    166 |
| Production Go 檔案                |    307 |    311 |
| Production LOC                    | 33,899 | 33,937 |
| 最大目錄深度                      |      6 |      6 |
| 直接 import 邊數                  |  1,289 |  1,278 |
| 直接引用 LINE SDK 的檔案／套件    | 27／15 | 24／13 |
| Messaging API client 建構位置     |      3 |      2 |
| Membership client 建構位置        |      3 |      1 |
| Messaging channel ID 環境讀取位置 |      9 |      1 |
| Repomix Go 檔案數                 |    425 |    430 |

新增的 production 行數包含平台錯誤分類、設定快照與路由入口。兩個 Messaging API client 建構位置分屬常駐服務與獨立的 Rich Menu 部署程序，後者另需 Blob API 及不同期限。SDK 不進入業務 Domain 或 Application；Go 架構測試檢查 SDK import 只能出現在平台、Webhook／導航、特定業務 LINE adapter 或組裝端。

在 `backend/internal` 實際變動 35 個 Go 路徑，含測試，移動以原、新路徑各計一次。除了上述套件，本次也同步更新根層架構測試、Summary capability 的檔案與函數索引、README、後端開發規範及指令參考。

## 驗收結果

以 Repomix 對重構前後的 `backend/internal/**/*.go` 分別快照，再以 gopls 查宣告、references 與 incoming calls。候選集合涵蓋 LINE 套件、領域 LINE adapter、SDK／平台引用及 LINE／LIFF 字樣，production 候選檔案由 57 個變成 61 個。兩次同口徑比較都僅在產生的 OpenAPI 檔選取群組、Space、Session 建立與 LIFF 相關符號。基準為 673 個符號、5,050 筆引用、849 筆上游呼叫；最終為 685、5,123、868。這些為查詢回傳筆數，包含宣告與重疊引用；套件耦合度採上表的直接 import 邊數。

- **基準**：在現有 Mini App 靜態匯出與本機 Firestore emulator 上執行完整 Go 測試及架構檢查；以獨立的基準原始碼快照確認 race 測試。基準未發現阻斷項目。
- **分批驗收**：依次檢查 gofmt、`go build ./...`、`go vet ./...`、受影響檔案的 gopls diagnostics、單元測試、架構測試，再啟用模擬器執行整合。租約競爭測試曾發現呼叫期限到達時 gRPC 回傳的錯誤形態不同，修正為傳回呼叫端 context 的錯誤後，原測試通過。
- **最終驗收**：完整 `go test -race -count=1 ./...` 使用獨立的 loopback Firestore emulator 及實際 Mini App 靜態匯出，含子測試共 492 項通過，零失敗、零跳過。覆蓋瀏覽器專案往返、Session 撤銷與到期、群組生命週期、Webhook 驗簽與重送、共用 client 的並行 context、Rich Menu 發布租約與歷史 alias。
- **安全與資料**：本次 API 索引只調整 Summary 實作位置、handler 名稱及規格連結。對外 OpenAPI operation、Session、Cookie、CSRF、Firestore 文件形狀與租約集合名稱維持原狀。供應端測試使用注入的 HTTP transport；正式 channel、真實裝置及線上發布仍需另外驗證。

從專案根目錄可使用下列指令重跑最終測試，前提是 Mini App 已完成靜態匯出，且本機已安裝 Playwright 與 Chromium。

```bash
MINIAPP_BROWSER_EXPORT="$PWD/frontend/miniapp/out" \
  scripts/test/firestore_emulator.sh -race -p 4 ./...
```
