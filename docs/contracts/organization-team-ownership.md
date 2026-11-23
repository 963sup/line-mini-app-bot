# Organization 所有權與 Team 歸屬

實作範圍：組織與 Team API、Firestore 寫入、側欄切換及團隊清單。文件路徑均位於 channels/{channelId} 下。現行群組資料與授權行為見[儲存庫規格](../domains/repository/design.md)。

## 企業、組織與團隊歸屬

現行實作

```mermaid
flowchart TB
  U["User"] -->|"ownerUserId"| O["Organization"]
  E["Enterprise"] -->|"旗下組織"| O
  E -->|"管理"| ET["Enterprise Team"]
  O -->|"管理"| OT["Organization Team"]
  ET -->|"可指派至多個旗下組織"| O
  ET --> I["teamId = LINE groupId"]
  OT --> I
  O -->|"個別授權"| R["Repository"]
```

## 管理操作與既有資料

| HTTP 操作                                                                      | 用途                               |
| ------------------------------------------------------------------------------ | ---------------------------------- |
| POST /api/v1/organizations                                                     | 登入者建立獨立組織。               |
| PUT /api/v1/organizations/{organizationId}/owner                               | 現任 Owner 移轉給組織現有成員。    |
| PUT /api/v1/teams/{teamId}/management                                          | 為已自動建立的 Team 指定管理歸屬。 |
| PUT 或 DELETE /api/v1/organizations/{organizationId}/enterprise-teams/{teamId} | 新增或移除同企業內的 Team 指派。   |

組織的 memberUserIds 代表有效成員；管理歸屬操作讀取服務端維護的 managerUserIds。企業自助開通與組織加入流程見[企業規格](../domains/enterprise/design.md)。既有 LINE 群組在下一筆可信事件時以 groupId 補齊 Team；事件只更新群組名稱與狀態，保留既有管理範圍與權限。

## LINE 一對一建立入口

現行實作

```mermaid
flowchart TB
  A["已驗證的 LINE 私訊事件"] --> B{"此事件已處理？"}
  B -->|"是"| C["回傳原結果"]
  B -->|"建立組織"| D["詢問名稱，等待五分鐘"]
  B -->|"建立企業"| E{"有可用開通資格？"}
  E -->|"有"| D
  E -->|"無"| F["告知無資格並結束"]
  D -->|"取消或逾時"| G["結束對話"]
  D -->|"名稱無效"| D
  D -->|"有效組織名稱"| H["交易：建立資源、設定 Owner、結束對話、記錄結果"]
  D -->|"有效企業名稱"| I{"交易內重新檢查資格"}
  I -->|"有"| H
  I -->|"無"| F
  H --> J["回覆名稱與開啟 MINI App 按鈕"]
```

在私訊輸入「建立組織」或「建立企業」，五分鐘內回覆名稱即可建立；輸入「取消」結束。建立企業時，詢問名稱及實際建立時都檢查開通資格。已驗證的 webhook 決定建立者身分。成功回覆附上對應 MINI App 入口。對話狀態依渠道與使用者隔離；資源建立、結束對話及事件回覆紀錄在同一個 Firestore 交易完成，重送同一事件會取得原結果。群組事件沿用原有 Team 流程。

<a id="ownership-boundaries"></a>

## 企業、組織與團隊的授權邊界

```mermaid
flowchart TD
    request["已驗證的操作身分"] --> action{"要改變哪種歸屬？"}
    action -->|企業開通| enterprise["Enterprise 核對開通資格與擁有者"]
    action -->|組織移轉| org["Organization 核對現任擁有者與受讓成員"]
    action -->|團隊掛接| team["Team 核對管理方、團隊端授權與既有歸屬"]
    enterprise --> owned["依各自規則提交或拒絕"]
    org --> owned
    team --> owned
```

[enterprise](../domains/enterprise/design.md) · [organization](../domains/organization/design.md) · [team](../domains/team/design.md)
