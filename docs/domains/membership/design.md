# Membership 成員資格

Membership 管理成員關係與資格判定。Organization、Enterprise 和 Team 各自擁有資源狀態與管理歸屬。[組裝入口](../../../backend/internal/bootstrap/server.go)將成員資格查詢接到 Team 與 Enterprise，並註冊 Membership 的 HTTP handler。

## 現行讀取與授權

- Team 的空值或 `legacy` 模式沿用既有 LINE 群組資格。`membership` 模式只採獨立關係。未知模式、停用的 Team 或缺少有效資格時拒絕存取。團隊管理歸屬和候選來源見 [Team 規格](../team/design.md#獨立-membership-的讀取邊界)。
- 獨立模式須有有效的 Team `member` 或 `maintainer` 關係。成員關係記錄的上層類型、資源 ID 與世代須符合管理歸屬指定的有效 Organization 或 Enterprise 資格。查詢沿用呼叫端的 Firestore 交易。上層撤權使衍生 Team 資格失效，重新加入後也須重新授予。[資格讀取實作](../../../backend/internal/business/membership/adapter/firestore/qualification.go) · [世代判定](../../../backend/internal/business/membership/domain/qualification.go)
- Organization 的 Bot 私聊邀請及現行成員操作沿用既有 Organization 授權模式。[LINE 邀請入口](../../../backend/internal/business/membership/adapter/line/invitations.go) · [HTTP 成員入口](../../../backend/internal/business/membership/adapter/http/members.go) · [組織邀請操作](../../proposals/membership.md#首版-bot-私聊入口)

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
