# Membership 成員資格

Membership 管理成員關係與資格判定。Organization、Enterprise 和 Team 各自擁有資源狀態與管理歸屬。[組裝入口](../../../backend/internal/bootstrap/server.go)將成員資格查詢接到 Team 與 Enterprise，並註冊 Membership 的 HTTP handler。

## 現行讀取與授權

- Team 的空值或 `legacy` 模式沿用既有 LINE 群組資格。`membership` 模式只採獨立關係。未知模式、停用的 Team 或缺少有效資格時拒絕存取。團隊管理歸屬和候選來源見 [Team 規格](../team/design.md#獨立-membership-的讀取邊界)。
- 獨立模式須有有效的 Team `member` 或 `maintainer` 關係。成員關係記錄的上層類型、資源 ID 與世代須符合管理歸屬指定的有效 Organization 或 Enterprise 資格。查詢沿用呼叫端的 Firestore 交易。上層撤權使衍生 Team 資格失效，重新加入後也須重新授予。[資格讀取實作](../../../backend/internal/business/membership/adapter/firestore/qualification.go) · [世代判定](../../../backend/internal/business/membership/domain/qualification.go)
- Organization 的 Bot 私聊邀請及現行成員操作沿用既有 Organization 授權模式。[LINE 邀請入口](../../../backend/internal/business/membership/adapter/line/invitations.go) · [HTTP 成員入口](../../../backend/internal/business/membership/adapter/http/members.go)

### 首版 Bot 私聊入口

1. A 建立組織後，向官方帳號傳送 `邀請加入 <組織 ID>`。Bot 建立邀請後回傳可直接轉傳的指令。
2. A 私下把 `接受邀請 <邀請碼>` 傳給 B。B 開啟自己的官方帳號私聊，送出該指令即可加入。有效期為 24 小時。A 也可傳送 `撤銷邀請 <邀請碼>`。
3. LINE Adapter 使用已驗簽私聊事件的 Actor 與事件 ID。Membership Application 核對目前 Owner、效期和接受資格。Firestore Adapter 在同一交易消耗邀請、寫入稽核並更新現行 `memberUserIds` 權威。Organization Owner 移轉在原交易中遞增 Owner 版本。

邀請碼依渠道、Actor 與事件 ID 經服務端密鑰產生；同一發邀事件可重建同一邀請碼，儲存只保留摘要。接受結果以同一已驗簽事件重送為準。Bot 在交易提交後才回覆。此相容切片僅可在 legacy 模式執行，未建立第二套成員權威，Organization 獨立資格切換仍待交付。

### MINI App 組織頁

組織頁預設開啟成員分頁，另有組織團隊與組織管理分頁。首版成員名單只對當前成員開放，使用組織範圍內的成員代碼供辨識。使用者可分享自己的代碼供 Owner 核對；原始 LINE 使用者 ID 不回傳。一般成員可退出，Owner 可移除他人，現任 Owner 必須先移轉才能退出。

現行 HTTP 契約為 `GET /api/v1/organizations/{organizationId}/members`、`DELETE /api/v1/organizations/{organizationId}/membership`（本人退出）及 `DELETE /api/v1/organizations/{organizationId}/members/{memberId}`（Owner 移除）。兩個異動都要求 `Idempotency-Key`。交易中以同一組織文件重新判斷權限，原子更新 `memberUserIds`、記錄最近退出時間並保存稽核與重送收據。最近退出時間阻擋之前簽發、尚未兌換的邀請重新開權。退出後重載組織選單，移除後重載名單。

來源：[組織成員 HTTP 契約](../../../api/contract/organization/membership.yaml) · [成員交易](../../../backend/internal/business/membership/adapter/firestore/members.go) · [組織頁](../../../frontend/miniapp/src/business/organization/OrganizationPage.tsx)

## 現行資格驗收

| 情境                                                                                                           | 預期結果                                                                                  | 驗收依據                                                                                                                                                                                                                              |
| -------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 已觀察的 Team 由同時具備 LINE 群組資格與有效企業管理資格的使用者開通。完成後該使用者仍具有效 Team 及企業關係。 | 開通前缺少 LINE 資格時拒絕。成功開通後，Team 資格依獨立關係讀取，即使原 LINE 資格已移除。 | [開通與資格讀取測試](../../../backend/internal/business/team/adapter/firestore/activation_test.go)使用實際的 Membership 查詢。                                                                                                        |
| Team 列表存在候選紀錄，但獨立資格查詢回傳拒絕。                                                                | 列表不顯示該 Team，直接讀取詳情也回找不到。                                               | [Team 目錄測試](../../../backend/internal/business/team/adapter/firestore/group_catalog_test.go)以替身控制資格查詢結果，驗證列表與詳情。                                                                                              |
| 既有 Team 關係仍在，先撤銷上層資格，再用新世代恢復上層資格。                                                   | 撤權後拒絕讀取；上層恢復後舊 Team 關係仍無效，直到依新世代重新授予。                      | 依[世代判定](../../../backend/internal/business/membership/domain/qualification.go)與[資格讀取實作](../../../backend/internal/business/membership/adapter/firestore/qualification.go)核對。驗收時在 Firestore emulator 逐步重讀資格。 |

前兩項測試需要 Firestore emulator。此表列的是驗收步驟與預期結果，正式 LINE 與 Firestore 環境仍待獨立驗收。

<a id="切換與驗收"></a>

## 後續切換

首版組織邀請沿用 Bot 私聊入口與既有到期時間；邀請過期時，由有權邀請的管理者重新發出。獨立模式的邀請效期、收據重送窗口與稽核保留期間留待啟用前定案。

獨立資格的全面切換仍須完成各資源的正向讀寫、入口與環境驗收。切換順序、儲存遷移與交付條件由 [Membership 提案](../../proposals/membership.md#儲存與遷移方案)維護。
