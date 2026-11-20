# 團隊管理與指派

## Team（團隊）

| 資料或操作 | 規則與驗收條件 |
|---|---|
| `teams/{teamId}` | 文件 ID 等於 LINE groupId。首次綁定時成對寫入 managementScope（organization 或 enterprise）及 managementId（對應層級的 ID）。每個 Team 只有一個管理歸屬，舊群組可暫未綁定。 |
| 綁定群組 | 指定管理歸屬須具目標管理範圍的管理權及該群組的管理授權。交易內確認原歸屬；同一綁定可重送，不同管理者競爭時只有一方成功。可信 LINE 群組事件沿用既有流程自動建立 Team，不需人工核准；管理歸屬另行設定。 |
| `organizations/{organizationId}/enterpriseTeams/{teamId}` | 記錄一個 Enterprise Team 在此組織的指派。同一 Team 可指派給同企業的多個組織；成員名單由企業集中維護。 |

例如總部「資安組」可以用同一個 teamId 指派至台灣與日本兩個組織，各組織分別授予 Repository 權限。「台灣排班組」由台灣組織管理，只在台灣使用。後端判定資源存取時須核對目前的組織歸屬或有效指派、使用者資格及個別資源授權；既有 LINE 資料仍遵循原功能的成員核驗規則。

既有出勤與摘要繼續以原 groupId 查詢。既有群組在明確綁定前維持未歸屬；組織所有權本身不授予 LINE 群組資料讀取權。


[跨領域歸屬契約](../../architecture/contracts/organization-team-ownership.md) · [領域入口](README.md)
