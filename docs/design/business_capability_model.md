---
status: draft
owner: 各業務領域維護者
---

# 業務能力模型

## 1. 產品定位與業務目標

本產品讓成員在 LINE 對話與 MINI App 中協作，從企業與組織管理、專案執行到現場出勤與日誌紀錄。產品目標是讓成員在所屬工作範圍完成任務，讓具管理權限的人員檢視成果、處理申請與驗收。

### 我想做什麼？從哪裡開始？

在 Bot 私聊或群組聊天室輸入「選單」，再依下表選擇。粗體是現行選單上的文字。

| 我想做的事     | 在哪裡點選                          | 點選後會看到什麼                                                   |
| -------------- | ----------------------------------- | ------------------------------------------------------------------ |
| 查我自己的出勤 | 私聊 → **本人出勤紀錄**             | 選擇群組、月份，查看自己的月統計與逐日明細                         |
| 記錄上班或下班 | 群組 → **打卡**                     | 顯示打卡選項，接著選擇本次要記錄的動作                             |
| 看群組出勤概況 | 群組 → **出勤**                     | 在聊天室查看群組統計卡片，內含人數、資料時間與計算說明             |
| 漏打卡或要請假 | 群組 → **補卡申請** 或 **請假登記** | 收到表單入口，開啟 MINI App 填寫申請                               |
| 寫一筆工作日誌 | 群組 → **日誌**                     | 開啟日誌填寫頁；群組有唯一有效的工作空間識別碼時，入口會帶入該空間 |
| 建立組織或企業 | 私聊 → **建立組織** 或 **建立企業** | 收到對應的建立入口                                                 |

「日誌」需要有效的 HTTPS MINI App 網址才會出現在群組選單。其他入口的完整顯示條件見 [LINE 互動盤點](line_interaction_model.md)。

## 2. 業務領域（Business Domain）

每列以業務結果識別能力。契約鍵供查閱第 10 節的能力登錄檔。

| 業務領域     | 業務能力                           | 代表契約鍵                                                                                                                |
| ------------ | ---------------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| Enterprise   | 企業管理                           | enterprise                                                                                                                |
| Organization | 組織管理                           | organization                                                                                                              |
| Team         | 團隊與群組管理                     | team                                                                                                                      |
| Membership   | 成員資格管理                       | membership                                                                                                                |
| Repository   | 工作空間、標籤、里程碑、Wiki       | repository.spaces、repository.labels、repository.milestones、repository.wiki                                              |
| Project      | 專案、項目、欄位、檢視、驗收       | project、project.items、project.fields、project.views、project.inspections                                                |
| Issue        | 問題管理                           | issue                                                                                                                     |
| Discussion   | 討論與留言                         | discussion、discussion.comments                                                                                           |
| Journal      | 日誌管理                           | journal                                                                                                                   |
| Attendance   | 打卡、出勤、工時、請假、補正、審核 | attendance.punch、attendance.records、attendance.worktime、attendance.leave、attendance.corrections、attendance.approvals |
| Object       | 物件、範本、文件擷取               | object、object.poppler                                                                                                    |
| Explore      | 跨領域資源搜尋                     | explore                                                                                                                   |
| Summary      | 訊息收集與摘要                     | summary.messagecollection、summary.summarization                                                                          |
| Weather      | 天氣資訊                           | weather.forecast                                                                                                          |

## 3. 業務能力（Business Capability）

能力描述使用者能達成的結果。例如「出勤查詢」讓本人檢視出勤紀錄；「出勤看板」讓獲授權者查看群組統計。兩者同屬 Attendance，但查詢對象與資料範圍不同。同一能力可有多個介面入口。

## 4. 行為主體（Actor）

| 主體           | 任務與範圍                                   |
| -------------- | -------------------------------------------- |
| 成員本人       | 打卡、查閱本人出勤、提出補正                 |
| 協作者         | 在獲授權的工作空間處理專案、問題、討論及日誌 |
| 管理者／審核者 | 在授權範圍管理資源或審核申請                 |
| 系統工作程序   | 代表系統處理訊息收集與摘要工作               |

這些名稱描述參與流程的身分；實際權限依[授權模型](authorization.md)判定。

## 5. 業務資源（Resource）

