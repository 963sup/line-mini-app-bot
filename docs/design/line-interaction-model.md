---
status: implemented
owner: LINE integration
---

# LINE 互動介面盤點

本文件保存盤點、遷移矩陣與驗證證據。互動分配、詞彙、Action、私聊／群組規則與禁止事項只以 [LINE 互動標準](../standards/line-interactions.md) 為權威；後端所有權見 [LINE 後端責任與驗收](../architecture/line.md)。

## 盤點範圍

已逐一檢查 `bootstrap/line*`、`transport/line`、`platform/line`、所有 `business/*/adapter/line`、Webhook 指令與 postback 登錄、Messaging 發送、Rich Menu 與 `frontend/miniapp`。Room 沒有核准的業務規則，維持不支援；MINI App 的 DDD、授權、OpenAPI 與資料模型未變更。

## 十種介面現況

| 介面                    | 重構後用途                     | 現行案例                                           |
| ----------------------- | ------------------------------ | -------------------------------------------------- |
| Text                    | 通知、結果、拒絕、錯誤、無資料 | membership、summary、weather、出勤結果與錯誤       |
| Quick Reply             | 當前對話的暫時選項             | 功能選單、打卡操作                                 |
| Buttons Template        | 單一資源或表單入口             | MINI App、請假、補卡、漏下班、組織／企業、物件範本 |
| Confirm Template        | 二元決策                       | 尚無核准案例；不得為了使用介面而新增流程           |
| Carousel Template       | 多筆同類資源                   | 本人出勤群組、月份、我的工程                       |
| Image Carousel Template | 圖片是辨識主體的多筆資源       | 尚無核准案例                                       |
| Flex                    | 結構化資訊或統計               | 本人月度出勤、群組出勤看板                         |
| Rich Menu               | 固定入口                       | 出勤三頁：上班、轉場支援、下班                     |
| LINE MINI App           | 複雜表單、查詢、編輯           | 現有 `frontend/miniapp` 路由與表單                 |
| Imagemap                | 圖片區域操作                   | 尚無核准案例                                       |

`TemplateMessage` 只是 Buttons、Confirm、Carousel 與 Image Carousel 的外殼，不另算互動目的。

## Quick Reply 完整登錄

`/menu`、`選單`、`功能`、`/help`、群組加入事件與一對一未知文字均收斂到相同 Quick Reply。`appURL` 由 Bootstrap 注入；只有可解析且具有 host 的 HTTPS URL 可產生日誌與空間 URI。Quick Reply 的 MINI App postback 只在 `appURL` 非空時顯示，點擊後再由 Buttons Template 回覆正式入口。

### 一對一聊天室

| 順序 | 功能 ID               | 顯示名稱     | Action   | Payload／目的地         | 出現條件              |
| ---: | --------------------- | ------------ | -------- | ----------------------- | --------------------- |
|    1 | `attendance-history`  | 本人出勤紀錄 | Message  | `出勤紀錄`              | 一對一選單固定顯示    |
|    2 | `create-organization` | 建立組織     | Message  | `建立組織`              | 一對一選單固定顯示    |
|    3 | `create-enterprise`   | 建立企業     | Message  | `建立企業`              | 一對一選單固定顯示    |
|    4 | `repository-space`    | 空間         | URI      | `appURL?repositories=1` | `appURL` 是有效 HTTPS |
|    5 | `open-mini-app`       | MINI App     | Postback | `action=miniapp`        | `appURL` 非空         |

### 群組聊天室

| 順序 | 功能 ID                | 顯示名稱 | Action   | Payload／目的地                                                | 出現條件                                                                              |
| ---: | ---------------------- | -------- | -------- | -------------------------------------------------------------- | ------------------------------------------------------------------------------------- |
|    1 | `attendance-punch`     | 打卡     | Message  | `打卡`                                                         | 群組選單固定顯示                                                                      |
|    2 | `journal-compose`      | 日誌     | URI      | `appURL` + `journal=1&journalCompose=1&journalRepository=<ID>` | `appURL` 是有效 HTTPS；只有恰好一個有效 repository ID 時附 ID，否則仍可開啟未綁定日誌 |
|    3 | `attendance-board`     | 出勤     | Message  | `出勤`                                                         | 群組選單固定顯示                                                                      |
|    4 | `leave-registration`   | 請假登記 | Message  | `請假`                                                         | 群組選單固定顯示                                                                      |
|    5 | `correction-request`   | 補卡申請 | Message  | `補卡`                                                         | 群組選單固定顯示                                                                      |
|    6 | `conversation-summary` | 群聊摘要 | Message  | `摘要`                                                         | 群組選單固定顯示                                                                      |
|    7 | `repository-space`     | 空間     | URI      | `appURL?repositories=1&repository=<ID>`                        | `appURL` 是有效 HTTPS，且恰好一個 ID 符合 32 位小寫十六進位格式                       |
|    8 | `open-mini-app`        | MINI App | Postback | `action=miniapp`                                               | `appURL` 非空                                                                         |