以出勤為例，Punch Event 記錄一次打卡，Attendance Record 是查詢出勤的對象，Attendance Board 呈現群組彙總，Correction 保存補正申請。Journal 是工作日誌。資源識別、歸屬與操作範圍見[資源模型](resource_model.md)。

## 6. 業務操作（Action）

Clock In 表示上班打卡，Create、Read、Update 分別表示建立、讀取與修改資源，Approve 表示審核通過。業務操作描述使用者意圖；LINE 的 Message、Postback 與 URI 描述觸發方式。

## 7. 跨領域業務流程（Workflow）

| 使用者目標         | 領域如何協作                                                                       | 結果                   |
| ------------------ | ---------------------------------------------------------------------------------- | ---------------------- |
| 搜尋並查看工作資源 | Explore 回傳搜尋結果，使用者開啟選定資源，由所屬領域依授權讀取                     | 查看資源詳情           |
| 查閱群組出勤       | 依 Team 與 Membership 所代表的群組與成員範圍確認存取資格，再由 Attendance 查詢統計 | 取得指定群組的出勤看板 |

此處描述業務參與關係。事件及狀態變更分別見第 10 節的事件與狀態模型。

## 8. 功能與互動介面對應

**下表是建議配置。要找今天可以點的按鈕，請看第 1 節。** 例如，建議配置將定位打卡放在 MINI App；現行群組選單也提供「打卡」，點選後會顯示打卡選項。

私聊與群組透過 Messaging API 顯示訊息。「入口」表示點選後前往另一個介面；「授權摘要」表示依權限顯示部分資訊；「通知」只傳達消息，審核者需到審核頁作決定。「—」表示這份建議未指定該介面的操作。

| 業務能力 | Resource          | Action   | 1：1 | 群組                   | MINI App |
| -------- | ----------------- | -------- | ---- | ---------------------- | -------- |
| 打卡     | Punch Event       | Clock In | 操作 | —                      | 定位打卡 |
| 出勤查詢 | Attendance Record | Read     | 查詢 | 授權摘要               | 詳細紀錄 |
| 出勤看板 | Attendance Board  | Read     | —    | 聊天室統計卡片（Flex） | 統計看板 |
| 建立日誌 | Journal           | Create   | 入口 | 入口                   | 表單     |
| 查詢日誌 | Journal           | Read     | 入口 | 授權摘要               | 列表     |
| 編輯日誌 | Journal           | Update   | —    | —                      | 編輯表單 |
| 補正審核 | Correction        | Approve  | 通知 | —                      | 審核頁面 |

## 9. 功能實作狀態

| 範圍                                 | 可確認的狀態                                     | 查核依據                                           |
| ------------------------------------ | ------------------------------------------------ | -------------------------------------------------- |
| 第 2 節各領域                        | 代表能力已有契約登錄，可依鍵追查 Adapter 與 Core | [能力登錄檔](../../api/contract/capabilities.json) |
| 打卡選項、本人月度出勤、群組出勤看板 | 現行互動盤點已記錄 Quick Reply 與兩個 Flex 案例  | [LINE 互動盤點](line_interaction_model.md)         |
| 第 8 節其餘介面配置                  | 建議設計，逐格實作與端到端驗收待確認             | 對照各能力契約、介面入口與行為測試                 |

## 10. 業務能力與契約索引

依第 2 節的契約鍵查[能力登錄檔](../../api/contract/capabilities.json)。每筆的 `contract`、`adapter`、`core`、`specification` 分別指向 API 契約、轉接層、核心實作與規格。

| 要回答的問題               | 權威來源                                                                                                |
| -------------------------- | ------------------------------------------------------------------------------------------------------- |
| 系統能完成哪些業務目標？   | 本業務能力模型                                                                                          |
| 管理哪些資源？             | [Resource Model](resource_model.md)                                                                     |
| 遵守哪些業務規則？         | [Domain Rules](domain_rules.md)                                                                         |
| 誰可以操作？               | [Authorization](authorization.md)                                                                       |
| 操作如何改變狀態？         | [State Transition](state_transition.md)                                                                 |
| 發生哪些事件？             | [Event Model](event_model.md)                                                                           |
| 資料如何組織？             | [Data Model](data_model.md)                                                                             |
| 使用者如何操作並取得結果？ | [LINE Interaction Model](line_interaction_model.md) 與[互動分配標準](../standards/line_interactions.md) |