### 打卡流程

群組輸入「打卡」後回覆 Text + Quick Reply，依序為「上班」「下班」「休息開始」「休息結束」「確認本班休息紀錄」。五項皆為 URI Action，目的地是後端建立的短效 launch token 對應 `appURL?launch=<token>`；缺少群組、使用者、`appURL` 或 launch store 時不產生此選單，改以 Text 說明錯誤。

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
| 請假、補卡、漏下班表單            | Buttons Template        | Buttons Template  | 原分配正確                   |
| 通知、摘要、天氣、邀請、錯誤      | Text                    | Text              | 原分配正確                   |
| 固定出勤入口                      | Rich Menu               | Rich Menu         | 原分配正確                   |

## 指令、Postback 與適用規則

| 領域         | 私聊           | 群組                           | 指令／Action                                  | 結果介面                                   |
| ------------ | -------------- | ------------------------------ | --------------------------------------------- | ------------------------------------------ |
| navigation   | 是             | 是                             | `/help`、`/menu`、`/miniapp`、別名、`miniapp` | Quick Reply／Buttons／Text error           |
| attendance   | 本人紀錄       | 打卡、看板、請假、補卡、漏下班 | `attendance-history` postback                 | Quick Reply／Buttons／Carousel／Flex／Text |
| organization | 建立組織／企業 | 否                             | 文字指令                                      | Buttons 或 Text                            |
| membership   | 邀請管理       | 依既有授權                     | 文字指令                                      | Text                                       |
| repository   | 我的工程       | 觀測與群組資源解析             | 文字指令                                      | Carousel 或 Text                           |
| object       | 物件範本       | 依現有 handler                 | 文字指令                                      | Buttons                                    |
| summary      | 否             | 群聊摘要                       | `摘要`                                        | Text                                       |
| weather      | 否             | 位置訊息                       | location event                                | Text                                       |

所有 postback 仍由 `ParsePostback` 驗證為單一 `action`；所有資源 ID 與授權在 Adapter／用例重新驗證。介面可見性不構成授權。

## 已消除的不一致

- 移除 Help 專用 Flex 與 `HelpCard`，避免 `/help`、`/menu` 與未知文字產生兩套功能標準。
- 顯示詞固定為「本人出勤紀錄」「請假登記」「補卡申請」「群聊摘要」；Message payload 維持既有短指令，未破壞路由。
- 單一 URI 不再塞入 Text 或單項 Quick Reply；多筆資源不再用裸 URL Text。
- 出勤群組與月份選擇不再由同一個通用 Flex helper 假裝成不同互動目的。
- 領域直接使用 LINE SDK 原生型別，沒有新增 Message、Template、Action 或路由鏡像層，也沒有新增依賴。

## 必要例外

| 例外                  | 原因                                     | 保護方式                               |
| --------------------- | ---------------------------------------- | -------------------------------------- |
| 本人月度出勤使用 Flex | 同時呈現月統計、逐日明細、狀態與分頁     | 標準登記；架構 allowlist；具體回覆測試 |
| 群組出勤看板使用 Flex | 同時呈現三種狀態人數、資料時間與計算說明 | 標準登記；架構 allowlist；具體回覆測試 |

Confirm、Image Carousel 與 Imagemap 沒有目前需求，因此維持零實作；這不是缺漏，也不得建立示範性流程。

## 驗證與量測

`backend/line_interaction_architecture_test.go` 禁止兩個已登記檔案以外的正式碼建立 Flex。各 Adapter 測試核對具體 SDK 型別、標籤、Action、payload、URI 與私聊／群組限制。最終 LOC、依賴變化及全量測試結果以本次變更的交付報告為準。
